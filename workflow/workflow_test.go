package workflow

import (
	"io/ioutil"
	"testing"

	mesg "github.com/mesg-foundation/go-application"
	"github.com/mesg-foundation/go-application/mesgtest"
	"github.com/stretchr/testify/require"
	"github.com/stvp/assert"
)

func newAppAndServer(t *testing.T) (*mesg.Application, *mesgtest.Server) {
	testServer := mesgtest.NewServer()
	app, err := mesg.New(
		mesg.DialOption(testServer.Socket()),
		mesg.LogOutputOption(ioutil.Discard),
	)
	assert.Nil(t, err)
	assert.NotNil(t, app)
	return app, testServer
}

func TestNew(t *testing.T) {
	events := make(chan Event)
	executions := make(chan Execution)
	def := WorkflowDefinition{
		Name:        "discord-invites",
		Description: "send discord invites to your fellows",
		Services: []ServiceDefinition{
			{Name: "webhook", ID: "4f7891f77a6333787075e95b6d3d73ad50b5d1e9"},
			{Name: "discord", ID: "1daf16ca98322024824f307a9e11c88e0aba55e2"},
		},
		Configs: []ConfigDefinition{
			{
				Key:   "sendgridAPIKey",
				Value: "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",
			},
		},
		Events: []EventDefinition{
			{
				ServiceName: "webhook",
				EventKey:    "request",
				Map: []MapDefinition{
					{
						Key:   "email",
						Value: "$data.data.email",
					},
					{
						Key:   "sendgridAPIKey",
						Value: "$configs.sendgridAPIKey",
					},
				},
				Execute: ExecuteDefinition{ServiceName: "discord", TaskKey: "send"},
			},
		},
	}

	app, server := newAppAndServer(t)
	go server.Start()

	w := New(def,
		mesgOption(app),
		EventsOption(events),
		ExecutionsOption(executions),
	)

	require.NoError(t, w.Run())

	eventData := webhookRequestEvent{Data: webhookRequestEventBody{Email: "1"}}
	require.NoError(t, server.EmitEvent(def.Services[0].ID, def.Events[0].EventKey, eventData))

	event := <-events
	require.Equal(t, def.Events[0].EventKey, event.EventKey)

	execution := <-executions
	require.Equal(t, def.Events[0].Execute.TaskKey, execution.TaskKey)

	le := <-server.LastExecute()
	assert.Equal(t, def.Services[1].ID, le.ServiceID())
	assert.Equal(t, def.Events[0].Execute.TaskKey, le.Task())

	var taskData discordSendTaskInputs
	assert.Nil(t, le.Data(&taskData))
	assert.Equal(t, def.Configs[0].Value, taskData.SendgridAPIKey)
	assert.Equal(t, eventData.Data.Email, taskData.Email)
}

type webhookRequestEvent struct {
	Data webhookRequestEventBody `json:"data"`
}

type webhookRequestEventBody struct {
	Email string `json:"email"`
}

type discordSendTaskInputs struct {
	Email          string `json:"email"`
	SendgridAPIKey string `json:"sendgridAPIKey"`
}
