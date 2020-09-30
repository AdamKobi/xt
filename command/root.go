package command

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is dynamically set by the toolchain or overridden by the Makefile.
var (
	Version = "DEV"
	logger  = logging.GetLogger()
)

func init() {
	if Version == "DEV" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
	config.InitConfig()

	RootCmd.Version = strings.TrimPrefix(Version, "v")
	RootCmd.AddCommand(versionCmd)

	RootCmd.PersistentFlags().StringP("profile", "p", config.DefaultProfile(), fmt.Sprint("Select profile to use (required): ", strings.Join(config.GetProfiles(), "|")))
	RootCmd.PersistentFlags().StringP("tag", "t", "Name", "Search instances by this tag")
	RootCmd.PersistentFlags().Bool("help", false, "Show help for command")
	RootCmd.Flags().Bool("version", false, "Show xt version")
	viper.BindPFlags(RootCmd.PersistentFlags())
}

// RootCmd is the entry point of command-line execution
var RootCmd = &cobra.Command{
	Use:   "xt",
	Short: "Cloud utils",
	Long:  `Xt provides different cloud utils`,

	SilenceErrors: true,
	SilenceUsage:  true,
}

var versionCmd = &cobra.Command{
	Use:    "version",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("xt version ", RootCmd.Version)
	},
}
