package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/providers"
	"github.com/adamkobi/xt/pkg/ssh"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	execCmdViper = viper.New()
)

func init() {
	config.InitViper(execCmdViper)
	RootCmd.AddCommand(execCmd)
	execCmd.PersistentFlags().BoolP("force", "f", false, "run command without approval, good for automation")
	execCmd.PersistentFlags().BoolP("tty", "x", false, "request tty")
	execCmd.PersistentFlags().BoolP("all", "a", false, "run flow on all servers")
	execCmdViper.BindPFlags(execCmd.PersistentFlags())
}

var execCmd = &cobra.Command{
	Use:   "exec SERVERS [COMMAND]",
	Short: "Execute remote commands",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		searchPattern := args[0]
		remoteCmd := args[1:]

		instanceIDs, err := providers.GetIds(searchPattern)
		if err != nil {
			logger.Fatalf("fetching instance ids failed, %v", err)
		}

		if execCmdViper.GetBool("all") {
			if !checkUserApproval(instanceIDs, remoteCmd) {
				logger.Fatalf("cancelled...")
			}
			executers := ssh.CreateSSHExecuters(instanceIDs, remoteCmd)
			ssh.RunMany(executers)
			os.Exit(0)
		}

		instanceID := utils.SelectInstance(instanceIDs, searchPattern)
		if !checkUserApproval([]string{instanceID}, remoteCmd) {
			logger.Fatal("cancelled...")
		}

		e := ssh.NewSSHExecuter(instanceID, remoteCmd)
		if execCmdViper.GetBool("tty") {
			e.CommandWithTTY()
			os.Exit(0)
		}
		output := e.CommandWithOutput()
		ctx := logger.WithFields(log.Fields{
			"host":    e.Hostname,
			"command": e.RemoteCmd,
		})
		if output.Error {
			ctx.Errorf("command failed")
			fmt.Fprint(os.Stderr, output.Data)
		} else {
			ctx.Info("command output")
			fmt.Fprint(os.Stdout, output.Data)
		}
	},
}

func checkUserApproval(instanceIDs, remoteCmd []string) bool {
	if !execCmdViper.GetBool("force") && !utils.GetApproval(strings.Join(remoteCmd, " "), instanceIDs) {
		return false
	}
	return true
}
