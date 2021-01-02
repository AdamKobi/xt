package executer

import (
	"fmt"
	"os"
	"os/exec"
)

//NewExecuter creates a new SSH executer with required args
func NewSSHExecuter(o *Options) (*Factory, error) {
	if err := validate(o); err != nil {
		return nil, err
	}

	args := o.sshArgs()
	if o.RemoteCmd != nil {
		args = append(args, o.RemoteCmd...)
	}
	return &Factory{
		Cmd:     exec.Command(SSH, args...),
		Options: *o,
	}, nil
}

func validate(o *Options) error {
	if o.Hostname == "" {
		return fmt.Errorf("hostname must be set")
	}

	if o.User == "" {
		return fmt.Errorf("user must be set")
	}

	if o.Domain == "" {
		return fmt.Errorf("domain must be set")
	}
	return nil
}

//Connect will run command and wait for shell to return (mainly used for SSH)
func (e *Factory) Connect() {
	e.CommandWithTTY()
}

//CommandWithOutput will run a single command and return stdout or stderr
func (e *Factory) CommandWithOutput() Output {
	output, err := e.Cmd.CombinedOutput()
	if err != nil {
		return Output{
			Stderr: output,
		}
	}
	return Output{
		Stdout: output,
	}
}

//CommandWithTTY will run command and request for TTY
func (e *Factory) CommandWithTTY() error {
	e.Cmd.Stdout = os.Stdout
	e.Cmd.Stderr = os.Stderr
	e.Cmd.Stdin = os.Stdin
	return e.Cmd.Run()
}

func (o *Options) sshArgs() []string {
	var args []string
	connStr := fmt.Sprintf("%s@%s%s", o.User, o.Hostname, o.Domain)
	args = append(args, o.Args...)
	args = append(args, connStr)
	return args
}
