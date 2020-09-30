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
		searchPattern := args[0]
		flowID := args[1]

		instanceIDs, err := providers.GetIds(searchPattern)
		if err != nil {
			logger.Fatalf("fetching instance ids failed, %v", err)
		}

		if !flowCmdViper.IsSet("flows." + flowID) {
			logger.Fatalf("%s not found in config file", "flows["+flowID+"]")
		}

		var (
			instanceID, identifier, input string
			json                          map[string]map[string]string
		)

		if !flowCmdViper.GetBool("all") {
			instanceID = utils.SelectInstance(instanceIDs, searchPattern)
		}

		flows := config.GetFlows()
		for _, cmd := range flows[flowID] {

			if flowCmdViper.GetBool("all") {
				remoteCmd := strings.Split(cmd.Run, " ")
				executers := ssh.CreateSSHExecuters(instanceIDs, remoteCmd)
				ssh.RunMany(executers)
			} else {
				cmd.Run = strings.Replace(cmd.Run, "__IDENTEFIER__", identifier, -1)
				remoteCmd := strings.Split(cmd.Run, " ")
				ctx := logger.WithFields(log.Fields{
					"host":    instanceID,
					"command": remoteCmd,
				})
				e := ssh.NewSSHExecuter(instanceID, remoteCmd)
				if cmd.TTY {
					e.CommandWithTTY()
				} else {
					output := e.CommandWithOutput()

					if output.Error {
						ctx.Fatalf("command failed")
						fmt.Fprint(os.Stderr, output.Data)
					}
					input = output.Data
				}
				if cmd.Print {
					ctx.Info("printing info...")
					ctx = logger.WithFields(log.Fields{
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
