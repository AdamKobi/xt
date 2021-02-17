package flow

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/adamkobi/xt/pkg/cmdutil"
	addCmd "github.com/adamkobi/xt/pkg/command/flow/add"
	deleteCmd "github.com/adamkobi/xt/pkg/command/flow/delete"
	listCmd "github.com/adamkobi/xt/pkg/command/flow/list"
	runCmd "github.com/adamkobi/xt/pkg/command/flow/run"
	"github.com/spf13/cobra"
)

func NewCmdFlow(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flow <command>",
		Short: "Run flows remotley",
		Long: heredoc.Doc(`
			Create complex flows to remote servers using a series of commands from config file.

			Define new flows that will run and return output that can be selected and manipulated in next commands.
		`),
		Example: heredoc.Doc(`
			# connect to management server, query pods by label and return output for selection to next command
			  connect-pods:
			    - run: kubectl get pods -o json -l app=webserver
				  selector: metadata.name
				  root: items
				  output_format: json
			    - run: kubectl exec -it {{.metadata_name}} bash
				  output_format: ""
			# connect to management server, query pods by label and print selected keys as a formated table
			  print-pods:
				- run: kubectl get pods -o json -l app=webserver
				  selector: metadata.name
				  root: items
				  keys:
					- spec.nodeName
					- metadata.labels.role
				  output: json
				  print: true
		`),
	}

	cmd.AddCommand(listCmd.NewCmdList(f))
	cmd.AddCommand(runCmd.NewCmdRun(f))
	cmd.AddCommand(addCmd.NewCmdAdd(f))
	cmd.AddCommand(deleteCmd.NewCmdDelete(f))
	return cmd
}
