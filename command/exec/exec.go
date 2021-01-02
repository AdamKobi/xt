package exec

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/adamkobi/xt/command/factory"
	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/executer"
	"github.com/adamkobi/xt/pkg/provider"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Options struct {
	Config        func() (*config.Config, error)
	Log           func() *logrus.Logger
	Debug         func()
	SearchPattern string
	Profile       string
	Tag           string
	RemoteCmd     []string
	All           bool
	Force         bool
}

//NewCmdExec creates an exec command
func NewCmdExec(f *factory.CmdConfig) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		Debug:  f.Debug,
		Log:    f.Log,
	}

	cmd := &cobra.Command{
		Use:     "exec <servers> <command> [flags]",
		Short:   "Execute remote commands",
		Args:    cobra.MinimumNArgs(2),
		Aliases: []string{"x", "ex", "exe"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SearchPattern = args[0]
			opts.RemoteCmd = args[1:]
			opts.Tag, _ = cmd.Flags().GetString("tag")
			opts.Profile, _ = cmd.Flags().GetString("profile")
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				opts.Debug()
			}
			return runExec(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "run command on all servers matching search pattern")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "run command without requesting approval")
	return cmd
}

func checkUserApproval(opts *executer.Options) bool {
	if opts.Hostname != "" {
		return utils.GetApproval(strings.Join(opts.RemoteCmd, " "), []string{opts.Hostname})
	}
	return utils.GetApproval(strings.Join(opts.RemoteCmd, " "), opts.Hostnames)
}

func createSSHExecuters(opts *executer.Options) ([]executer.Factory, error) {
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

func runAll(opts *executer.Options) error {
	executers, err := createSSHExecuters(opts)
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
				fmt.Fprintf(os.Stderr, string(out.Stderr))
			} else {
				log.Infof("command succuss on host %s", e.Options.Hostname)
				fmt.Fprintf(os.Stdout, string(out.Stdout))
			}
		case <-timeout:
			log.Warn("command timedout")
		}
	}
	return nil
}

func runExec(opts *Options) error {
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
		User:      profile.SSHOptions.User,
		Domain:    profile.SSHOptions.Domain,
		Binary:    executer.SSH,
		Args:      profile.SSHArgs(),
		RemoteCmd: opts.RemoteCmd,
		Hostnames: svcProvider.Names(),
	}

	if !opts.All {
		executerOptions.Hostname, err = provider.SelectHost(svcProvider.Names(), opts.SearchPattern)
		if err != nil {
			return err
		}
	}

	if !opts.Force && !checkUserApproval(executerOptions) {
		return fmt.Errorf("command cancelled")
	}

	if opts.All {
		runAll(executerOptions)
	} else {
		if e, err := executer.New(executerOptions); err == nil {
			return e.CommandWithTTY()
		}
		return err
	}
	return nil
}
