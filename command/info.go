package command

import (
	"github.com/adamkobi/xt/config"
	log "github.com/adamkobi/xt/pkg/logging"
	"github.com/adamkobi/xt/pkg/providers"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var infoCmdViper = viper.New()

func init() {
	config.InitViper(infoCmdViper)
	RootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info SERVERS",
	Short: "Gather data on instances and print as table",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tag := viper.GetString("tag")
		searchPattern := args[0] + "*"
		profile := config.GetProfile()
		instances, err := providers.GetInstances(tag, searchPattern)
		if err != nil {
			log.Main.Fatal(err.Error())
		}

		for _, providerInstances := range instances {
			utils.PrintInfo(profile.GetStringSlice("commands.info.fields"), providerInstances)
		}
	},
}
