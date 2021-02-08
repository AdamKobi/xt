package flow

import (
	"github.com/adamkobi/xt/command/factory"
	addCmd "github.com/adamkobi/xt/command/flow/add"
	deleteCmd "github.com/adamkobi/xt/command/flow/delete"
	listCmd "github.com/adamkobi/xt/command/flow/list"
	runCmd "github.com/adamkobi/xt/command/flow/run"
	"github.com/spf13/cobra"
)

func NewCmdFlow(f *factory.CmdConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "flow <command>",
		Short:   "Run flows remotley",
		Aliases: []string{"fl", "flo"},
	}

	cmd.AddCommand(listCmd.NewCmdList(f))
	cmd.AddCommand(runCmd.NewCmdRun(f))
	cmd.AddCommand(addCmd.NewCmdAdd(f))
	cmd.AddCommand(deleteCmd.NewCmdDelete(f))
	return cmd
}
