package aws

import (
	"os/exec"

	"github.com/adamkobi/xt/internal/instance"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//Provider describes AWS configs
type Provider struct {
	Client  *ec2.EC2
	Options Options
}

//Options is all the options AWSProvider can receive
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

//ErrorNotFound is returned when no key exists for equivelent in EC2Instance struct
const ErrorNotFound = "not found"

//New returns AWS provider configs
func New(opts *Options) (*Provider, error) {
	client, err := newEC2(opts)
	if err != nil {
		return nil, err
	}

	return &Provider{
		Client: client,
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
	}, nil
}

func newEC2(opts *Options) (*ec2.EC2, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           opts.CredsProfile,
	})

	if err != nil {
		return nil, err
	}

	if _, err := sess.Config.Credentials.Get(); err != nil {
		if err := ssoLogin(opts); err != nil {
			return nil, err
		}
	}

	return ec2.New(sess), nil
}

func ssoLogin(opts *Options) error {
	binary := "aws"
	if _, err := exec.LookPath(binary); err != nil {
		return err
	}

	args := []string{"sso", "login", "--profile", opts.CredsProfile}
	cmd := exec.Command(binary, args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

//Get will filter all instances according to tag
func (p *Provider) Get() (instance.XTInstances, error) {
	filters := []*ec2.Filter{
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
		tag := "tag:" + tagKey
		filters = append(filters, &ec2.Filter{
			Name:   aws.String(tag),
			Values: []*string{aws.String(tagValue)},
		})
	}

	params := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	res, err := p.Client.DescribeInstances(params)
	if err != nil {
		return nil, err
	}

	return parseOutput(res, p.Options.Tag), nil
}

func parseOutput(ec2 *ec2.DescribeInstancesOutput, searchTag string) instance.XTInstances {
	var instances instance.XTInstances
	for idx := range ec2.Reservations {
		for _, inst := range ec2.Reservations[idx].Instances {
			var name string
			for _, tag := range inst.Tags {
				if *tag.Key == searchTag {
					name = *tag.Value
				}
			}
			instance := instance.XTInstance{
				InstanceName:      name,
				InstanceID:        getValue(inst.InstanceId),
				ImageID:           getValue(inst.ImageId),
				InstanceType:      getValue(inst.InstanceType),
				PrivateIPAddress:  getValue(inst.PrivateIpAddress),
				PublicIPAddress:   getValue(inst.PublicIpAddress),
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
