package delete

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

	FlowID string
}

func NewCmdDelete(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		IO:     f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "delete <flow>",
		Short: "Delete flow",
		Long:  "Delete flow, will remove this from config file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.FlowID = args[0]
			return runDelete(opts)
		},
	}

	return cmd
}

func runDelete(opts *Options) error {
	cfg, _ := opts.Config()
	cs := opts.IO.ColorScheme()
	_, err := cfg.Flow(opts.FlowID)
	if err != nil {
		return err
	}

	//TODO add confirm before delete
	delete(cfg.FlowOptions, opts.FlowID)

	d, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	err = config.WriteDefaultConfigFile(d)
	if err != nil {
		return err
	}
	fmt.Fprintf(opts.IO.Out, "%s flow %s deleted successfully", cs.SuccessIcon())
	return nil
}
