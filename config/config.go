package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mgutz/ansi"
)

const JSON = "json"
const notSetError = "%s must be set"

var selectedProfile string

type Config struct {
	FlowOptions    map[string][]FlowOptions  `yaml:"flows"`
	ProfileOptions map[string]ProfileOptions `yaml:"profiles"`
	SSHOptions     SSHOptions                `yaml:"ssh"`
}

type FlowOptions struct {
	Run      string   `yaml:"run"`
	Selector string   `yaml:"selector,omitempty"`
	Parse    string   `yaml:"parse,omitempty"`
	Type     string   `yaml:"type,omitempty"`
	Keys     []string `yaml:"keys,omitempty"`
	Print    bool     `yaml:"print,omitempty"`
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

//SSHArgs returns ssh options if they exist in profile else returns default
func (p *ProfileOptions) SCPArgs() []string {
	sshOptions := p.SSHOptions.Args
	if len(sshOptions) != 0 {
		return sshOptions
	}
	return defaultSCPOptions()
}

func defaultSSHOptions() []string {
	return []string{
		"-ta", "-C", "-o", "LogLevel=ERROR", "-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null", "-o", "ControlPath=~/.ssh/cm-%C",
		"-o", "ControlMaster=auto", "-o", "ControlPersist=5m"}
}

func defaultSCPOptions() []string {
	return []string{
		"-C", "-o", "LogLevel=ERROR", "-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null", "-o", "ControlPath=~/.ssh/cm-%C",
		"-o", "ControlMaster=auto", "-o", "ControlPersist=5m"}
}

//GetProfiles return a slice of the profiles found in config file
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

//GetFlows returns flows part of config
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

//GetProfile return a subset of config file by profile selected in flags
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

func (p *ProfileOptions) Message() {
	if p.DisplayMsg != "" {
		msgSlice := strings.Split(p.DisplayMsg, "\n")
		maxLength := len(msgSlice[0])
		for _, msgPart := range msgSlice[1:] {
			if len(msgPart) > maxLength {
				maxLength = len(msgPart)
			}
		}
		border := strings.Repeat("*", maxLength+4)
		fmt.Fprintf(os.Stdout, ansi.Color(border+"\n", "yellow+b"))
		for _, msgPart := range msgSlice {
			msgPartLen := len(msgPart)
			spacesRequired := maxLength - msgPartLen + 1
			formatedMsgPart := fmt.Sprintf("* %s%s*\n", msgPart, strings.Repeat(" ", spacesRequired))
			fmt.Fprintf(os.Stdout, ansi.Color(formatedMsgPart, "yellow+b"))
		}
		fmt.Fprintf(os.Stdout, ansi.Color(border+"\n", "yellow+b"))
	}
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

//GetProvider returns the selected profile config
func (p *ProfileOptions) Provider() ProviderOptions {
	return p.ProviderOptions
}

func (f *FlowOptions) Validate() error {
	if f.Run == "" {
		return fmt.Errorf("run is required for running a flow")
	}
	switch f.Type {
	case JSON:
		if f.Parse == "" {
			return fmt.Errorf("parse must be set when using json type")
		}
		if f.Selector == "" {
			return fmt.Errorf("selector must be set when using json type")
		}
		return nil
	default:
		return nil
	}
}
