package connect

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/adamkobi/xt/internal/config"
	"github.com/adamkobi/xt/internal/instance"
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
	Profile       string
	Tag           string
}

//NewCmdConnect creates a connect command
func NewCmdConnect(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		IO:     f.IOStreams,
	}
	cmd := &cobra.Command{
		Use:   "connect <servers>",
		Short: "SSH to server",
		Long: heredoc.Doc(`
				Connect to remote servers by specifying a human readable name.

				Query by tag name, can specify part of the tag as well.
		`),
		Example: heredoc.Doc(`
				# query server group webserver with default tag and profile
				$ xt connect web

				# query server group webserver with custom tag role
				$ xt connect -t role web

				# query server group webserver with production profile
				$ xt connect -p production web
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SearchPattern = strings.TrimSuffix(args[0], "*")
			opts.Tag, _ = cmd.Flags().GetString("tag")
			opts.Profile, _ = cmd.Flags().GetString("profile")

			return runConnect(opts)
		},
	}

	return cmd
}

func runConnect(opts *Options) error {
	cfg, _ := opts.Config()
	cs := opts.IO.ColorScheme()
	profile, err := cfg.Profile(opts.Profile)
	if err != nil {
		return err
	}

	if profile.DisplayMsg != "" {
		fmt.Fprintf(opts.IO.Out, cs.Red("%s"), profile.Message())
	}

	var instances instance.XTInstances
	for _, p := range profile.ProviderOptions {
		opts := &provider.Options{
			Name:          p.Name,
			VPC:           p.VPC,
			Region:        p.Region,
			CredsProfile:  p.CredsProfile,
			Tag:           opts.Tag,
			SearchPattern: opts.SearchPattern,
		}
		svc, err := provider.New(opts)
		if err != nil {
			return err
		}

		i, err := svc.Get()
		if err != nil {
			return err
		}
		instances = append(instances, i...)
	}

	cmdOpts := &executer.Options{
		IO:     opts.IO,
		User:   profile.SSHOptions.User,
		Domain: profile.SSHOptions.Domain,
		Binary: executer.SSH,
		Args:   profile.SSHArgs(),
	}

	cmdOpts.Selected, err = utils.Select(opts.IO, instances.Names(), opts.SearchPattern)
	if err != nil {
		return err
	}

	e, err := executer.New(cmdOpts)
	if err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Out, "connecting to %s\n", cs.Bold(cmdOpts.Selected))
	return e.Connect()
}
