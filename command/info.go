package command

import (
	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/providers"
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
		searchPattern := args[0]
		instances, err := providers.GetInstances(searchPattern)
		if err != nil {
			logger.Fatal(err.Error())
		}

		providers.PrintInfo(instances)
	},
}
