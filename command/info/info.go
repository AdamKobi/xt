package infocmd

import (
	"github.com/adamkobi/xt/command/factory"
	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/provider"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Options struct {
	Config        func() (*config.Config, error)
	Log           func() *logrus.Logger
	Debug         func()
	SearchPattern string
	Profile       string
	Tag           string
}

//NewCmdInfo creates an info command
func NewCmdInfo(f *factory.CmdConfig) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		Debug:  f.Debug,
		Log:    f.Log,
	}
	cmd := &cobra.Command{
		Use:   "info <servers>",
		Short: "Gather data on instances and print as table",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SearchPattern = args[0]
			opts.Tag, _ = cmd.Flags().GetString("tag")
			opts.Profile, _ = cmd.Flags().GetString("profile")
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				opts.Debug()
			}

			return runInfo(opts)
		},
	}

	return cmd
}

func runInfo(opts *Options) error {
	cfg, _ := opts.Config()
	profile, err := cfg.Profile(opts.Profile)
	if err != nil {
		return err
	}

	profile.Message()

	providerOpts := &provider.Options{
		Name:          profile.ProviderOptions.Name,
		VPC:           profile.ProviderOptions.VPC,
		Region:        profile.ProviderOptions.Region,
		CredsProfile:  profile.ProviderOptions.CredsProfile,
		Tag:           opts.Tag,
		SearchPattern: opts.SearchPattern,
	}

	svcProvider, err := provider.New(providerOpts)
	if err != nil {
		return err
	}

	if err := svcProvider.Instances(); err != nil {
		return err
	}
	svcProvider.Print()
	return nil
}
