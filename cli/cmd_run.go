package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/mesg-foundation/core/x/xsignal"
	"github.com/mesg-foundation/workflow-application/workflow"

	"github.com/spf13/cobra"
)

var (
	colorGreen  = color.New(color.FgGreen)
	colorYellow = color.New(color.FgYellow)
	colorBold   = color.New(color.Bold)

	colorAttention = color.New(color.FgYellow, color.Bold)
	colorInfo      = color.New(color.FgBlue)
	colorError     = color.New(color.FgRed, color.Bold)
)

type runCmd struct {
	cmd *cobra.Command
}

func newRunCmd() *runCmd {
	rc := &runCmd{
		cmd: &cobra.Command{
			Use:   "run",
			Short: "Run your workflow",
		},
	}
	rc.cmd.RunE = rc.run
	return rc
}

func (c *runCmd) run(cmd *cobra.Command, args []string) error {
	// check workflow file argument.
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return errors.New("workflow yml file should be provided")
	}

	// spinner for initialization.
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()
	defer s.Stop()

	// check if workflow file is exists.
	path := args[0]
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	var (
		events     = make(chan workflow.Event)
		executions = make(chan workflow.Execution)
	)

	// run the workflow.
	w, err := workflow.NewFromYAML(f,
		workflow.EventsOption(events),
		workflow.ExecutionsOption(executions),
	)
	if err != nil {
		return err
	}
	if err := w.Run(); err != nil {
		return err
	}

	// application started.
	s.Stop()
	fmt.Println(colorGreen.Sprintf("✔ %s workflow started", colorBold.Sprintf("%s", w.Name)))
	fmt.Println(colorYellow.Sprintf("%s", w.Description))

	// listen for workflow events and an interrupt.
	for {
		select {
		case event := <-events:
			c.onEvent(event)

		case execution := <-executions:
			c.onExecution(execution)

		case err := <-w.Err:
			return err

		case <-xsignal.WaitForInterrupt():
			return w.Destroy()
		}
	}
}

func (c *runCmd) onEvent(info workflow.Event) {
	data, err := json.MarshalIndent(info.ExecutionData, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Printf(">> event %s received on %s service, execution data will be: %s\n",
		colorAttention.Sprintf("%s", info.EventKey),
		colorAttention.Sprintf("%s", info.ServiceName),
		colorInfo.Sprintf(" %+v", string(data)))
}

func (c *runCmd) onExecution(info workflow.Execution) {
	if info.Error == nil {
		fmt.Printf("<< execution successfully made for %s task on %s service\n",
			colorAttention.Sprintf("%s", info.TaskKey),
			colorAttention.Sprintf("%s", info.ServiceName))
		return
	}

	fmt.Printf("<< execution completed with an error for %s task on %s service: %s\n",
		colorAttention.Sprintf("%s", info.TaskKey),
		colorAttention.Sprintf("%s", info.ServiceName),
		colorError.Sprintf("%q", info.Error),
	)
}

func (c *runCmd) getCmd() *cobra.Command {
	return c.cmd
}
