package utils

import (
	"fmt"
	"os"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/adamkobi/xt/config"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
)

const keyNotFound = "not found"

//GetApproval requests user permission to run commands
func GetApproval(cmd string, instances []string) bool {
	approval := false
	msg := fmt.Sprintf("Execute Command\nCommand:%s\n\nHosts:\n%s\n\nAre you sure you want to continue", cmd, strings.Join(instances, "\n"))
	prompt := &survey.Confirm{
		Message: msg,
	}
	err := survey.AskOne(prompt, &approval, nil)
	if err != nil {
		return false
	}
	return approval
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

//UnmarshalKeys
func UnmarshalKeys(json string, flow config.FlowOptions) (map[string]map[string]string, error) {
	instances := make(map[string]map[string]string)

	if !gjson.Get(json, flow.Parse).IsArray() {
		return nil, fmt.Errorf("%s is not a list, parse must be a list", flow.Parse)
	}

	for _, item := range gjson.Get(json, flow.Parse).Array() {
		selector := gjson.Get(item.Raw, flow.Selector).String()
		instances[selector] = make(map[string]string)
		if flow.Keys != nil {
			for _, key := range flow.Keys {
				if gjson.Get(item.Raw, key).Exists() {
					instances[selector][key] = gjson.Get(item.Raw, key).String()
				} else {
					instances[selector][key] = keyNotFound
				}
			}
		}
	}
	return instances, nil
}

//GetSelectors returns a slice of the root map keys
func GetSelectors(data map[string]map[string]string) []string {
	var selectors []string
	for s := range data {
		selectors = append(selectors, s)
	}
	return selectors
}

//PrintJSON writes the fields requested by user to console in a formated table
func PrintJSON(keys []string, data map[string]map[string]string) error {
	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"name"}
	if len(keys) == 0 {
		return fmt.Errorf("no keys provided, unable to proceed")
	}
	header = append(header, keys...)
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

	for name, attrib := range data {
		instance := []string{name}
		for _, key := range keys {
			if val, ok := attrib[key]; ok {
				instance = append(instance, val)
			} else {
				instance = append(instance, "undefined")
			}
		}
		table.Append(instance)
	}
	table.Render()
	return nil
}
