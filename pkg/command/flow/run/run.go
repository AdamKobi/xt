package run

import (
	"bytes"
	"fmt"
	"strings"

	"text/template"

	"github.com/MakeNowJust/heredoc"
	"github.com/adamkobi/xt/internal/config"
	"github.com/adamkobi/xt/internal/instance"
	"github.com/adamkobi/xt/pkg/cmdutil"
	"github.com/adamkobi/xt/pkg/executer"
	"github.com/adamkobi/xt/pkg/iostreams"
	"github.com/adamkobi/xt/pkg/provider"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type Options struct {
	Config        func() (*config.Config, error)
	IO            *iostreams.IOStreams
	SearchPattern string
	Profile       string
	Tag           string
	FlowID        string
}

func NewCmdRun(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Config: f.Config,
		IO:     f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "run <flow> <servers> [flags]",
		Short: "Execute multiple remote commands from config file",
		Long: heredoc.Doc(`
				Run a series of commands from config file.

				Use the output of the previous command as input to the current command.

				Manipulate JSON keys and interpolate them into commands.
		`),
		Example: heredoc.Doc(`
				$ xt flow run connect-pods web
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.FlowID = args[0]
			opts.SearchPattern = strings.TrimSuffix(args[1], "*")
			opts.Tag, _ = cmd.Flags().GetString("tag")
			opts.Profile, _ = cmd.Flags().GetString("profile")

			return runFlow(opts)
		},
	}

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
	return runCommands(cmdOpts, flow)
}

func runCommands(opts *executer.Options, flow []config.FlowOptions) error {
	var (
		selectedKeys  map[string]string
		runCmd        string
		renderedTempl bytes.Buffer
	)

	for idx, cmd := range flow {

		if selectedKeys != nil {
			runCmdTempl, err := template.New(fmt.Sprintf("cmd_%d", idx)).Parse(cmd.Run)
			if err != nil {
				return err
			}

			runCmdTempl.Execute(&renderedTempl, selectedKeys)
			runCmd = renderedTempl.String()
		} else {
			runCmd = cmd.Run
		}

		optsClone := *opts
		optsClone.RemoteCmd = strings.Split(runCmd, " ")
		e, err := executer.New(&optsClone)
		if err != nil {
			return err
		}

		switch cmd.OutputFormat {
		case "json":
			output, err := e.Output()
			if err != nil {
				return err
			}

			parsedOutput, err := parseJSONFromFlow(string(output), cmd)
			if err != nil {
				return err
			}

			if cmd.Print {
				printJSON(optsClone.IO, cmd.GetKeys(), parsedOutput)
			}

			if idx < len(flow)-1 {
				selectorName, err := utils.Select(opts.IO, getSelectors(parsedOutput, cmd), "")
				if err != nil {
					return err
				}

				selectedKeys = getDataFromSelected(parsedOutput, selectorName)

			}
		default:
			if err := e.Connect(); err != nil {
				return err
			}
		}
	}
	return nil
}

//parseJSONFromFlow fetches all selectors from input json and returns an json with selectors and values
func parseJSONFromFlow(json string, cmd config.FlowOptions) ([]map[string]string, error) {
	rootSlice := gjson.Get(json, cmd.Root)
	if !rootSlice.IsArray() {
		return nil, fmt.Errorf("%s is not a list, parse must be a list", cmd.Root)
	}

	var parsedJSON []map[string]string
	for _, item := range rootSlice.Array() {
		var parsedObject = make(map[string]string)
		for _, key := range cmd.Keys {
			result := item.Get(key.Path)
			if result.Exists() {
				parsedObject[key.Name] = result.String()
			} else {
				return nil, fmt.Errorf("key `%s` not found", key)
			}
		}
		parsedJSON = append(parsedJSON, parsedObject)
	}
	return parsedJSON, nil
}

//getSelectors returns a slice of the root map keys
func getSelectors(data []map[string]string, cmd config.FlowOptions) []string {
	var selectors []string
	for _, s := range data {
		if selector, ok := s[cmd.Selector]; ok {
			selectors = append(selectors, selector)
		}
	}
	return selectors
}

//selectorKeys returns a map of the selected selector
func getDataFromSelected(data []map[string]string, selected string) map[string]string {
	for _, item := range data {
		for _, v := range item {
			if v == selected {
				return item
			}
		}
	}
	return nil
}

//printJSON writes the fields requested by user to console in a formated table
func printJSON(io *iostreams.IOStreams, headers []string, data []map[string]string) {
	cs := io.ColorScheme()
	if len(data) == 0 {
		fmt.Fprintf(io.ErrOut, cs.Gray("no data received"))
	}

	table := utils.NewTablePrinter(io)

	for _, header := range headers {
		table.AddField(header, nil, cs.MagentaBold)
	}
	table.EndRow()

	for _, item := range data {
		for _, header := range headers {
			if val, ok := item[header]; ok {
				table.AddField(val, nil, cs.Green)
			}
		}
		table.EndRow()
	}

	_ = table.Render()
}
