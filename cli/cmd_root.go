package cli

import (
	"github.com/spf13/cobra"
)

type rootCmd struct {
	cmd *cobra.Command
}

func newRootCmd() *rootCmd {
	rc := &rootCmd{
		cmd: &cobra.Command{
			Use:   "mwf",
			Short: "mwf helps to create and run mesg workflows",
		},
	}
	rc.cmd.Run = rc.run
	return rc
}

func (c *rootCmd) run(cmd *cobra.Command, args []string) {
	newRunCmd().getCmd().Execute()
}

func (c *rootCmd) getCmd() *cobra.Command {
	return c.cmd
}

func (c *rootCmd) addCommands(cmds ...*cobra.Command) {
	c.cmd.AddCommand(cmds...)
}
