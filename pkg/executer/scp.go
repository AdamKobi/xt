package executer

import (
	"fmt"
	"os/exec"
)

//NewSCPExecuter creates a new SCP executer with required args
func NewSCPExecuter(o *Options) (*Factory, error) {
	if err := validate(o); err != nil {
		return nil, err
	}
	args := o.scpArgs()
	return &Factory{
		Cmd:     exec.Command(SCP, args...),
		Options: *o,
	}, nil
}

func (o *Options) scpArgs() []string {
	var args []string
	connStr := fmt.Sprintf("%s@%s%s:%s", o.User, o.Hostname, o.Domain, o.RemotePath)
	args = append(args, o.Args...)
	if o.Upload {
		args = append(args, o.LocalPath, connStr)
		return args
	}
	args = append(args, connStr, o.LocalPath)
	return args
}
