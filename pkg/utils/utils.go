package utils

import (
	"fmt"
	"os"
	"sort"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/adamkobi/xt/config"
	log "github.com/adamkobi/xt/pkg/logging"
	"github.com/mgutz/ansi"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
)

//GetChoices will prompt user with server names found and require user to choose server
func getChoices(name, message string, names []string) string {
	var answers string
	sort.Strings(names)
	var qs = []*survey.Question{
		{
			Name: name,
			Prompt: &survey.Select{
				Message:  message,
				Options:  names,
				PageSize: 15,
			},
		},
	}
	err := survey.Ask(qs, &answers)
	if err != nil {
		log.Main.Fatal(err.Error())
	}
	return answers
}

//GetApproval requests user permission to run commands
func GetApproval(cmd string, instances []string) bool {
	approval := false
	keyColor := ansi.ColorFunc("red+h:black")
	valueColor := ansi.ColorFunc("green+h:black")
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("\n"+keyColor("Command: ")+"%s\n"+keyColor("Hosts/Pods:")+"\n%s\n"+keyColor("Are you sure you want to continue"), valueColor(cmd), valueColor(strings.Join(instances, "\n"))),
	}
	err := survey.AskOne(prompt, &approval, survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Text = keyColor("Will run the following:")
	}))
	if err != nil {
		return false
	}
	return approval
}

//SelectInstance returns the user selected instance or default instance
func SelectInstance(instances []string, searchPattern string) string {
	if len(instances) == 1 {
		log.Main.Info("Found 1 host: " + instances[0])
		return instances[0]
	} else if len(instances) > 1 {
		log.Main.Infof(fmt.Sprintf("Found %d hosts", len(instances)))
		choicesName := "SSHSeverList"
		choicesMessage := "Instances:"
		return getChoices(choicesName, choicesMessage, instances)
	} else {
		log.Main.Info("NO SERVERS FOUND, trying to connect to " + searchPattern)
		return searchPattern
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

func UnmarshalKeys(json string, dataStruct config.Command) (map[string]map[string]string, error) {
	instances := make(map[string]map[string]string)

	if !gjson.Get(json, dataStruct.Root).IsArray() {
		return nil, fmt.Errorf("%s is not an list", dataStruct.Root)
	}

	for _, object := range gjson.Get(json, dataStruct.Root).Array() {
		strObject := object.String()
		if !gjson.Get(strObject, dataStruct.Identifier).Exists() {
			return nil, fmt.Errorf("%s not in json, id is required", dataStruct.Identifier)
		}
		id := gjson.Get(strObject, dataStruct.Identifier).String()
		instances[id] = make(map[string]string)
		if dataStruct.Keys != nil {
			for _, key := range dataStruct.Keys {
				if gjson.Get(strObject, key).Exists() {
					instances[id][key] = gjson.Get(strObject, key).String()
				} else {
					instances[id][key] = "not found"
				}
			}
		}
	}
	return instances, nil
}

//GetIds returns a slice of ids
func GetIds(data map[string]map[string]string) []string {
	var ids []string
	for k := range data {
		ids = append(ids, k)
	}
	return ids
}

//PrintInfo prints the fields requested by user as a table info of the instances
func PrintInfo(keys []string, data map[string]map[string]string) error {
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
