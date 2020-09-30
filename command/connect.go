package command

import (
	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/providers"
	"github.com/adamkobi/xt/pkg/ssh"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	connectCmdViper = viper.New()
)

func init() {
	config.InitViper(connectCmdViper)
	RootCmd.AddCommand(connectCmd)
}

var connectCmd = &cobra.Command{
	Use:   "connect SERVERS",
	Short: "SSH to server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchPattern := args[0]
		instances, err := providers.GetIds(searchPattern)
		if err != nil {
			logger.Fatal(err.Error())
		}
		instance := utils.SelectInstance(instances, searchPattern)
		logger.Infof("Connecting to %s...", instance)
		ssh.Connect(instance)
	},
}
