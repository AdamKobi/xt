package executer

import (
	"fmt"
	"os/exec"
)

const SSH = "ssh"
const SCP = "scp"

type Factory struct {
	Cmd     *exec.Cmd
	Options Options
}

type Options struct {
	Hostname   string
	Hostnames  []string
	User       string
	Domain     string
	Binary     string
	Args       []string
	RemoteCmd  []string
	LocalPath  string
	RemotePath string
	Upload     bool
}

type Output struct {
	Stdout []byte
	Stderr []byte
}

//New creates a new executer for the required binary
func New(o *Options) (*Factory, error) {
	switch o.Binary {
	case SSH:
		return NewSSHExecuter(o)
	case SCP:
		return NewSCPExecuter(o)
	default:
		return nil, fmt.Errorf("unsupported binary ", o.Binary)
	}
}

func RunCommands(executers []Factory, output chan Output) {
	for _, e := range executers {
		go func(e Factory) {
			output <- e.CommandWithOutput()
		}(e)
	}
}
