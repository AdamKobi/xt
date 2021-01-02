package download

import (
	"path"
	"time"

	"github.com/adamkobi/xt/command/factory"
	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/executer"
	"github.com/adamkobi/xt/pkg/provider"
	"github.com/apex/log"
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
	LocalPath     string
	RemotePath    string
	All           bool
}

//NewCmdDownload creates a new download command
func NewCmdDownload(f *factory.CmdConfig) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		Debug:  f.Debug,
		Log:    f.Log,
	}
	cmd := &cobra.Command{
		Use:   "download <servers> <localpath> <remotepath>",
		Short: "download files from remote servers",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SearchPattern = args[0]
			opts.LocalPath = args[1]
			opts.RemotePath = args[2]
			opts.Tag, _ = cmd.Flags().GetString("tag")
			opts.Profile, _ = cmd.Flags().GetString("profile")
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				opts.Debug()
			}
			return runDownload(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "download files from all servers matching search pattern, files will be have the hostname prefix")
	return cmd
}

//CreateSCPExecuters creates multiple SCP executers with same local and remote file paths
func createSCPExecuters(opts *executer.Options) ([]executer.Factory, error) {
	var executers []executer.Factory
	for _, host := range opts.Hostnames {
		opts.Hostname = host
		executer, err := executer.New(opts)
		if err != nil {
			return nil, err
		}
		executers = append(executers, *executer)
	}
	return executers, nil
}

func createSSHExecuters(opts *executer.Options) ([]executer.Factory, error) {
	var executers []executer.Factory
	for _, host := range opts.Hostnames {
		localDir := path.Dir(opts.LocalPath)
		remoteFile := path.Base(opts.RemotePath)
		opts.Hostname = host
		opts.LocalPath = path.Join(localDir, host+"_"+remoteFile)
		executer, err := executer.New(opts)
		if err != nil {
			return nil, err
		}
		executers = append(executers, *executer)
	}
	return executers, nil
}

func uploadAll(opts *executer.Options) error {
	executers, err := createSCPExecuters(opts)
	if err != nil {
		return err
	}
	output := make(chan executer.Output, len(executers))
	timeout := time.After(60 * time.Second)
	executer.RunCommands(executers, output)
	for _, e := range executers {
		select {
		case out := <-output:
			if len(out.Stderr) != 0 {
				log.Errorf("command failed on host %s", e.Options.Hostname)
				log.Error(string(out.Stderr))
			} else {
				log.Infof("command succuss on host %s", e.Options.Hostname)
				log.Info(string(out.Stdout))
			}
		case <-timeout:
			log.Warn("command timedout")
		}
	}
	return nil
}

func runDownload(opts *Options) error {
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

	executerOptions := &executer.Options{
		User:       profile.SSHOptions.User,
		Domain:     profile.SSHOptions.Domain,
		Binary:     executer.SCP,
		Args:       profile.SCPArgs(),
		Hostnames:  svcProvider.Names(),
		LocalPath:  opts.LocalPath,
		RemotePath: opts.RemotePath,
	}

	if opts.All {
		uploadAll(executerOptions)
	} else {
		executerOptions.Hostname, err = provider.SelectHost(svcProvider.Names(), opts.SearchPattern)
		if err != nil {
			return err
		}
		if e, err := executer.New(executerOptions); err == nil {
			return e.CommandWithTTY()
		}
		return err
	}
	return nil
}
