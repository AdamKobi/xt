package file

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/adamkobi/xt/pkg/cmdutil"
	getCmd "github.com/adamkobi/xt/pkg/command/file/get"
	putCmd "github.com/adamkobi/xt/pkg/command/file/put"
	"github.com/spf13/cobra"
)

func NewCmdFile(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file <command>",
		Short: "Work with remote files",
		Long:  "Upload and Download files from remote servers using SCP protocol",
		Example: heredoc.Doc(`
				$ xt file put localfile.json /tmp/remotefile.json web
				$ xt file get /tmp/remotefile.json localfile.json web
		`),
	}

	cmd.AddCommand(getCmd.NewCmdDownload(f))
	cmd.AddCommand(putCmd.NewCmdUpload(f))
	return cmd
}
