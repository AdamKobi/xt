package root

import (
	"fmt"
	"strings"

	connectCmd "github.com/adamkobi/xt/command/connect"
	execCmd "github.com/adamkobi/xt/command/exec"
	"github.com/adamkobi/xt/command/factory"

	downloadCmd "github.com/adamkobi/xt/command/download"
	flowCmd "github.com/adamkobi/xt/command/flow"
	infoCmd "github.com/adamkobi/xt/command/info"
	uploadCmd "github.com/adamkobi/xt/command/upload"

	versionCmd "github.com/adamkobi/xt/command/version"

	"github.com/spf13/cobra"
)

func NewCmdRoot(f *factory.CmdConfig, version, buildDate string) *cobra.Command {
	cobra.EnableCommandSorting = false

	cmd := &cobra.Command{
		Use:           "xt <command> <subcommand> [flags]",
		Short:         "Xt provides cloud utils",
		Long:          `Connectivity to multiple cloud providers with human readable names.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cfg, _ := f.Config()

	cmd.PersistentFlags().StringP("profile", "p", cfg.DefaultProfile(), fmt.Sprint("Select profile to use (required): ", strings.Join(cfg.Profiles(), "|")))
	cmd.PersistentFlags().StringP("tag", "t", "Name", "Search instances by this tag")
	cmd.PersistentFlags().BoolP("debug", "", false, "Debug level")
	cmd.PersistentFlags().Bool("help", false, "Show help for command")
	cmd.PersistentFlags().Bool("version", false, "Show xt version")

	// Child commands
	cmd.AddCommand(versionCmd.NewCmdVersion(version, buildDate))
	cmd.AddCommand(connectCmd.NewCmdConnect(f))
	cmd.AddCommand(infoCmd.NewCmdInfo(f))
	cmd.AddCommand(execCmd.NewCmdExec(f))
	cmd.AddCommand(flowCmd.NewCmdFlow(f))
	cmd.AddCommand(uploadCmd.NewCmdUpload(f))
	cmd.AddCommand(downloadCmd.NewCmdDownload(f))

	return cmd
}
