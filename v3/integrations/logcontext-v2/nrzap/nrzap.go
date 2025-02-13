package nrzap

import (
	"errors"

	"github.com/TykTechnologies/newrelic-go-agent/v3/internal"
	"github.com/TykTechnologies/newrelic-go-agent/v3/newrelic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() { internal.TrackUsage("integration", "logcontext-v2", "zap") }

// NewRelicZapCore implements zap.Core
type NewRelicZapCore struct {
	core zapcore.Core
	nr   newrelicApplicationState
}

// newrelicApplicationState is a private struct that stores newrelic application data
// for automatic behind the scenes log collection logic.
type newrelicApplicationState struct {
	app *newrelic.Application
	txn *newrelic.Transaction
}

// internal handler function to manage writing a log to the new relic application
func (nr *newrelicApplicationState) recordLog(entry zapcore.Entry, fields []zap.Field) {
	data := newrelic.LogData{
		Timestamp: entry.Time.UnixMilli(),
		Severity:  entry.Level.String(),
		Message:   entry.Message,
	}

	if nr.txn != nil {
		nr.txn.RecordLog(data)
	} else if nr.app != nil {
		nr.app.RecordLog(data)
	}
}

var (
	// ErrNilZapcore is an error caused by calling a WrapXCore function on a nil zapcore.Core object
	ErrNilZapcore = errors.New("cannot wrap nil zapcore.Core object")
	// ErrNilApp is an error caused by calling WrapBackgroundCore with a nil newrelic.Application
	ErrNilApp = errors.New("wrapped a zapcore.Core with a nil New Relic application; logs will not be captured")
	// ErrNilTxn is an error caused by calling WrapTransactionCore with a nil newrelic.Transaction
	ErrNilTxn = errors.New("wrapped a zapcore.Core with a nil New Relic transaction; logs will not be captured")
)

// NewBackgroundCore creates a new NewRelicZapCore object, which is a wrapped zapcore.Core object. This wrapped object
// captures background logs in context and sends them to New Relic.
//
// Errors will be returned if the zapcore object is nil, or if the application is nil. It is up to the user to decide
// how to handle the case where the newrelic.Application is nil.
// In the case that the newrelic.Application is nil, a valid NewRelicZapCore object will still be returned.
func WrapBackgroundCore(core zapcore.Core, app *newrelic.Application) (*NewRelicZapCore, error) {
	if core == nil {
		return nil, ErrNilZapcore
	}

	var err error
	if app == nil {
		err = ErrNilApp
	}

	return &NewRelicZapCore{
		core: core,
		nr: newrelicApplicationState{
			app: app,
		},
	}, err
}

// WrapTransactionCore creates a new NewRelicZapCore object, which is a wrapped zapcore.Core object. This wrapped object
// captures logs in context of a transaction and sends them to New Relic.
//
// Errors will be returned if the zapcore object is nil, or if the application is nil. It is up to the user to decide
// how to handle the case where the newrelic.Transaction is nil.
// In the case that the newrelic.Application is nil, a valid NewRelicZapCore object will still be returned.
func WrapTransactionCore(core zapcore.Core, txn *newrelic.Transaction) (*NewRelicZapCore, error) {
	if core == nil {
		return nil, ErrNilZapcore
	}

	var err error
	if txn == nil {
		err = ErrNilTxn
	}
	return &NewRelicZapCore{
		core: core,
		nr: newrelicApplicationState{
			txn: txn,
		},
	}, err
}

// With makes a copy of a NewRelicZapCore with new zap.Fields. It calls zapcore.With() on the zap core object
// then makes a deepcopy of the NewRelicApplicationState object so the original
// object can be deallocated when it's no longer in scope.
func (c NewRelicZapCore) With(fields []zap.Field) zapcore.Core {
	return NewRelicZapCore{
		core: c.core.With(fields),
		nr: newrelicApplicationState{
			c.nr.app,
			c.nr.txn,
		},
	}
}

// Check simply calls zapcore.Check on the Core object.
func (c NewRelicZapCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	ce := c.core.Check(entry, checkedEntry)
	ce.AddCore(entry, c)
	return ce
}

// Write wraps zapcore.Write and captures the log entry and sends that data to New Relic.
func (c NewRelicZapCore) Write(entry zapcore.Entry, fields []zap.Field) error {
	c.nr.recordLog(entry, fields)
	return nil
}

// Sync simply calls zapcore.Sync on the Core object.
func (c NewRelicZapCore) Sync() error {
	return c.core.Sync()
}

// Enabled simply calls zapcore.Enabled on the zapcore.Level passed to it.
func (c NewRelicZapCore) Enabled(level zapcore.Level) bool {
	return c.core.Enabled(level)
}
