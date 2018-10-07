package workflow

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	mesg "github.com/mesg-foundation/go-application"
)

// Workflow is an application workflow for connecting MESG services.
type Workflow struct {
	Name        string
	Description string

	executions chan Execution
	events     chan Event
	Err        chan error

	app *mesg.Application
	def WorkflowDefinition

	log       *log.Logger
	logOutput io.Writer
}

// New creates a new workflow with given workflow definition and options.
func New(def WorkflowDefinition, options ...Option) *Workflow {
	w := &Workflow{
		def:         def,
		logOutput:   os.Stdout,
		executions:  make(chan Execution),
		events:      make(chan Event),
		Err:         make(chan error, 1),
		Name:        def.Name,
		Description: def.Description,
	}
	for _, option := range options {
		option(w)
	}
	w.log = log.New(w.logOutput, "workflow", log.LstdFlags)
	return w
}

// NewFromYAML creates a new workflow with given yaml reader r and options.
func NewFromYAML(r io.Reader, options ...Option) (*Workflow, error) {
	def, err := ParseYAML(r)
	if err != nil {
		return nil, err
	}
	return New(def, options...), nil
}

// Option is the configuration for creating workflows.
type Option func(*Workflow)

// ExecutionsOption sends each execution to executions chan.
func ExecutionsOption(executions chan Execution) Option {
	return func(w *Workflow) {
		w.executions = executions
	}
}

// EventsOption sends each event to events chan.
func EventsOption(events chan Event) Option {
	return func(w *Workflow) {
		w.events = events
	}
}

// mesgOption returns an option for setting mesg application app.
func mesgOption(app *mesg.Application) Option {
	return func(w *Workflow) {
		w.app = app
	}
}

// Run runs the workflow.
func (w *Workflow) Run() error {
	if w.app == nil {
		var err error
		if w.app, err = mesg.New(
			mesg.LogOutputOption(ioutil.Discard),
		); err != nil {
			return err
		}
	}

	var listeners []*mesg.Listener

	for _, event := range w.def.Events {
		ln, err := w.whenEvent(event)
		if err != nil {
			return err
		}
		listeners = append(listeners, ln)
	}

	go w.listenErrors(listeners)
	return nil
}

func (w *Workflow) listenErrors(listeners []*mesg.Listener) {
	errC := make(chan error, len(listeners))
	for _, ln := range listeners {
		go func(ln *mesg.Listener) {
			errC <- <-ln.Err
		}(ln)
	}
	w.Err <- <-errC
	w.Destroy()
}

// Destroy gracefully destroyes workflow.
func (w *Workflow) Destroy() error {
	return w.app.Close()
}

// Event represent and incoming event.
type Event struct {
	ServiceName   string
	EventKey      string
	ExecutionData interface{}
}

// Execution represent an execution result.
type Execution struct {
	ServiceName string
	TaskKey     string
	Error       error
}

func (w *Workflow) getServiceID(name string) string {
	for _, s := range w.def.Services {
		if s.Name == name {
			return s.ID
		}
	}
	return ""
}
