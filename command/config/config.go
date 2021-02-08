package config

import (
	"github.com/adamkobi/xt/command/factory"
	"github.com/adamkobi/xt/pkg/executer"
	"github.com/adamkobi/xt/pkg/provider"
	"github.com/spf13/cobra"
)

//NewCmdConnect creates a connect command
func NewCmdConfig(f *factory.CmdConfig) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		Debug:  f.Debug,
		Log:    f.Log,
	}
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Config commands",
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

	executerOptions.Hostname, err = provider.SelectHost(svcProvider.Names(), opts.SearchPattern)
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
