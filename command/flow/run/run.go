package run

import (
	"bytes"
	"fmt"
	"strings"

	"text/template"

	"github.com/adamkobi/xt/command/factory"
	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/executer"
	"github.com/adamkobi/xt/pkg/provider"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Options struct {
	Config        func() (*config.Config, error)
	Log           func() *logrus.Logger
	Debug         func()
	SearchPattern string
	Profile       string
	Tag           string
	FlowID        string
	All           bool
}

func NewCmdRun(f *factory.CmdConfig) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		Debug:  f.Debug,
		Log:    f.Log,
	}

	cmd := &cobra.Command{
		Use:   "run <servers> <flow> [flags]",
		Short: "Execute multiple remote commands from config file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SearchPattern = args[0]
			opts.FlowID = args[1]
			opts.Tag, _ = cmd.Flags().GetString("tag")
			opts.Profile, _ = cmd.Flags().GetString("profile")
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				opts.Debug()
			}

			return runFlow(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "run flow on all servers")
	return cmd
}
func runFlow(opts *Options) error {
	cfg, _ := opts.Config()
	profile, err := cfg.Profile(opts.Profile)
	if err != nil {
		return err
	}

	flow, err := cfg.Flow(opts.FlowID)
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
		User:   profile.SSHOptions.User,
		Domain: profile.SSHOptions.Domain,
		Binary: executer.SSH,
		Args:   profile.SSHArgs(),
	}

	executerOptions.Hostname, err = utils.Select(svcProvider.Names())
	if err != nil {
		return err
	}
	return runCommands(executerOptions, flow)
}
func runCommands(opts *executer.Options, flow []config.FlowOptions) error {
	var (
		keys          map[string]string
		runCmd        string
		renderedTempl bytes.Buffer
	)

	for idx, cmd := range flow {

		runCmdTempl, err := template.New(fmt.Sprintf("cmd_%d", idx)).Parse(cmd.Run)
		if err != nil {
			return err
		}

		if keys != nil {
			fmt.Println(keys)
			runCmdTempl.Execute(&renderedTempl, keys)
			runCmd = renderedTempl.String()
		} else {
			runCmd = cmd.Run
		}

		opts.RemoteCmd = strings.Split(runCmd, " ")
		fmt.Println(opts.RemoteCmd)
		e, err := executer.NewSSHExecuter(opts)
		if err != nil {
			return err
		}

		switch cmd.OutputFormat {
		case "json":
			output := e.CommandWithOutput()
			if len(output.Stderr) != 0 {
				return fmt.Errorf(string(output.Stderr))
			}

			parsedOutput, err := utils.UnmarshalJSON(string(output.Stdout), cmd)
			if err != nil {
				return err
			}

			if cmd.Print {
				utils.PrintJSON(parsedOutput)
			}

			if idx < len(flow)-1 {
				selectorName, err := utils.Select(utils.GetSelector(parsedOutput, cmd))
				if err != nil {
					return err
				}

				keys, err = utils.SelectorKeys(parsedOutput, selectorName)
				if err != nil {
					return err
				}
			}
		default:
			if err := e.CommandWithTTY(); err != nil {
				return err
			}
		}
	}
	return nil
}
