package run

import (
	"fmt"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/adamkobi/xt/internal/config"
	"github.com/adamkobi/xt/pkg/cmdutil"
	"github.com/adamkobi/xt/pkg/executer"
	"github.com/adamkobi/xt/pkg/iostreams"
	"github.com/adamkobi/xt/pkg/provider"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/spf13/cobra"
)

type Options struct {
	Config func() (*config.Config, error)
	IO     *iostreams.IOStreams

	SearchPattern string
	RemoteCmd     []string

	Profile string
	Tag     string
	All     bool
	Force   bool
}

//NewCmdRun creates an exec command
func NewCmdRun(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		IO:     f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "run <servers> <command> [flags]",
		Short: "Execute remote commands",
		Long:  "Execute commands on one or more remote servers and return output",
		Example: heredoc.Doc(`
				$ xt run web "ls -la"
				$ xt run -af web "cat ~/.bash_profile"
		`),
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SearchPattern = strings.TrimSuffix(args[0], "*")
			opts.RemoteCmd = args[1:]
			opts.Tag, _ = cmd.Flags().GetString("tag")
			opts.Profile, _ = cmd.Flags().GetString("profile")

			return runCmds(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "run command on all servers matching search pattern")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "run command without requesting approval")
	return cmd
}

func runCmds(opts *Options) error {
	cfg, _ := opts.Config()
	profile, err := cfg.Profile(opts.Profile)
	if err != nil {
		return err
	}

	cs := opts.IO.ColorScheme()
	if profile.DisplayMsg != "" {
		fmt.Fprintf(opts.IO.Out, cs.Red("%s"), profile.Message())
	}

	providerOpts := &provider.Options{
		Name:          profile.ProviderOptions.Name,
		VPC:           profile.ProviderOptions.VPC,
		Region:        profile.ProviderOptions.Region,
		CredsProfile:  profile.ProviderOptions.CredsProfile,
		Tag:           opts.Tag,
		SearchPattern: opts.SearchPattern,
	}

	svc, err := provider.New(providerOpts)
	if err != nil {
		return err
	}

	instances, err := svc.Get()
	if err != nil {
		return err
	}

	cmdOpts := &executer.Options{
		IO:        opts.IO,
		User:      profile.SSHOptions.User,
		Domain:    profile.SSHOptions.Domain,
		Binary:    executer.SSH,
		Args:      profile.SSHArgs(),
		RemoteCmd: opts.RemoteCmd,
		Hostnames: instances.Names(),
	}

	if !opts.All {
		cmdOpts.Selected, err = utils.Select(opts.IO, instances.Names(), opts.SearchPattern)
		if err != nil {
			return err
		}
		cmdOpts.Hostnames = []string{cmdOpts.Selected}
	}

	if !opts.Force {
		var approved bool
		err = survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf(
				"Will Execute\n$ %s\nOn\n%s\n\n",
				strings.Join(cmdOpts.RemoteCmd, " "),
				strings.Join(cmdOpts.Hostnames, "\n")),
		}, &approved)
		if err != nil {
			return err
		}
		if !approved {
			return fmt.Errorf("command cancelled")
		}
	}

	if !opts.All {
		e, err := executer.New(cmdOpts)
		if err != nil {
			return err
		}
		return e.Connect()

	}
	executers, err := executer.CreateAll(cmdOpts)
	if err != nil {
		return err
	}
	executer.RunCommands(opts.IO, executers)
	return nil
}
