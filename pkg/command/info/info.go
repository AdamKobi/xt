package infocmd

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/adamkobi/xt/internal/config"
	"github.com/adamkobi/xt/internal/instance"
	"github.com/adamkobi/xt/pkg/cmdutil"
	"github.com/adamkobi/xt/pkg/iostreams"
	"github.com/adamkobi/xt/pkg/provider"
	"github.com/spf13/cobra"
)

type Options struct {
	Config        func() (*config.Config, error)
	IO            *iostreams.IOStreams
	SearchPattern string
	Profile       string
	Tag           string
}

//NewCmdInfo creates an info command
func NewCmdInfo(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		IO:     f.IOStreams,
	}
	cmd := &cobra.Command{
		Use:   "info <servers>",
		Short: "Gather data on instances and print as formated table",
		Long:  "Query data from api and print general data on instances, such as names, IPs and status",
		Example: heredoc.Doc(`
				$ xt info web
				$ xt -p production info mongo
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SearchPattern = strings.TrimSuffix(args[0], "*")
			opts.Tag, _ = cmd.Flags().GetString("tag")
			opts.Profile, _ = cmd.Flags().GetString("profile")

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
	instances.Print(opts.IO)
	return nil
}
