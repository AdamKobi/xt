package providers

import (
	"github.com/adamkobi/xt/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var defaultKeys = []string{
	"instanceId", "image", "type", "lifecycle", "arn", "privateIpAddress",
	"key", "launchTime", "state", "availabilityZone", "privateDNS", "subnet", "vpc",
}

//NewAWSProvider returns AWS provider configs
func NewAWSProvider() *AWSProvider {
	profile := config.GetProfile()
	return &AWSProvider{
		VPC:          profile.GetString("providers.aws.vpc-id"),
		Region:       profile.GetString("providers.aws.region"),
		CredsProfile: profile.GetString("providers.aws.creds-profile"),
	}
}

func (p *AWSProvider) session() *AWSProvider {
	p.Session = session.Must(session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: &p.Region},
		Profile: p.CredsProfile,
	}))
	return p
}

func (p *AWSProvider) ec2() *ec2.EC2 {
	return ec2.New(p.Session)
}

//FindByID will filter all instances according to tag and id
func (p *AWSProvider) FindByID(tag, id string) (map[string]map[string]string, error) {
	ec2Service := p.session().ec2()
	ec2Tag := "tag:" + tag
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(p.VPC)},
			},
			{
				Name:   aws.String(ec2Tag),
				Values: []*string{aws.String(id)},
			},
		},
	}
	res, err := ec2Service.DescribeInstances(params)
	if err != nil {
		return nil, err
	}
	return unmarshal(res), nil
}

func unmarshal(dio *ec2.DescribeInstancesOutput) map[string]map[string]string {
	instances := make(map[string]map[string]string)
	for idx := range dio.Reservations {
		for _, inst := range dio.Reservations[idx].Instances {
			name := "undefined"
			for _, tag := range inst.Tags {
				if *tag.Key == "Name" {
					name = *tag.Value
				}
			}

			instances[name] = map[string]string{
				"arch":             *inst.Architecture,
				"hypervisor":       *inst.Hypervisor,
				"arn":              *inst.IamInstanceProfile.Arn,
				"image":            *inst.ImageId,
				"instanceId":       *inst.InstanceId,
				"lifecycle":        *inst.InstanceLifecycle,
				"type":             *inst.InstanceType,
				"key":              *inst.KeyName,
				"launchTime":       inst.LaunchTime.String(),
				"state":            *inst.Monitoring.State,
				"availabilityZone": *inst.Placement.AvailabilityZone,
				"privateDNS":       *inst.PrivateDnsName,
				"privateIpAddress": *inst.PrivateIpAddress,
				"publicDNS":        *inst.PublicDnsName,
				"subnet":           *inst.SubnetId,
				"virtualization":   *inst.VirtualizationType,
				"vpc":              *inst.VpcId,
			}
		}
	}
	return instances
}
