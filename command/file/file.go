package file

import (
	"github.com/adamkobi/xt/command/factory"
	downloadCmd "github.com/adamkobi/xt/command/file/download"
	uploadCmd "github.com/adamkobi/xt/command/file/upload"
	"github.com/spf13/cobra"
)

func NewCmdFile(f *factory.CmdConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "file <command>",
		Short:   "Work with remote files",
		Aliases: []string{"f"},
	}

	cmd.AddCommand(downloadCmd.NewCmdDownload(f))
	cmd.AddCommand(uploadCmd.NewCmdUpload(f))
	return cmd
}
