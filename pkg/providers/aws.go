package providers

import (
	"github.com/adamkobi/xt/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//AWSProvider describes AWS configs
type AWSProvider struct {
	VPC, Region, CredsProfile,
	AccessKeyID, SecretAccessKey string
	Session    *session.Session
	SearchTags config.SearchTags
}

//EC2Instance describes all the fields that Xt collects from EC2 instances
type EC2Instance struct {
	Name, InstanceID, ImageID, PrivateIPAddress, InstanceType,
	AvailabilityZone, InstanceLifecycle, LaunchTime, SubnetID string
}

//ErrorNotFound is returned when no key exists for equivelent in EC2Instance struct
const ErrorNotFound = "undefined"

//NewAWSProvider returns AWS provider configs
func NewAWSProvider(cfg config.Provider) *AWSProvider {
	if cfg.SearchTags.Dynamic == "" {
		cfg.SearchTags.Dynamic = "Name"
	}
	return &AWSProvider{
		VPC:          cfg.VPC,
		Region:       cfg.Region,
		CredsProfile: cfg.CredsProfile,
		SearchTags:   cfg.SearchTags,
	}
}

func NewEC2(p *AWSProvider) *ec2.EC2 {
	p.Session = session.Must(session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: &p.Region},
		Profile: p.CredsProfile,
	}))
	return ec2.New(p.Session)
}

//FindByID will filter all instances according to tag and id
func (p *AWSProvider) FindByID(searchPattern string) ([]EC2Instance, error) {
	// c := config.GetProfile()
	ec2Service := NewEC2(p)
	ec2Filters := []*ec2.Filter{
		{
			Name:   aws.String("vpc-id"),
			Values: []*string{aws.String(p.VPC)},
		},
		{
			//Get tag dynamically from flag or config
			Name:   aws.String("tag:" + p.SearchTags.Dynamic),
			Values: []*string{aws.String(searchPattern + "*")},
		},
	}
	for tagKey, tagValue := range p.SearchTags.Static {
		ec2Tag := "tag:" + tagKey
		ec2Filters = append(ec2Filters, &ec2.Filter{
			Name:   aws.String(ec2Tag),
			Values: []*string{aws.String(tagValue)},
		})
	}

	params := &ec2.DescribeInstancesInput{
		Filters: ec2Filters,
	}

	res, err := ec2Service.DescribeInstances(params)
	if err != nil {
		return nil, err
	}

	return unmarshal(res), nil
}

func unmarshal(dio *ec2.DescribeInstancesOutput) []EC2Instance {
	var instances []EC2Instance

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
			instances = append(instances, instance)
		}
	}
	return instances
}

func getValue(val *string) string {
	if val != nil {
		return *val
	}
	return ErrorNotFound
}
