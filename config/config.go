package config

import (
	"fmt"
	"sort"

	"github.com/adamkobi/xt/pkg/logging"
	"github.com/spf13/viper"
)

var configFilePaths = []string{"$HOME/.xt", "."}
var XT Config

const (
	configName = "config"
	configType = "yaml"
)

type Config struct {
	Flows    map[string][]Command `mapstructure:"flows"`
	SSH      SSHGlobal            `mapstructure:"ssh"`
	Profiles map[string]Profile   `mapstructure:"profiles"`
}

type Data struct {
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

type SSHGlobal struct {
	Args []string `mapstructure:"args,omitempty"`
}

type Profile struct {
	Default   bool                `mapstructure:"default"`
	Providers map[string]Provider `mapstructure:"providers"`
	SSH       SSHPrivate          `mapstructure:"ssh"`
}

type Provider struct {
	CredsProfile string `mapstructure:"creds-profile"`
	Region       string `mapstructure:"region"`
	VPC          string `mapstructure:"vpc-id"`
}

type SSHPrivate struct {
	User   string `mapstructure:"user"`
	Domain string `mapstructure:"domain"`
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
	for p := range viper.GetStringMapString("profiles") {
		profiles = append(profiles, p)
	}
	sort.Strings(profiles)
	return profiles
}

//DefaultProfile return the selected default profile from config file
func DefaultProfile() string {
	for _, p := range GetProfiles() {
		if viper.GetBool(fmt.Sprintf("profiles.%s.default", p)) {
			return p
		}
	}
	return "undefined"
}

//GetProfile return a subset of config file by profile selected in flags
func GetProfile() *viper.Viper {
	return viper.Sub(fmt.Sprintf("profiles.%s", viper.GetString("profile")))
}

func InitViper(v *viper.Viper) {
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	for _, path := range configFilePaths {
		v.AddConfigPath(path)
	}
	if err := v.ReadInConfig(); err == nil {
		if v == viper.GetViper() {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
	err := v.Unmarshal(&XT)
	if err != nil {
		logging.Main.Fatalf("unable to decode into struct, %v", err)
	}
	v.SetDefault("ssh.args", []string{
		"-C", "-o", "LogLevel=ERROR", "-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null"})
	v.SetDefault("commands.info.keys", []string{
		"instanceId", "image", "type", "lifecycle", "arn", "privateIpAddress",
		"key", "launchTime", "state", "availabilityZone", "privateDNS", "subnet", "vpc",
	})
}
