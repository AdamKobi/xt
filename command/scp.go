package command

import (
	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/providers"
	"github.com/adamkobi/xt/pkg/ssh"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scpCmdViper = viper.New()

func init() {
	config.InitViper(scpCmdViper)
	RootCmd.AddCommand(scpCmd)
	scpCmd.PersistentFlags().BoolP("upload", "u", false, "upload file to remote server (default is download)")
	scpCmd.PersistentFlags().BoolP("all", "a", false, "run flow on all servers")
	scpCmdViper.BindPFlags(scpCmd.PersistentFlags())
}

var scpCmd = &cobra.Command{
	Use:   "scp SERVERS LocalPath RemotePath",
	Short: "upload/download files from remote servers",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		searchPattern := args[0]
		localFilePath := args[1]
		remoteFilePath := args[2]

		instanceIDs, err := providers.GetIds(searchPattern)
		if err != nil {
			logger.Fatalf("fetching instance ids failed, %v", err)
		}

		upload := scpCmdViper.GetBool("upload")
		if scpCmdViper.GetBool("all") {
			e := ssh.CreateSCPExecuters(instanceIDs, localFilePath, remoteFilePath, upload)
			ssh.RunMany(e)
		} else {
			instanceID := utils.SelectInstance(instanceIDs, searchPattern)
			logger.Infof("Connecting to %s...", instanceID)
			e := ssh.NewSCPExecuter(instanceID, localFilePath, remoteFilePath, upload)
			e.CommandWithTTY()
		}
	},
}
