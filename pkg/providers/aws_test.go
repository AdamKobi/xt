package providers

import (
	"testing"

	"github.com/adamkobi/xt/config"
)

var cfg = config.Config{
	Profiles: map[string]config.Profile{
		"test": config.Profile{
			Providers: map[string]config.Provider{
				"aws": config.Provider{
					CredsProfile: "test-creds-profile",
					Region:       "eu-west-1",
					VPC:          "vpc-test",
					SearchTags: config.SearchTags{
						Dynamic: "testTag",
						Static: map[string]string{
							"constTag": "constValue",
						},
					},
				},
				"gcp": config.Provider{
					CredsProfile: "test-creds-profile",
					Region:       "eu-west-1",
					VPC:          "vpc-test",
					SearchTags: config.SearchTags{
						Dynamic: "testTag",
						Static: map[string]string{
							"constTag": "constValue",
						},
					},
				},
			},
		},
	},
}

func TestGetProviders(t *testing.T) {
	config.Config
	getProviders()

}
