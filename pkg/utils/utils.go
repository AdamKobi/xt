package utils

import (
	"fmt"
	"os"
	"sort"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/adamkobi/xt/config"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
)

const keyNotFound = "not found"

//GetApproval requests user permission to run commands
func ApproveRun(cmd string, options []string) bool {
	approval := false
	msg := strings.Builder{}
	msg.WriteString("Execute Command\n")
	msg.WriteString(fmt.Sprintf("Command: %s\n\n", cmd))
	msg.WriteString(fmt.Sprintf("Hosts:\n%s\n\n", strings.Join(options, "\n")))
	msg.WriteString("Are you sure you want to continue")
	prompt := &survey.Confirm{
		Message: msg.String(),
	}
	err := survey.AskOne(prompt, &approval)
	if err != nil {
		return false
	}
	return approval
}

//Select returns the user selected instance or default instance
func Select(options []string) (string, error) {
	if len(options) == 1 {
		return options[0], nil
	}
	msg := "Hosts:"
	return GetChoices(options, msg)
}

//GetChoices will prompt user with server names found and require user to choose server
func GetChoices(options []string, msg string) (string, error) {
	var answers string
	sort.Strings(options)
	var prompt = &survey.Select{
		Message:  msg,
		Options:  options,
		PageSize: 15,
	}
	err := survey.AskOne(prompt, &answers)
	if err != nil {
		return "", err
	}
	return answers, nil
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

//UnmarshalJSON fetches all selectors from input json and returns an json with selectors and values
func UnmarshalJSON(json string, cmd config.FlowOptions) ([]map[string]string, error) {
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

//GetSelector returns a slice of the root map keys
func GetSelector(data []map[string]string, cmd config.FlowOptions) []string {
	var selectors []string
	for _, s := range data {
		if selector, ok := s[cmd.Selector]; ok {
			selectors = append(selectors, selector)
		}
	}
	return selectors
}

//SelectorKeys returns a map of the selected selector
func SelectorKeys(data []map[string]string, selector string) (map[string]string, error) {
	normalizedKeys := make(map[string]string)
	for _, item := range data {
		for k, v := range item {
			if v == selector {
				normalizedKeys[strings.ReplaceAll(k, ".", "_")] = v
			}
		}
	}
	if len(normalizedKeys) == 0 {
		return nil, fmt.Errorf("selector %s not found: lookup error", selector)
	}
	return normalizedKeys, nil
}

func validateHeader(headers []string, testedHeader string) bool {
	for _, h := range headers {
		if h == testedHeader {
			return true
		}
	}
	return false
}

//PrintJSON writes the fields requested by user to console in a formated table
func PrintJSON(data []map[string]string) error {
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
