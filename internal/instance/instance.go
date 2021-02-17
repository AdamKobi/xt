package instance

import (
	"sort"

	"github.com/adamkobi/xt/pkg/iostreams"
	"github.com/adamkobi/xt/pkg/utils"
)

//XTInstance describes all the fields that Xt collects from EC2 instances
type XTInstance struct {
	InstanceName      string
	InstanceID        string
	ImageID           string
	PrivateIPAddress  string
	PublicIPAddress   string
	InstanceType      string
	AvailabilityZone  string
	InstanceLifecycle string
	LaunchTime        string
	SubnetID          string
}

func (i *XTInstance) name() string {
	return i.InstanceName
}

//XTInstances Describes a slice of XTInstance
type XTInstances []XTInstance

//Names returns a slice of the instances search tag value
func (i *XTInstances) Names() []string {
	var instances []string
	for _, instance := range *i {
		instances = append(instances, instance.name())
	}
	sort.Strings(instances)
	return instances
}

func (i *XTInstances) Print(io *iostreams.IOStreams) {
	cs := io.ColorScheme()
	table := utils.NewTablePrinter(io)
	headerFields := []string{"Instance Name", "Instance ID", "Type", "Image ID", "Private IP Address",
		"Public IP Address", "Availability Zone", "Subnet", "Launch Time", "Lifecycle"}
	for _, header := range headerFields {
		table.AddField(header, nil, cs.MagentaBold)
	}
	table.EndRow()
	for _, inst := range *i {
		table.AddField(inst.InstanceName, nil, cs.Green)
		table.AddField(inst.InstanceID, nil, cs.Green)
		table.AddField(inst.InstanceType, nil, cs.Green)
		table.AddField(inst.ImageID, nil, cs.Green)
		table.AddField(inst.PrivateIPAddress, nil, cs.Green)
		table.AddField(inst.PublicIPAddress, nil, cs.Green)
		table.AddField(inst.AvailabilityZone, nil, cs.Green)
		table.AddField(inst.SubnetID, nil, cs.Green)
		table.AddField(inst.LaunchTime, nil, cs.Green)
		table.AddField(inst.InstanceLifecycle, nil, cs.Green)
		table.EndRow()
	}
	_ = table.Render()
}
