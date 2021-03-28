package config

import (
	"fmt"
	"sort"
	"strings"
)

const JSON = "json"
const notSetError = "%s must be set"

type Config struct {
	FlowOptions    map[string][]FlowOptions  `yaml:"flows"`
	ProfileOptions map[string]ProfileOptions `yaml:"profiles"`
	SSHOptions     SSHOptions                `yaml:"ssh"`
}

type FlowOptions struct {
	Run          string `yaml:"run"`
	Selector     string `yaml:"selector"`
	Keys         []Pair `yaml:"keys,omitempty"`
	Root         string `yaml:"root,omitempty"`
	OutputFormat string `yaml:"output_format"`
	Print        bool   `yaml:"print,omitempty"`
}

type Pair struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type ProfileOptions struct {
	Default         bool            `yaml:"default"`
	ProviderOptions ProviderOptions `yaml:"provider"`
	SSHOptions      SSHOptions      `yaml:"ssh"`
	DisplayMsg      string          `yaml:"message,omitempty"`
}

type ProviderOptions struct {
	Name         string            `yaml:"name"`
	CredsProfile string            `yaml:"creds-profile"`
	Region       string            `yaml:"region"`
	VPC          string            `yaml:"vpc-id"`
	Filters      map[string]string `yaml:"filters,omitempty"`
}

type SSHOptions struct {
	User   string   `yaml:"user"`
	Domain string   `yaml:"domain"`
	Args   []string `yaml:"options"`
}

//SSHArgs returns ssh options if they exist in profile else returns default
func (p *ProfileOptions) SSHArgs() []string {
	sshOptions := p.SSHOptions.Args
	if len(sshOptions) != 0 {
		return sshOptions
	}
	return defaultSSHOptions()
}

//SCPArgs returns ssh options if they exist in profile else returns default
func (p *ProfileOptions) SCPArgs() []string {
	sshOptions := p.SSHOptions.Args
	if len(sshOptions) != 0 {
		return sshOptions
	}
	return defaultSCPOptions()
}

func defaultSSHOptions() []string {
	return []string{
		"-Ct", "-o", "LogLevel=INFO", "-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null", "-o", "ControlPath=~/.ssh/cm-%C",
		"-o", "ControlMaster=auto", "-o", "ControlPersist=5m"}
}

func defaultSCPOptions() []string {
	return []string{
		"-C", "-o", "LogLevel=INFO", "-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null", "-o", "ControlPath=~/.ssh/cm-%C",
		"-o", "ControlMaster=auto", "-o", "ControlPersist=5m"}
}

//Profiles return a slice of the profiles found in config file
func (c *Config) Profiles() []string {
	var profiles []string
	for p := range c.ProfileOptions {
		profiles = append(profiles, p)
	}
	sort.Strings(profiles)
	return profiles
}

//DefaultProfile return the selected default profile from config file
func (c *Config) DefaultProfile() string {
	for p, opts := range c.ProfileOptions {
		if opts.Default {
			return p
		}
	}
	return "no default profile set in config"
}

//Flows returns flows part of config
func (c *Config) Flows() map[string][]FlowOptions {
	return c.FlowOptions
}

//Flow returns a slice of the selected flow
func (c *Config) Flow(flowID string) ([]FlowOptions, error) {
	flows := c.Flows()
	if p, ok := flows[flowID]; ok {
		for _, cmd := range p {
			if err := cmd.Validate(); err != nil {
				return nil, err
			}
		}
		return flows[flowID], nil
	}
	return nil, fmt.Errorf("flow %s not found", flowID)
}

//Profile return a subset of config file by profile selected in flags
func (c *Config) Profile(profileID string) (*ProfileOptions, error) {
	if p, ok := c.ProfileOptions[profileID]; ok {
		if err := p.Validate(); err != nil {
			return nil, err
		}
		return &p, nil
	}
	return nil, fmt.Errorf("%s profile not found in config file", profileID)
}

func (p *ProfileOptions) Validate() error {
	if err := p.SSHOptions.validate(); err != nil {
		return err
	}
	if err := p.ProviderOptions.validate(); err != nil {
		return err
	}
	return nil
}

func (p *ProfileOptions) Message() string {
	if p.DisplayMsg != "" {
		msg := strings.Builder{}
		msgSlice := strings.Split(p.DisplayMsg, "\n")
		maxLength := len(msgSlice[0])
		for _, msgPart := range msgSlice[1:] {
			if len(msgPart) > maxLength {
				maxLength = len(msgPart)
			}
		}
		border := strings.Repeat("-", maxLength+4)
		msg.WriteString("\n" + border + "\n")
		for _, msgPart := range msgSlice {
			msgPartLen := len(msgPart)
			spacesRequired := maxLength - msgPartLen + 1
			formatedMsgPart := fmt.Sprintf("| %s%s|\n", msgPart, strings.Repeat(" ", spacesRequired))
			msg.WriteString(formatedMsgPart)
		}
		msg.WriteString(border + "\n\n")
		return msg.String()
	}
	return ""
}

func (s *SSHOptions) validate() error {
	if s.User == "" {
		return fmt.Errorf(fmt.Sprintf(notSetError, "profile.ssh.user"))
	}
	if s.Domain == "" {
		return fmt.Errorf(fmt.Sprintf(notSetError, "profile.ssh.domain"))
	}
	return nil
}

func (p *ProviderOptions) validate() error {
	if p.Name == "" {
		return fmt.Errorf(fmt.Sprintf(notSetError, "profile.provider.name"))
	}
	if p.Region == "" {
		return fmt.Errorf(fmt.Sprintf(notSetError, "profile.provider.region"))
	}
	if p.CredsProfile == "" {
		return fmt.Errorf(fmt.Sprintf(notSetError, "profile.provider.creds-profile"))
	}
	if p.VPC == "" {
		return fmt.Errorf(fmt.Sprintf(notSetError, "profile.provider.vpc-id"))
	}
	return nil
}

//Provider returns the selected profile config
func (p *ProfileOptions) Provider() ProviderOptions {
	return p.ProviderOptions
}

func (f *FlowOptions) Validate() error {
	if f.Run == "" {
		return fmt.Errorf("run is required for running a flow")
	}
	switch f.OutputFormat {
	case JSON:
		if f.Root == "" {
			return fmt.Errorf("parse must be set when using json type")
		}
		if len(f.Keys) < 1 {
			return fmt.Errorf("keys must be set when using json type")
		}
		if f.Selector != "" {
			valid := false
			for _, key := range f.Keys {
				if f.Selector == key.Name {
					valid = true
				}
			}
			if !valid {
				return fmt.Errorf("selector must equal to one of keys provided")
			}
		}
		return nil
	default:
		return nil
	}
}

func (f *FlowOptions) GetSelector() *Pair {
	for _, key := range f.Keys {
		if f.Selector == key.Name {
			return &key
		}
	}
	return nil
}

func (f *FlowOptions) GetKeys() []string {
	var keys []string
	for _, key := range f.Keys {
		keys = append(keys, key.Name)
	}
	return keys
}
