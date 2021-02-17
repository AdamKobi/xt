package run

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"text/template"

	"github.com/MakeNowJust/heredoc"
	"github.com/adamkobi/xt/internal/config"
	"github.com/adamkobi/xt/pkg/cmdutil"
	"github.com/adamkobi/xt/pkg/executer"
	"github.com/adamkobi/xt/pkg/iostreams"
	"github.com/adamkobi/xt/pkg/provider"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/olekukonko/tablewriter"
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
		fmt.Fprintf(opts.IO.Out, cs.Yellow("%s"), profile.Message())
	}

	providerOpts := &provider.Options{
		Name:          profile.ProviderOptions.Name,
		VPC:           profile.ProviderOptions.VPC,
		Region:        profile.ProviderOptions.Region,
		CredsProfile:  profile.ProviderOptions.CredsProfile,
		Tag:           opts.Tag,
		SearchPattern: opts.SearchPattern,
	}

	svc, err := provider.New(providerOpts)
	if err != nil {
		return err
	}

	instances, err := svc.Get()
	if err != nil {
		return err
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
			runCmdTempl.Execute(&renderedTempl, keys)
			runCmd = renderedTempl.String()
		} else {
			runCmd = cmd.Run
		}

		opts.RemoteCmd = strings.Split(runCmd, " ")
		e, err := executer.New(opts)
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
				printJSON(parsedOutput)
			}

			if idx < len(flow)-1 {
				selectorName, err := utils.Select(opts.IO, getSelectors(parsedOutput, cmd), "")
				if err != nil {
					return err
				}

				keys, err = selectorKeys(parsedOutput, selectorName)
				if err != nil {
					return err
				}
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
	var parsedJSON []map[string]string
	rootSlice := gjson.Get(json, cmd.Root)
	if !rootSlice.IsArray() {
		return nil, fmt.Errorf("%s is not a list, parse must be a list", cmd.Root)
	}

	for _, item := range rootSlice.Array() {
		var parsedItem = make(map[string]string)
		result := item.Get(cmd.Selector)
		if result.Exists() {
			parsedItem[cmd.Selector] = result.String()
			parsedJSON = append(parsedJSON, parsedItem)
		} else {
			return nil, fmt.Errorf("selector `%s` not found", cmd.Selector)
		}

		for _, key := range cmd.Keys {
			result := item.Get(key)
			if result.Exists() {
				parsedItem[key] = result.String()
			} else {
				return nil, fmt.Errorf("key `%s` not found", key)
			}
		}
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
func selectorKeys(data []map[string]string, selector string) (map[string]string, error) {
	var wantedItem map[string]string
	normalizedKeys := make(map[string]string)
	for _, item := range data {
		for _, v := range item {
			if v == selector {
				wantedItem = item
				break
			}
		}
	}

	for k, v := range wantedItem {
		normalizedKeys[strings.ReplaceAll(k, ".", "_")] = v
	}

	if len(normalizedKeys) == 0 {
		return nil, fmt.Errorf("selector %s not found: lookup error", selector)
	}
	return normalizedKeys, nil
}

//printJSON writes the fields requested by user to console in a formated table
func printJSON(data []map[string]string) error {
	if data == nil {
		return fmt.Errorf("no data recevied, unable to proceed")
	}

	table := tablewriter.NewWriter(os.Stdout)
	var header []string

	for key := range data[0] {
		nameSlice := strings.Split(key, ".")
		normalizedName := nameSlice[len(nameSlice)-1]
		if validateHeader(header, normalizedName) {
			normalizedName = nameSlice[len(nameSlice)-2] + "." + normalizedName
		}
		header = append(header, normalizedName)
	}

	for _, item := range data {
		var line []string
		for _, v := range item {
			line = append(line, v)
		}
		table.Append(line)
	}

	table.SetHeader(header)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
	table.Render()
	return nil
}

func validateHeader(headers []string, testedHeader string) bool {
	for _, h := range headers {
		if h == testedHeader {
			return true
		}
	}
	return false
}
