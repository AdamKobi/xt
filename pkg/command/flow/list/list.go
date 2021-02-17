package list

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/adamkobi/xt/internal/config"
	"github.com/adamkobi/xt/pkg/cmdutil"
	"github.com/adamkobi/xt/pkg/iostreams"
	"github.com/spf13/cobra"
)

type Options struct {
	Config func() (*config.Config, error)
	IO     *iostreams.IOStreams
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		IO:     f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List flows",
		Long:  "List all flows from config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts)
		},
	}

	return cmd
}

func runList(opts *Options) error {
	cfg, _ := opts.Config()
	flows := cfg.Flows()
	if len(flows) == 0 {
		return fmt.Errorf("flows not found in config file")
	}

	d, err := yaml.Marshal(flows)
	if err != nil {
		return err
	}

	fmt.Fprint(opts.IO.Out, string(d))
	return nil
}
