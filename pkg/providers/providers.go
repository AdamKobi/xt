package providers

import (
	"fmt"

	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/utils"
	"github.com/aws/aws-sdk-go/aws/session"
)

//Provider is an interface describing actions in cloud provider
type Provider interface {
	FindByID(tag, id string) (map[string]map[string]string, error)
}

//AWSProvider describes AWS configs
type AWSProvider struct {
	VPC             string
	Region          string
	CredsProfile    string
	AccessKeyID     string
	SecretAccessKey string
	Session         *session.Session
}

//GetProviders returns slice of all current found providers
func getProviders() []Provider {
	profile := config.GetProfile()
	var providers []Provider
	if profile.IsSet("providers.aws") {
		providers = append(providers, NewAWSProvider())
	}
	// if profile.IsSet("providers.gcp") {
	// 	return "gcp"
	// }
	// if profile.IsSet("providers.azure") {
	// 	return "azure"
	// }
	return providers
}

//GetInstances returns slice of instance data per provider
func GetInstances(tag, searchPattern string) ([]map[string]map[string]string, error) {
	providers := getProviders()
	if len(providers) == 0 {
		return nil, fmt.Errorf("could not find supported provider, please verify you have this set in the config file")
	}

	var allInstances []map[string]map[string]string
	for _, provider := range providers {
		instances, err := provider.FindByID(tag, searchPattern)
		if err != nil {
			return nil, err
		}
		allInstances = append(allInstances, instances)
	}
	return allInstances, nil
}

func GetIds(tag, searchPattern string) ([]string, error) {
	instancesByProvider, err := GetInstances(tag, searchPattern)
	if err != nil {
		return nil, err
	}

	var instanceIds []string
	for _, instanceGroup := range instancesByProvider {
		instanceIds = append(instanceIds, utils.GetIds(instanceGroup)...)
	}
	return instanceIds, nil
}
