package command

import (
	"fmt"
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

var flowCmdViper = viper.New()

func init() {
	config.InitViper(flowCmdViper)
	RootCmd.AddCommand(flowCmd)
	flowCmd.PersistentFlags().BoolP("all", "a", false, "run flow on all servers")
	flowCmdViper.BindPFlags(flowCmd.PersistentFlags())
}

var flowCmd = &cobra.Command{
	Use:   "flow SERVERS FLOW",
	Short: "Execute remote commands from config file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		tag := viper.GetString("tag")
		searchPattern := args[0] + "*"
		flowID := args[1]
		instanceIds, err := providers.GetIds(tag, searchPattern)
		if err != nil {
			logging.Main.WithFields(log.Fields{
				"search pattern": searchPattern,
				"search tag":     tag,
			}).Fatalf("fetching ids failed, %v", err)
		}

		if !flowCmdViper.IsSet("flows." + flowID) {
			logging.Main.WithFields(log.Fields{
				"flowId": flowID,
				"path":   "flows." + flowID,
			}).Fatalf("flow not found in config file")
		}

		var (
			instanceID, identifier, input string
			json                          map[string]map[string]string
		)

		if !flowCmdViper.GetBool("all") {
			instanceID = utils.SelectInstance(instanceIds, searchPattern)
		}
		for _, cmd := range config.XT.Flows[flowID] {
			if flowCmdViper.GetBool("all") {
				remoteCmd := strings.Split(cmd.Run, " ")
				executers := ssh.CreateSSHExecuters(instanceIds, "", remoteCmd)
				ssh.RunMultiple(executers)
			} else {
				cmd.Run = strings.Replace(cmd.Run, "__IDENTEFIER__", identifier, -1)
				remoteCmd := strings.Split(cmd.Run, " ")
				ctx := logging.Main.WithFields(log.Fields{
					"host":    instanceID,
					"command": remoteCmd,
				})
				executer := ssh.NewSSHExecuter(instanceID, remoteCmd)
				if cmd.TTY {
					ssh.CommandWithTTY(executer)
				} else {
					output := ssh.CommandWithOutput(executer)

					if output.Stderr != "" {
						ctx.Fatalf("failed running command, received %s", output.Stderr)
					}
					input = output.Stdout
				}
				if cmd.Print {
					ctx.Info("printing info...")
					ctx = logging.Main.WithFields(log.Fields{
						"root":       cmd.Root,
						"identifier": cmd.Identifier,
						"keys":       cmd.Keys,
						"print":      cmd.Print,
						"type":       cmd.Type,
					})
					switch cmd.Type {
					case "json":
						if cmd.Root == "" && cmd.Identifier != "" {
							ctx.Fatalf("json data parsing requires root path of array and an identifier")
						}
						json, err = utils.UnmarshalKeys(input, cmd)
						if err != nil {
							ctx.Fatalf("parse remote command data failed, %v", err)
						}
						utils.PrintInfo(cmd.Keys, json)
						break
					default:
						fmt.Println(input)
					}
				}
				if cmd.Identifier != "" {
					identifier = utils.SelectInstance(utils.GetIds(json), cmd.Identifier)
				}
			}
		}
	},
}
