package put

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
	LocalPath     string
	RemotePath    string

	Profile string
	Tag     string
	All     bool
}

//NewCmdUpload creates a new upload command
func NewCmdUpload(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		IO:     f.IOStreams,
	}
	cmd := &cobra.Command{
		Use:   "put <localpath> <remotepath> <servers> [flags]",
		Short: "Upload files to remote servers",
		Long: heredoc.Doc(`
			Use SCP to upload files to remote servers.

			Support uploading to multiple servers.
		`),
		Example: heredoc.Doc(`
			# upload file to remote server
			$ xt file put localfile.json /tmp/ web

			# upload file to multiple servers
			$ xt file put localfile.json /tmp/ web -a
		`),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.LocalPath = args[0]
			opts.RemotePath = args[1]
			opts.SearchPattern = strings.TrimSuffix(args[2], "*")
			opts.Tag, _ = cmd.Flags().GetString("tag")
			opts.Profile, _ = cmd.Flags().GetString("profile")

			return runUpload(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "Upload files to all remote servers matching search pattern")
	return cmd
}

func runUpload(opts *Options) error {
	cfg, _ := opts.Config()
	profile, err := cfg.Profile(opts.Profile)
	if err != nil {
		return err
	}

	cs := opts.IO.ColorScheme()
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
		IO:         opts.IO,
		User:       profile.SSHOptions.User,
		Domain:     profile.SSHOptions.Domain,
		Binary:     executer.SCP,
		Args:       profile.SCPArgs(),
		Hostnames:  instances.Names(),
		LocalPath:  opts.LocalPath,
		RemotePath: opts.RemotePath,
	}

	if !opts.All {
		cmdOpts.Selected, err = utils.Select(opts.IO, instances.Names(), opts.SearchPattern)
		if err != nil {
			return err
		}

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
