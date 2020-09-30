package config

import (
	"log"
	"sort"

	"github.com/adamkobi/xt/pkg/logging"
	"github.com/spf13/viper"
)

var (
	logger          = logging.GetLogger()
	cfg             Config
	configFilePaths = []string{"$HOME/.xt", "."}
	sshOptions      = []string{
		"-C", "-o", "LogLevel=ERROR", "-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null", "-o", "ControlPath=~/.ssh/cm-%C",
		"-o", "ControlMaster=auto", "-o", "ControlPersist=1m", "-at"}
	awsDefaultKeys = []string{
		"InstanceId", "ImageId", "PrivateIpAddress", "InstanceType",
		"Placement.AvailabilityZone", "InstanceLifecycle", "LaunchTime", "SubnetId",
	}
	providersDefault = Provider{
		SearchTags: SearchTags{
			Dynamic: "Name",
		}}
)

const (
	configName = "config"
	configType = "yaml"
)

type Config struct {
	Flows     map[string][]Command `mapstructure:"flows"`
	SSH       `mapstructure:"ssh"`
	Profiles  map[string]Profile  `mapstructure:"profiles"`
	Providers map[string]Provider `mapstructure:"providers"`
}

type Command struct {
	Run        string   `mapstructure:"run"`
	TTY        bool     `mapstructure:"tty,omitempty"`
	Identifier string   `mapstructure:"identifier,omitempty"`
	Root       string   `mapstructure:"root,omitempty"`
	Type       string   `mapstructure:"type,omitempty"`
	Keys       []string `mapstructure:"keys,omitempty"`
	Print      bool     `mapstructure:"print,omitempty"`
}

type Profile struct {
	Default   bool                `mapstructure:"default"`
	Providers map[string]Provider `mapstructure:"providers"`
	SSH       SSH                 `mapstructure:"ssh"`
}

type Provider struct {
	CredsProfile string     `mapstructure:"creds-profile"`
	Region       string     `mapstructure:"region"`
	VPC          string     `mapstructure:"vpc-id"`
	SearchTags   SearchTags `mapstructure:"filters"`
}

type SearchTags struct {
	Dynamic string            `mapstructure:"dynamic"`
	Static  map[string]string `mapstructure:"static"`
}
type SSH struct {
	User    string   `mapstructure:"user"`
	Domain  string   `mapstructure:"domain"`
	Options []string `mapstructure:"options,omitempty"`
}

func (c *Command) IsSet() bool {
	if c.Identifier == "" && c.Root == "" && len(c.Keys) == 0 {
		return false
	}
	return true
}

func InitConfig() {
	InitViper(viper.GetViper())
}

//GetProfiles return a slice of the profiles found in config file
func GetProfiles() []string {
	var profiles []string
	for p := range cfg.Profiles {
		profiles = append(profiles, p)
	}
	sort.Strings(profiles)
	return profiles
}

//DefaultProfile return the selected default profile from config file
func DefaultProfile() string {
	for p, opts := range cfg.Profiles {
		if opts.Default {
			return p
		}
	}
	return "no default profile set in config"
}

//GetFlows returns flows part of config
func GetFlows() map[string][]Command {
	return cfg.Flows
}

//GetProviders returns slice of all current found providers
func GetProviders() map[string]Provider {
	profile := GetProfile()
	return profile.Providers
}

//GetProfile return a subset of config file by profile selected in flags
func GetProfile() Profile {
	selectedProfile := viper.GetString("profile")
	profile := cfg.Profiles[selectedProfile]
	if profile.SSH.Options == nil {
		profile.SSH.Options = sshOptions
	}
	return profile
}

//InitViper will create a new initialized viper per cmd
func InitViper(v *viper.Viper) {
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	for _, path := range configFilePaths {
		v.AddConfigPath(path)
	}

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("unable to parse config, %v", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	setDefaults()
}

func setDefaults() {
	if cfg.SSH.Options == nil {
		cfg.SSH.Options = sshOptions
	}
}
