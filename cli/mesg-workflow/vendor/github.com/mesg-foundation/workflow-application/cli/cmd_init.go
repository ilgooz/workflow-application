package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

type initCmd struct {
	cmd *cobra.Command
}

func newInitCmd() *initCmd {
	rc := &initCmd{
		cmd: &cobra.Command{
			Use:   "init",
			Short: "Initialize an empty workflow",
		},
	}
	rc.cmd.Run = rc.run
	return rc
}

func (c *initCmd) run(cmd *cobra.Command, args []string) {
	fmt.Println("init")
}

func (c *initCmd) getCmd() *cobra.Command {
	return c.cmd
}
