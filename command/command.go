package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/logging"
	"github.com/adamkobi/xt/pkg/providers"
	"github.com/adamkobi/xt/pkg/ssh"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var execCmdViper = viper.New()

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
		tag := viper.GetString("tag")
		searchPattern := args[0] + "*"
		remoteCmd := args[1:]

		instanceIds, err := providers.GetIds(tag, searchPattern)
		if err != nil {
			logging.Main.Fatalf("fetching instance ids failed, %v", err)
		}
		if execCmdViper.GetBool("all") {
			if !execCmdViper.GetBool("force") && !utils.GetApproval(strings.Join(remoteCmd, " "), instanceIds) {
				logging.Main.Error("cancelled...")
				os.Exit(0)
			}
			executers := ssh.CreateSSHExecuters(instanceIds, "", remoteCmd)
			ssh.RunMultiple(executers)
		} else {
			instanceID := utils.SelectInstance(instanceIds, searchPattern)
			if !execCmdViper.GetBool("force") && !utils.GetApproval(strings.Join(remoteCmd, " "), []string{instanceID}) {
				logging.Main.Error("cancelled...")
				os.Exit(0)
			}
			logging.Main.Infof("Connecting to %s...", instanceID)
			executer := ssh.NewSSHExecuter(instanceID, remoteCmd)
			if execCmdViper.GetBool("tty") {
				ssh.CommandWithTTY(executer)
			} else {
				output := ssh.CommandWithOutput(executer)
				if output.Stderr != "" {
					logging.Main.WithFields(log.Fields{
						"host":    executer.Host,
						"command": executer.RemoteCmd,
					}).Fatal("command failed")
				}
				fmt.Println(output.Stdout)
			}
		}
	},
}
