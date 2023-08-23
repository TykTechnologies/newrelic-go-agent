// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package newrelic

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/TykTechnologies/newrelic-go-agent/v3/internal/jsonx"
)

const (
	// panicErrorKlass is the error klass used for errors generated by
	// recovering panics in txn.End.
	panicErrorKlass = "panic"
)

func panicValueMsg(v interface{}) string {
	switch val := v.(type) {
	case error:
		return val.Error()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// txnErrorFromPanic creates a new TxnError from a panic.
func txnErrorFromPanic(now time.Time, v interface{}) errorData {
	return errorData{
		When:  now,
		Msg:   panicValueMsg(v),
		Klass: panicErrorKlass,
	}
}

// txnErrorFromResponseCode creates a new TxnError from an http response code.
func txnErrorFromResponseCode(now time.Time, code int) errorData {
	codeStr := strconv.Itoa(code)
	msg := http.StatusText(code)
	if msg == "" {
		// Use a generic message if the code was not an http code
		// to support gRPC.
		msg = "response code " + codeStr
	}
	return errorData{
		When:  now,
		Msg:   msg,
		Klass: codeStr,
	}
}

// errorData contains the information about a recorded error.
type errorData struct {
	When            time.Time
	Stack           stackTrace
	RawError        error
	ExtraAttributes map[string]interface{}
	ErrorGroup      string
	Msg             string
	Klass           string
	SpanID          string
	Expect          bool
}

// txnError combines error data with information about a transaction.  txnError is used for
// both error events and traced errors.
type txnError struct {
	errorData
	txnEvent
}

// errorEvent and tracedError are separate types so that error events and traced errors can have
// different WriteJSON methods.
type errorEvent txnError

type tracedError txnError

// txnErrors is a set of errors captured in a Transaction.
type txnErrors []*errorData

// NewTxnErrors returns a new empty txnErrors.
func newTxnErrors(max int) txnErrors {
	return make([]*errorData, 0, max)
}

// Add adds a TxnError.
func (errors *txnErrors) Add(e errorData) {
	if len(*errors) < cap(*errors) {
		*errors = append(*errors, &e)
	}
}

func (h *tracedError) WriteJSON(buf *bytes.Buffer) {
	buf.WriteByte('[')
	jsonx.AppendFloat(buf, timeToFloatMilliseconds(h.When))
	buf.WriteByte(',')
	jsonx.AppendString(buf, h.FinalName)
	buf.WriteByte(',')
	jsonx.AppendString(buf, h.Msg)
	buf.WriteByte(',')
	jsonx.AppendString(buf, h.Klass)
	buf.WriteByte(',')

	buf.WriteByte('{')
	buf.WriteString(`"agentAttributes"`)
	buf.WriteByte(':')
	agentAttributesJSON(h.Attrs, buf, destError)
	buf.WriteByte(',')
	buf.WriteString(`"userAttributes"`)
	buf.WriteByte(':')
	userAttributesJSON(h.Attrs, buf, destError, h.errorData.ExtraAttributes)
	buf.WriteByte(',')
	buf.WriteString(`"intrinsics"`)
	buf.WriteByte(':')
	intrinsicsJSON(&h.txnEvent, buf, h.errorData.Expect)
	if nil != h.Stack {
		buf.WriteByte(',')
		buf.WriteString(`"stack_trace"`)
		buf.WriteByte(':')
		h.Stack.WriteJSON(buf)
	}
	buf.WriteByte('}')

	buf.WriteByte(']')
}

// MarshalJSON is used for testing.
func (h *tracedError) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	h.WriteJSON(buf)
	return buf.Bytes(), nil
}

type harvestErrors []*tracedError

func newHarvestErrors(max int) harvestErrors {
	return make([]*tracedError, 0, max)
}

// mergeTxnErrors merges a transaction's errors into the harvest's errors.
func mergeTxnErrors(errors *harvestErrors, errs txnErrors, txnEvent txnEvent, hs *highSecuritySettings) {
	for _, e := range errs {
		if len(*errors) == cap(*errors) {
			return
		}

		e.scrubErrorForHighSecurity(hs)
		*errors = append(*errors, &tracedError{
			txnEvent:  txnEvent,
			errorData: *e,
		})
	}
}

func (errors harvestErrors) Data(agentRunID string, harvestStart time.Time) ([]byte, error) {
	if len(errors) == 0 {
		return nil, nil
	}
	estimate := 1024 * len(errors)
	buf := bytes.NewBuffer(make([]byte, 0, estimate))
	buf.WriteByte('[')
	jsonx.AppendString(buf, agentRunID)
	buf.WriteByte(',')
	buf.WriteByte('[')
	for i, e := range errors {
		if i > 0 {
			buf.WriteByte(',')
		}
		e.WriteJSON(buf)
	}
	buf.WriteByte(']')
	buf.WriteByte(']')
	return buf.Bytes(), nil
}

func (errors harvestErrors) MergeIntoHarvest(h *harvest) {}

func (errors harvestErrors) EndpointMethod() string {
	return cmdErrorData
}

// applyErrorGroup applies the error group callback function to an errorData object. It will either consume the txn object
// or the txnEvent in that order. If both are nil, nothing will happen.
func (errData *errorData) applyErrorGroup(txnEvent *txnEvent) {
	if txnEvent == nil || txnEvent.errGroupCallback == nil {
		return
	}

	errorInfo := ErrorInfo{
		txnAttributes:   txnEvent.Attrs,
		TransactionName: txnEvent.FinalName,
		errAttributes:   errData.ExtraAttributes,
		stackTrace:      errData.Stack,
		Error:           errData.RawError,
		TimeOccured:     errData.When,
		Message:         errData.Msg,
		Class:           errData.Klass,
		Expected:        errData.Expect,
	}

	// If a user defined an error group callback function, execute it to generate the error group string.
	errGroup := txnEvent.errGroupCallback(errorInfo)

	if errGroup != "" {
		errData.ErrorGroup = errGroup
	}
}

type highSecuritySettings struct {
	enabled                   bool
	allowRawExceptionMessages bool
}

func (errData *errorData) scrubErrorForHighSecurity(hs *highSecuritySettings) {
	if hs == nil {
		return
	}

	//txn.Config.HighSecurity
	if hs.enabled {
		errData.Msg = highSecurityErrorMsg
	}

	//!txn.Reply.SecurityPolicies.AllowRawExceptionMessages.Enabled()
	if !hs.allowRawExceptionMessages {
		errData.Msg = securityPolicyErrorMsg
	}
}

func scrubbedErrorMessage(msg string, txn *txn) string {
	if txn == nil {
		return msg
	}

	if txn.Config.HighSecurity {
		return highSecurityErrorMsg
	}

	if !txn.Reply.SecurityPolicies.AllowRawExceptionMessages.Enabled() {
		return securityPolicyErrorMsg
	}

	return msg
}
