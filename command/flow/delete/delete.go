package delete

import (
	"gopkg.in/yaml.v3"

	"github.com/adamkobi/xt/command/factory"
	"github.com/adamkobi/xt/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Options struct {
	Config func() (*config.Config, error)
	Log    func() *logrus.Logger
	Debug  func()
	FlowID string
}

func NewCmdDelete(f *factory.CmdConfig) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		Debug:  f.Debug,
		Log:    f.Log,
	}

	cmd := &cobra.Command{
		Use:   "delete <flow>",
		Short: "Delete flow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.FlowID = args[0]
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				opts.Debug()
			}

			return runDelete(opts)
		},
	}

	return cmd
}

func runDelete(opts *Options) error {
	cfg, _ := opts.Config()
	_, err := cfg.Flow(opts.FlowID)
	if err != nil {
		return err
	}

	delete(cfg.FlowOptions, opts.FlowID)

	d, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return config.WriteDefaultConfigFile(d)
}
