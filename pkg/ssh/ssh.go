package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/logging"
	"github.com/apex/log"
)

const sshBinary = "/usr/bin/ssh"
const scpBinary = "/usr/bin/scp"

type Executer struct {
	Hostname  string
	RemoteCmd []string
	Cmd       *exec.Cmd
}

type Output struct {
	Data  string
	Error bool
}

var logger = logging.GetLogger()

//Connect will run command and wait for shell to return (mainly used for SSH)
func Connect(instance string) {
	logger.Infof("Connecting to %s...", instance)
	e := NewSSHExecuter(instance, nil)
	e.CommandWithTTY()
}

//RunMultiple will run multiple commands on remote servers and return output/error
func RunMany(executers []*Executer) {
	output := make(chan Output, 10)
	timeout := time.After(60 * time.Second)

	for _, executer := range executers {
		go func(executer *Executer) {
			output <- executer.CommandWithOutput()
		}(executer)
	}
	for _, executer := range executers {
		ctx := logger.WithFields(log.Fields{
			"host":    executer.Hostname,
			"command": executer.RemoteCmd,
		})
		select {
		case out := <-output:
			if out.Error {
				ctx.Errorf("command failed")
				fmt.Fprint(os.Stderr, out.Data)
			} else {
				ctx.Info("command success")
				fmt.Fprint(os.Stdout, out.Data)
			}
		case <-timeout:
			ctx.Errorf("command timedout!")
		}
	}
}

//CommandWithOutput will run a single command and return stdout or stderr
func (e Executer) CommandWithOutput() Output {
	logger.WithFields(log.Fields{
		"host":    e.Hostname,
		"command": e.RemoteCmd,
	}).Info("runnning command...")

	output, err := e.Cmd.CombinedOutput()
	if err != nil {
		return Output{
			Data:  string(output),
			Error: true,
		}
	}
	return Output{
		Data: string(output),
	}
}

//CommandWithTTY will run command and request for TTY
func (e Executer) CommandWithTTY() {
	e.Cmd.Stdout = os.Stdout
	e.Cmd.Stderr = os.Stderr
	e.Cmd.Stdin = os.Stdin

	if err := e.Cmd.Run(); err != nil {
		panic(err)
	}
	os.Exit(0)
}

func getSSHArgs(host string) []string {
	p := config.GetProfile()
	domain := p.SSH.Domain
	sshUser := p.SSH.User
	connStr := fmt.Sprintf("%s@%s%s", sshUser, host, domain)
	args := p.SSH.Options
	args = append(args, connStr)
	return args
}

//NewExecuter creates a new SSH executer with required args
func NewSSHExecuter(instance string, remoteCmd []string) *Executer {
	args := getSSHArgs(instance)
	if remoteCmd != nil {
		args = append(args, remoteCmd...)
	}
	return &Executer{
		Hostname:  instance,
		RemoteCmd: remoteCmd,
		Cmd:       exec.Command(sshBinary, args...),
	}
}

func CreateSSHExecuters(instanceIds []string, remoteCmd []string) []*Executer {
	var executers []*Executer
	for _, instanceId := range instanceIds {
		executer := NewSSHExecuter(instanceId, remoteCmd)
		executers = append(executers, executer)
	}
	return executers
}
