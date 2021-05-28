package get

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

	Profile  string
	Tag      string
	Dest     string
	All      bool
	Download bool
}

//NewCmdDownload creates a new download command
func NewCmdDownload(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		IO:     f.IOStreams,
	}
	cmd := &cobra.Command{
		Use:   "get <remotepath> <localpath> <servers> [flags]",
		Short: "Download files from remote servers",
		Long: heredoc.Doc(`
			Use SCP to download files from remote servers.

			Support downloading from multiple servers, downloaded files will have the servers prefix.
		`),
		Example: heredoc.Doc(`
			# download file from remote server
			$ xt file get /tmp/remotefile.json localfile.json  web

			# download file from multiple servers
			$ xt file get -a /tmp/remotefile.json .  web
		`),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.RemotePath = args[0]
			opts.LocalPath = args[1]
			opts.SearchPattern = strings.TrimSuffix(args[2], "*")
			opts.Tag, _ = cmd.Flags().GetString("tag")
			opts.Profile, _ = cmd.Flags().GetString("profile")

			return runDownload(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "Download files from all servers matching search pattern")
	cmd.Flags().StringVarP(&opts.Dest, "dest", "d", "xt-downloads", "Output destination for get command")
	return cmd
}

func runDownload(opts *Options) error {
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
		IO:         opts.IO,
		User:       profile.SSHOptions.User,
		Domain:     profile.SSHOptions.Domain,
		Binary:     executer.SCP,
		Args:       profile.SCPArgs(),
		Hostnames:  instances.Names(),
		LocalPath:  opts.LocalPath,
		RemotePath: opts.RemotePath,
		Download:   true,
	}

	if !opts.All {
		cmdOpts.Selected, err = utils.Select(opts.IO, instances.Names(), opts.SearchPattern)
		if err != nil {
			return err
		}

		c, err := executer.New(cmdOpts)
		if err != nil {
			return err
		}

		return c.Connect()
	}
	executers, err := executer.CreateAll(cmdOpts)
	if err != nil {
		return err
	}

	executer.RunCommands(opts.IO, executers)
	return nil
}
