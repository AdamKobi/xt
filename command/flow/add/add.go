package add

import (
	"fmt"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/adamkobi/xt/command/factory"
	"github.com/adamkobi/xt/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Config func() (*config.Config, error)
	Log    func() *logrus.Logger
	Debug  func()
	FlowID string
}

func NewCmdAdd(f *factory.CmdConfig) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		Debug:  f.Debug,
		Log:    f.Log,
	}

	cmd := &cobra.Command{
		Use:   "add <flow>",
		Short: "Add new flows",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.FlowID = args[0]
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				opts.Debug()
			}

			return runAdd(opts)
		},
	}

	return cmd
}

func runAdd(opts *Options) error {
	cfg, _ := opts.Config()
	var commands []config.FlowOptions

	fmt.Printf("Adding new flow %s\n", opts.FlowID)
	for {
		addMoreCommands := &survey.Confirm{
			Message: "Add additional commands?",
			Default: true,
		}

		runInput := &survey.Input{
			Message: "Please type command to run remotley",
			Help:    "Shell command that will be executed remotley during each step in Flow",
		}
		outputFormat := &survey.Select{
			Message: "Choose output format",
			Help:    "Text format will be printed without parsing. Json format can be parsed and used in future steps",
			Options: []string{"text", "json"},
			Default: "text",
		}

		selector := &survey.Input{
			Message: "JSON selector that will be used for the picker",
			Help:    "Selector will return a picker to select the correct data set to collect from.\nFor nested keys use dot notation",
		}

		key := &survey.Input{
			Message: "JSON key to parse from command output",
			Help:    "JSON key is collected from output and than can be used on next commands.\nFor nested keys use dot notation",
		}

		addKeys := &survey.Confirm{
			Message: "Add JSON keys to parse from output?",
			Default: false,
		}

		addMoreKeys := &survey.Confirm{
			Message: "Add additional keys to parse?",
			Default: true,
		}

		table := &survey.Confirm{
			Message: "Print parsed JSON as table before running next command",
			Help:    "Will parse both selector and keys and print them as table",
			Default: false,
		}

		var cmd config.FlowOptions

		if err := survey.AskOne(runInput, &cmd.Run); err != nil {
			return err
		}

		if err := survey.AskOne(outputFormat, &cmd.OutputFormat); err != nil {
			return err
		}

		if cmd.OutputFormat == "json" {
			if err := survey.AskOne(selector, &cmd.Selector); err != nil {
				return err
			}

			addKeysAnswer := false
			if err := survey.AskOne(addKeys, &addKeysAnswer); err != nil {
				return err
			}
			if addKeysAnswer {
				addMoreKeysAnswer := true
				for {
					if !addMoreKeysAnswer {
						break
					}
					var keyAnswer string
					if err := survey.AskOne(key, &keyAnswer); err != nil {
						return err
					}
					cmd.Keys = append(cmd.Keys, keyAnswer)

					if err := survey.AskOne(addMoreKeys, &addMoreKeysAnswer); err != nil {
						return err
					}
				}
			}

			if err := survey.AskOne(table, &cmd.Print); err != nil {
				return err
			}
		}
		commands = append(commands, cmd)
		var addMoreCommandsAnswer bool
		if err := survey.AskOne(addMoreCommands, &addMoreCommandsAnswer); err != nil {
			return err
		}

		if !addMoreCommandsAnswer {
			break
		}
	}

	d, err := yaml.Marshal(commands)
	if err != nil {
		return err
	}

	fmt.Print(string(d))

	confirmFlow := &survey.Confirm{
		Message: "Confirm writing following flow to configuration",
		Help:    "Will add new flow to config file",
		Default: false,
	}

	confirmFlowAnswer := false
	if err := survey.AskOne(confirmFlow, &confirmFlowAnswer); err != nil {
		return err
	}

	cfg.FlowOptions[opts.FlowID] = commands
	d, err = yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return config.WriteDefaultConfigFile(d)
}
