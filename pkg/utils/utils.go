package utils

import (
	"fmt"
	"os"
	"sort"

	"github.com/adamkobi/xt/pkg/iostreams"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/olekukonko/tablewriter"
)

const keyNotFound = "not found"

//Select returns the user selected instance or default instance
func Select(io *iostreams.IOStreams, options []string, searchPattern string) (string, error) {
	cs := io.ColorScheme()
	switch len(options) {
	case 0:
		fmt.Fprintf(io.Out, fmt.Sprintf("%s No instances found matching %s\n", cs.WarningIcon(), cs.Bold(searchPattern)))
		var result bool
		if err := survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Connect to %s?", searchPattern),
			Default: true,
		}, &result); err != nil {
			return "", err
		}
		if !result {
			return "", fmt.Errorf("command cancelled")
		}
		return searchPattern, nil
	case 1:
		fmt.Fprintf(io.Out, fmt.Sprintf("found one host %s\n", cs.Bold(options[0])))
		return options[0], nil
	default:
		var result string
		sort.Strings(options)
		if err := survey.AskOne(&survey.Select{
			Message:  "Available Hosts:",
			Options:  options,
			PageSize: 15,
		}, &result); err != nil {
			return "", err
		}
		return result, nil
	}
}

//Table generates a table according to data provided
func Table(header []string, instances [][]string) {
	tbl := tablewriter.NewWriter(os.Stdout)
	tbl.SetHeader(header)

	for _, instance := range instances {
		tbl.Append(instance)
	}
	tbl.Render()
}
