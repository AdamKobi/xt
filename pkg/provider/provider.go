package provider

import (
	"fmt"

	"github.com/adamkobi/xt/internal/instance"
	awsProvider "github.com/adamkobi/xt/pkg/provider/aws"
)

var supportedProviders = [1]string{"aws"}

//Options is all the options AWSProvider can receive
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
	Get() (instance.XTInstances, error)
}

//New creates a new provider according to the provider type
func New(options *Options) (Provider, error) {
	switch options.Name {
	case "aws":
		opts := &awsProvider.Options{
			VPC:             options.VPC,
			Region:          options.Region,
			CredsProfile:    options.CredsProfile,
			AccessKeyID:     options.AccessKeyID,
			SecretAccessKey: options.SecretAccessKey,
			Tag:             options.Tag,
			SearchPattern:   options.SearchPattern,
			Filters:         options.Filters,
		}
		p, err := awsProvider.New(opts)
		if err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, fmt.Errorf("provider %s not supported", options.Name)
	}
}
