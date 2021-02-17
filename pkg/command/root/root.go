package root

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/adamkobi/xt/pkg/cmdutil"
	connectCmd "github.com/adamkobi/xt/pkg/command/connect"
	runCmd "github.com/adamkobi/xt/pkg/command/run"

	fileCmd "github.com/adamkobi/xt/pkg/command/file"
	flowCmd "github.com/adamkobi/xt/pkg/command/flow"
	infoCmd "github.com/adamkobi/xt/pkg/command/info"

	versionCmd "github.com/adamkobi/xt/pkg/command/version"

	"github.com/spf13/cobra"
)

func NewCmd(f *cmdutil.Factory, version, buildDate string) *cobra.Command {
	cobra.EnableCommandSorting = false

	cmd := &cobra.Command{
		Use:           "xt <command> <subcommand> [flags]",
		Short:         "Xt provides cloud utils",
		Long:          `Connectivity to multiple cloud providers with human readable names.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Example: heredoc.Doc(`
			$ xt connect hostGroup
			$ xt info hostGroup
			$ xt flow create connect-pods
		`),
	}

	cmd.SetOut(f.IOStreams.Out)
	cmd.SetErr(f.IOStreams.ErrOut)

	cs := f.IOStreams.ColorScheme()

	helpHelper := func(command *cobra.Command, args []string) {
		rootHelpFunc(cs, command, args)
	}
	cfg, _ := f.Config()

	cmd.PersistentFlags().Bool("help", false, "Show help for command")
	cmd.SetHelpFunc(helpHelper)
	cmd.SetUsageFunc(rootUsageFunc)
	cmd.SetFlagErrorFunc(rootFlagErrorFunc)

	formattedVersion := versionCmd.Format(version, buildDate)
	cmd.SetVersionTemplate(formattedVersion)
	cmd.Version = formattedVersion
	cmd.Flags().Bool("version", false, "Show xt version")

	cmd.PersistentFlags().StringP("profile", "p", cfg.DefaultProfile(), fmt.Sprint("Select profile to use (required): ", strings.Join(cfg.Profiles(), "|")))
	cmd.PersistentFlags().StringP("tag", "t", "Name", "Search instances by this tag")

	// Child commands
	cmd.AddCommand(versionCmd.NewCmdVersion(f, version, buildDate))
	cmd.AddCommand(connectCmd.NewCmdConnect(f))
	cmd.AddCommand(infoCmd.NewCmdInfo(f))
	cmd.AddCommand(runCmd.NewCmdRun(f))
	cmd.AddCommand(flowCmd.NewCmdFlow(f))
	cmd.AddCommand(fileCmd.NewCmdFile(f))

	return cmd
}
