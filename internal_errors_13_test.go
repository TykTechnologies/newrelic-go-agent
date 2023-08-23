// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// +build go1.13

package newrelic

import (
	"fmt"
	"testing"

	"github.com/TykTechnologies/newrelic-go-agent/internal"
)

func TestNoticedWrappedError(t *testing.T) {
	gamma := func() error {
		return Error{
			Message: "socket error",
			Class:   "socketError",
			Attributes: map[string]interface{}{
				"zip": "zap",
			},
		}
	}
	beta := func() error { return fmt.Errorf("problem in beta: %w", gamma()) }
	alpha := func() error { return fmt.Errorf("problem in alpha: %w", beta()) }

	app := testApp(nil, nil, t)
	txn := app.StartTransaction("hello", nil, nil)
	err := txn.NoticeError(alpha())
	if nil != err {
		t.Error(err)
	}
	txn.End()
	app.ExpectErrors(t, []internal.WantError{{
		TxnName: "OtherTransaction/Go/hello",
		Msg:     "problem in alpha: problem in beta: socket error",
		Klass:   "socketError",
		UserAttributes: map[string]interface{}{
			"zip": "zap",
		},
	}})
	app.ExpectErrorEvents(t, []internal.WantEvent{{
		Intrinsics: map[string]interface{}{
			"error.class":     "socketError",
			"error.message":   "problem in alpha: problem in beta: socket error",
			"transactionName": "OtherTransaction/Go/hello",
		},
		UserAttributes: map[string]interface{}{
			"zip": "zap",
		},
	}})
	app.ExpectMetrics(t, backgroundErrorMetrics)
}
