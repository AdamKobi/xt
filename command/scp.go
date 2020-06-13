package command

import (
	"fmt"

	"github.com/adamkobi/xt/config"
	log "github.com/adamkobi/xt/pkg/logging"
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
		tag := viper.GetString("tag")
		searchPattern := args[0] + "*"
		localFilePath := args[1]
		remoteFilePath := args[2]

		instanceIds, err := providers.GetIds(tag, searchPattern)
		if err != nil {
			log.Main.Fatalf("fetching instance ids failed, %v", err)
		}

		upload := scpCmdViper.GetBool("upload")
		if scpCmdViper.GetBool("all") {
			executers := ssh.CreateSCPExecuters(instanceIds, localFilePath, remoteFilePath, upload)
			ssh.RunMultiple(executers)
		} else {
			instanceID := utils.SelectInstance(instanceIds, searchPattern)
			log.Main.Infof("Connecting to %s...", instanceID)
			executer := ssh.NewSCPExecuter(instanceID, localFilePath, remoteFilePath, upload)
			fmt.Println(executer)
			ssh.CommandWithTTY(executer)
		}
	},
}
