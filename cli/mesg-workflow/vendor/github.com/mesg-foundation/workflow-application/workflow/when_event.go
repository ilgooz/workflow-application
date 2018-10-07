package workflow

import (
	mesg "github.com/mesg-foundation/go-application"
)

// whenEvent listens for events and executes task.
func (w *Workflow) whenEvent(event EventDefinition) (*mesg.Listener, error) {
	e := w.app.WhenEvent(
		w.getServiceID(event.ServiceName),
		mesg.EventKeyCondition(event.EventKey),
	)

	e.Map(func(ev *mesg.Event) mesg.Data {
		// parse event data.
		eventData := make(map[string]interface{})
		if err := ev.Data(&eventData); err != nil {
			w.log.Println(err)
			return nil
		}

		// return task outputs as described.
		mappings := event.Map
		if len(mappings) > 0 {
			parser := parser{
				configs: w.def.Configs,
				data:    eventData,
			}

			data := make(map[string]interface{})
			for _, mapping := range mappings {
				value, err := parser.Parse(mapping.Value)
				if err != nil {
					w.log.Println(err)
					return nil
				}
				data[mapping.Key] = value
			}

			w.events <- Event{
				ServiceName:   event.ServiceName,
				EventKey:      ev.Key,
				ExecutionData: data,
			}
			return data
		}

		w.events <- Event{
			ServiceName:   event.ServiceName,
			EventKey:      ev.Key,
			ExecutionData: eventData,
		}

		// use event data as task output.
		return eventData
	})

	executions := make(chan mesg.Execution)

	// configure task execution.
	ln, err := e.Execute(
		w.getServiceID(event.Execute.ServiceName),
		event.Execute.TaskKey,
		mesg.ExecutionOption(executions))
	if err != nil {
		return nil, err
	}

	go func() {
		for exec := range executions {
			w.executions <- Execution{
				ServiceName: event.Execute.ServiceName,
				TaskKey:     event.Execute.TaskKey,
				Error:       exec.Err,
			}
		}
	}()

	return ln, nil
}
