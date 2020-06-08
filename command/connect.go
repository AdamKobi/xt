package command

import (
	"github.com/adamkobi/xt/config"
	log "github.com/adamkobi/xt/pkg/logging"
	"github.com/adamkobi/xt/pkg/providers"
	"github.com/adamkobi/xt/pkg/ssh"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var connectCmdViper = viper.New()

func init() {
	config.InitViper(connectCmdViper)
	RootCmd.AddCommand(connectCmd)
}

var connectCmd = &cobra.Command{
	Use:   "connect SERVERS",
	Short: "SSH to server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tag := viper.GetString("tag")
		searchPattern := args[0] + "*"
		instances, err := providers.GetInstances(tag, searchPattern)
		if err != nil {
			log.Main.Fatal(err.Error())
		}
		var instNames []string
		for _, instance := range instances {
			instNames = append(instNames, utils.GetIds(instance)...)
		}
		instance := utils.SelectInstance(instNames, searchPattern)
		log.Main.Infof("Connecting to %s...", instance)
		ssh.Connect(instance)
	},
}
