package cli

import (
	"github.com/spf13/cobra"
)

type command interface {
	getCmd() *cobra.Command
}

func Run() {
	rootCmd := newRootCmd()
	rootCmd.addCommands(
		newInitCmd().getCmd(),
		newRunCmd().getCmd(),
	)
	rootCmd.getCmd().Execute()
}
