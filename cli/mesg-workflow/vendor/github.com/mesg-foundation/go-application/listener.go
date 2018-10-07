package mesg

import (
	"context"
	"sync"
)

// Listener is a task execution listener.
type Listener struct {
	// Err filled when Listener fails to continue.
	Err chan error

	executions chan Execution

	// cancel stops receiving from gRPC stream.
	cancel context.CancelFunc

	app *Application

	// gracefulWait will be in the done state when all processing
	// events or results are done.
	gracefulWait *sync.WaitGroup
}

// ListenerOption is the condition configurator for listener.
type ListenerOption func(*Listener)

func ExecutionOption(executions chan Execution) ListenerOption {
	return func(l *Listener) {
		l.executions = executions
	}
}

// Execution is a task execution.
type Execution struct {
	// ID is execution id of task.
	ExecutionID string

	// Err filled if an error occurs during task execution.
	Err error
}

func newListener(app *Application, gracefulWait *sync.WaitGroup, options ...ListenerOption) *Listener {
	l := &Listener{
		app:          app,
		gracefulWait: gracefulWait,
		Err:          make(chan error, 1),
	}
	for _, option := range options {
		option(l)
	}
	return l
}

func (l *Listener) sendError(err error) {
	l.Err <- err
}

func (l *Listener) sendExecution(e Execution) {
	if l.executions != nil {
		l.executions <- e
	}
}

// Close gracefully waits current events or results to complete their process and
// stops listening for future events or results.
func (l *Listener) Close() error {
	l.app.removeListener(l)
	l.cancel()
	l.gracefulWait.Wait()
	if l.executions != nil {
		close(l.executions)
	}
	return nil
}
