package xlogger

import (
	"strings"

	"go.uber.org/fx/fxevent"
)

// fxEventLogger wraps our logger.Logger to implement fxevent.Logger interface
type fxEventLogger struct {
	logger Logger
}

// NewFxEventLogger creates a new FX event logger from the provided logger
func NewFxEventLogger(logger Logger) fxevent.Logger {
	return &fxEventLogger{logger: logger}
}

// LogEvent implements fxevent.Logger interface
func (a *fxEventLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		a.logger.Debug("FX OnStart executing", String("function", e.FunctionName))
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			a.logger.Error("FX OnStart failed",
				String("function", e.FunctionName),
				Error(e.Err))
		} else {
			a.logger.Debug("FX OnStart completed", String("function", e.FunctionName))
		}
	case *fxevent.OnStopExecuting:
		a.logger.Debug("FX OnStop executing", String("function", e.FunctionName))
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			a.logger.Error("FX OnStop failed",
				String("function", e.FunctionName),
				Error(e.Err))
		} else {
			a.logger.Debug("FX OnStop completed", String("function", e.FunctionName))
		}
	case *fxevent.Supplied:
		a.logger.Debug("FX supplied", String("type", e.TypeName))
	case *fxevent.Provided:
		a.logger.Debug("FX provided", String("constructor", e.ConstructorName))
	case *fxevent.Invoking:
		a.logger.Debug("FX invoking", String("function", e.FunctionName))
	case *fxevent.Invoked:
		if e.Err != nil {
			a.logger.Error("FX invoke failed",
				String("function", e.FunctionName),
				Error(e.Err))
		} else {
			a.logger.Debug("FX invoke completed", String("function", e.FunctionName))
		}
	case *fxevent.Stopping:
		a.logger.Info("FX stopping", String("signal", strings.ToUpper(e.Signal.String())))
	case *fxevent.Stopped:
		if e.Err != nil {
			a.logger.Error("FX stop failed", Error(e.Err))
		} else {
			a.logger.Info("FX stopped successfully")
		}
	case *fxevent.RollingBack:
		a.logger.Error("FX rolling back due to start failure", Error(e.StartErr))
	case *fxevent.RolledBack:
		if e.Err != nil {
			a.logger.Error("FX rollback failed", Error(e.Err))
		} else {
			a.logger.Info("FX rolled back successfully")
		}
	case *fxevent.Started:
		if e.Err != nil {
			a.logger.Error("FX start failed", Error(e.Err))
		} else {
			a.logger.Info("FX started successfully")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			a.logger.Error("FX logger initialization failed", Error(e.Err))
		} else {
			a.logger.Debug("FX logger initialized", String("constructor", e.ConstructorName))
		}
	case *fxevent.BeforeRun:
		a.logger.Debug("FX before run", String("kind", e.Kind))
	case *fxevent.Run:
		a.logger.Debug("FX run",
			String("kind", e.Kind),
			String("name", e.Name))
	case *fxevent.Decorated:
		a.logger.Debug("FX decorated", String("decorator", e.DecoratorName))
	default:
		// Discard unknown events to reduce noise
	}
}
