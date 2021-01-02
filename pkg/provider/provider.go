package provider

import (
	"fmt"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	awsProvider "github.com/adamkobi/xt/pkg/provider/aws"
)

var supportedProviders = [1]string{"aws"}

//ProviderOptions is all the options AWSProvider can receive
type Options struct {
	Name            string
	VPC             string
	Region          string
	CredsProfile    string
	AccessKeyID     string
	SecretAccessKey string
	Tag             string
	SearchPattern   string
	Filters         map[string]string
}

//Provider is an interface describing actions in cloud provider
type Provider interface {
	Instances() error
	Names() []string
	Print() error
}

type Instancer interface {
	Name() string
}

func New(o *Options) (Provider, error) {
	switch o.Name {
	case "aws":
		opts := awsProvider.Options{
			VPC:             o.VPC,
			Region:          o.Region,
			CredsProfile:    o.CredsProfile,
			AccessKeyID:     o.AccessKeyID,
			SecretAccessKey: o.SecretAccessKey,
			Tag:             o.Tag,
			SearchPattern:   o.SearchPattern,
			Filters:         o.Filters,
		}
		return awsProvider.NewAWSProvider(opts), nil
	default:
		return nil, fmt.Errorf("provider %s not supported", o.Name)
	}
}

//GetInstances returns slice of instance data per provider
func GetInstancesByProvider(providers []Provider, s, t string) error {
	for _, provider := range providers {
		if err := provider.Instances(); err != nil {
			return err
		}
	}
	return nil
}

//SelectInstance returns the user selected instance or default instance
func SelectHost(hosts []string, searchPattern string) (string, error) {
	if len(hosts) == 0 {
		return searchPattern, nil
	}
	if len(hosts) == 1 {
		return hosts[0], nil
	}
	msg := "Hosts:"
	return GetChoices(hosts, msg)
}

//GetChoices will prompt user with server names found and require user to choose server
func GetChoices(hosts []string, msg string) (string, error) {
	var answers string
	sort.Strings(hosts)
	var qs = []*survey.Question{
		{
			Prompt: &survey.Select{
				Message:  msg,
				Options:  hosts,
				PageSize: 15,
			},
		},
	}
	err := survey.Ask(qs, &answers)
	if err != nil {
		return "", err
	}
	return answers, nil
}
