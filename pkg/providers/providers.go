package providers

import (
	"fmt"
	"os"

	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/logging"
	"github.com/olekukonko/tablewriter"
)

//Provider is an interface describing actions in cloud provider
type Provider interface {
	FindByID(searchPattern string) ([]EC2Instance, error)
}

var logger = logging.GetLogger()

func getProviders() []Provider {
	var foundProviders = []Provider{}
	providers := config.GetProviders()
	for provider, cfg := range providers {
		switch provider {
		case "aws":
			foundProviders = append(foundProviders, NewAWSProvider(cfg))
		case "gcp":
			logger.Error("gcp not supported, continuing...")
		case "azure":
			logger.Error("azure not supported, continuing...")
		default:
			logger.Errorf("%s not supported, continuing...", provider)
		}
	}

	if len(foundProviders) == 0 {
		logger.Fatal("no providers found, unable to proceed.")
	}

	return foundProviders
}

//GetInstances returns slice of instance data per provider
func GetInstances(searchPattern string) ([]EC2Instance, error) {
	providers := getProviders()
	if len(providers) == 0 {
		return nil, fmt.Errorf("could not find supported provider, please verify you have this set in the config file")
	}

	var allInstances []EC2Instance
	for _, provider := range providers {
		instances, err := provider.FindByID(searchPattern)
		if err != nil {
			return nil, err
		}
		allInstances = append(allInstances, instances...)
	}
	return allInstances, nil
}

//GetIds returns all instances names
func GetIds(searchPattern string) ([]string, error) {
	instances, err := GetInstances(searchPattern)
	if err != nil {
		return nil, err
	}
	var instanceNames []string
	for _, instance := range instances {
		instanceNames = append(instanceNames, instance.Name)
	}

	return instanceNames, nil
}

//PrintInfo prints the fields requested by user as a table info of the instances
func PrintInfo(instances []EC2Instance) error {
	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"name", "id", "type", "image", "ip address",
		"availability zone", "subnet", "launch time", "lifecycle"}
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

	for _, inst := range instances {
		table.Append([]string{inst.Name, inst.InstanceID, inst.InstanceType, inst.ImageID, inst.PrivateIPAddress,
			inst.AvailabilityZone, inst.SubnetID, inst.LaunchTime, inst.InstanceLifecycle})
	}
	table.Render()
	return nil
}
