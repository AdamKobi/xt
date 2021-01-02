package providers

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/olekukonko/tablewriter"
)

//AWSProvider describes AWS configs
type AWSProvider struct {
	Options Options
	Client  *ec2.EC2
	Hosts   []EC2Instance
}

//AwsOptions is all the options AWSProvider can receive
type Options struct {
	VPC             string
	Region          string
	CredsProfile    string
	AccessKeyID     string
	SecretAccessKey string
	Tag             string
	SearchPattern   string
	Filters         map[string]string
}

//EC2Instance describes all the fields that Xt collects from EC2 instances
type EC2Instance struct {
	Name, InstanceID, ImageID, PrivateIPAddress, InstanceType,
	AvailabilityZone, InstanceLifecycle, LaunchTime, SubnetID string
}

//ErrorNotFound is returned when no key exists for equivelent in EC2Instance struct
const ErrorNotFound = "not found"

//NewAWSProvider returns AWS provider configs
func NewAWSProvider(opts Options) *AWSProvider {
	return &AWSProvider{
		Options: Options{
			VPC:             opts.VPC,
			Region:          opts.Region,
			CredsProfile:    opts.CredsProfile,
			AccessKeyID:     opts.AccessKeyID,
			SecretAccessKey: opts.SecretAccessKey,
			SearchPattern:   opts.SearchPattern,
			Tag:             opts.Tag,
			Filters:         opts.Filters,
		},
	}
}

func (p *AWSProvider) ec2() {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: &p.Options.Region},
		Profile: p.Options.CredsProfile,
	}))
	p.Client = ec2.New(session)
}

//Instances will filter all instances according to tag
func (p *AWSProvider) Instances() error {
	p.ec2()
	ec2Filters := []*ec2.Filter{
		{
			Name:   aws.String("vpc-id"),
			Values: []*string{aws.String(p.Options.VPC)},
		},
		{
			Name:   aws.String("tag:" + p.Options.Tag),
			Values: []*string{aws.String(p.Options.SearchPattern + "*")},
		},
	}
	for tagKey, tagValue := range p.Options.Filters {
		ec2Tag := "tag:" + tagKey
		ec2Filters = append(ec2Filters, &ec2.Filter{
			Name:   aws.String(ec2Tag),
			Values: []*string{aws.String(tagValue)},
		})
	}

	params := &ec2.DescribeInstancesInput{
		Filters: ec2Filters,
	}

	res, err := p.Client.DescribeInstances(params)
	if err != nil {
		return err
	}
	p.parseInstancesOutput(res)
	return nil
}

//Names returns all instances names
func (p *AWSProvider) Names() []string {
	var instanceNames []string
	for _, instance := range p.Hosts {
		instanceNames = append(instanceNames, instance.Name)
	}
	return instanceNames
}

func (p *AWSProvider) parseInstancesOutput(dio *ec2.DescribeInstancesOutput) {
	for idx := range dio.Reservations {
		for _, inst := range dio.Reservations[idx].Instances {
			name := ""
			for _, tag := range inst.Tags {
				if *tag.Key == "Name" {
					name = *tag.Value
				}
			}
			instance := EC2Instance{
				Name:              name,
				InstanceID:        getValue(inst.InstanceId),
				ImageID:           getValue(inst.ImageId),
				InstanceType:      getValue(inst.InstanceType),
				PrivateIPAddress:  getValue(inst.PrivateIpAddress),
				SubnetID:          getValue(inst.SubnetId),
				AvailabilityZone:  getValue(inst.Placement.AvailabilityZone),
				InstanceLifecycle: getValue(inst.InstanceLifecycle),
				LaunchTime:        inst.LaunchTime.String(),
			}
			p.Hosts = append(p.Hosts, instance)
		}
	}
}

func getValue(val *string) string {
	if val != nil {
		return *val
	}
	return ErrorNotFound
}

//Print the fields requested by user as a table info of the instances
func (p *AWSProvider) Print() error {
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

	for _, inst := range p.Hosts {
		table.Append([]string{inst.Name, inst.InstanceID, inst.InstanceType, inst.ImageID, inst.PrivateIPAddress,
			inst.AvailabilityZone, inst.SubnetID, inst.LaunchTime, inst.InstanceLifecycle})
	}
	table.Render()
	return nil
}
