package list

import (
	"fmt"

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
}

func NewCmdList(f *factory.CmdConfig) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		Debug:  f.Debug,
		Log:    f.Log,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List flows",
		RunE: func(cmd *cobra.Command, args []string) error {
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				opts.Debug()
			}

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

	fmt.Print(string(d))
	return nil
}
