package connect

import (
	"github.com/adamkobi/xt/command/factory"
	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/executer"
	"github.com/adamkobi/xt/pkg/provider"
	"github.com/adamkobi/xt/pkg/utils"
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

//NewCmdConnect creates a connect command
func NewCmdConnect(f *factory.CmdConfig) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		Debug:  f.Debug,
		Log:    f.Log,
	}
	cmd := &cobra.Command{
		Use:     "connect <servers>",
		Short:   "SSH to server",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"c", "co", "con"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SearchPattern = args[0]
			opts.Tag, _ = cmd.Flags().GetString("tag")
			opts.Profile, _ = cmd.Flags().GetString("profile")
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				opts.Debug()
			}

			return runConnect(opts)
		},
	}

	return cmd
}

func runConnect(opts *Options) error {
	cfg, _ := opts.Config()
	log := opts.Log()
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

	executerOptions := &executer.Options{
		User:   profile.SSHOptions.User,
		Domain: profile.SSHOptions.Domain,
		Binary: executer.SSH,
		Args:   profile.SSHArgs(),
	}

	executerOptions.Hostname, err = utils.Select(svcProvider.Names())
	if err != nil {
		return err
	}

	executer, err := executer.New(executerOptions)
	if err != nil {
		return err
	}
	log.Infof("connecting to %s", executerOptions.Hostname)
	executer.Connect()
	return nil
}
