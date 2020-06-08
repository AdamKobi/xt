package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/pkg/logging"
	"github.com/apex/log"
	"github.com/spf13/viper"
)

const SSHBinary = "/usr/bin/ssh"
const SCPBinary = "/usr/bin/scp"

type SSHConfig struct {
}

type Output struct {
	Stdout string
	Stderr string
}

type Executer struct {
	Host      string
	RemoteCmd string
	Cmd       *exec.Cmd
}

//Connect will run command and wait for shell to return (mainly used for SSH)
func Connect(instance string) error {
	executer := NewSSHExecuter(instance, nil)
	return CommandWithTTY(executer)
}

//RunMultiple will run multiple commands on remote servers and return output/error
func RunMultiple(executers []*Executer) {
	outputChan := make(chan Output, 10)
	timeout := time.After(60 * time.Second)

	for _, executer := range executers {
		go func(executer *Executer) {
			outputChan <- CommandWithOutput(executer)
		}(executer)
	}
	for i := 0; i < len(executers); i++ {
		select {
		case output := <-outputChan:
			if output.Stderr != "" {
				logging.Main.WithFields(log.Fields{
					"host":    executers[i].Host,
					"command": executers[i].RemoteCmd,
				}).Errorf("command failed, %v", output.Stderr)
			} else {
				logging.Main.WithFields(log.Fields{
					"host":      executers[i].Host,
					"command":   executers[i].RemoteCmd,
					"exit-code": executers[i].Cmd.ProcessState,
				}).Info("command output")
				fmt.Print(output.Stdout)
			}
		case <-timeout:
			logging.Main.WithFields(log.Fields{
				"host":    executers[i].Host,
				"command": executers[i].RemoteCmd,
			}).Errorf("command timedout!")
		}
	}
}

//CommandWithOutput will run a single command and return stdout or stderr
func CommandWithOutput(executer *Executer) Output {
	logging.Main.WithFields(log.Fields{
		"host":    executer.Host,
		"command": executer.RemoteCmd,
	}).Info("runnning command...")
	r, err := executer.Cmd.CombinedOutput()
	outputStr := string(r)
	if err != nil {
		fmt.Println(executer.Cmd.ProcessState)
		return Output{Stderr: outputStr}
	}
	return Output{Stdout: outputStr}
}

//CommandWithTTY will run command and request for TTY
func CommandWithTTY(executer *Executer) error {
	executer.Cmd.Stdin = os.Stdin
	executer.Cmd.Stdout = os.Stdout
	executer.Cmd.Stderr = os.Stderr
	if err := executer.Cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getSSHArgs(host string) []string {
	current := config.GetProfile()
	domain := current.GetString("ssh.domain")
	sshUser := current.GetString("ssh.user")
	connStr := fmt.Sprintf("%s@%s%s", sshUser, host, domain)
	args := viper.GetStringSlice("ssh.args")
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
		Host:      instance,
		RemoteCmd: strings.Join(remoteCmd, " "),
		Cmd:       exec.Command(SSHBinary, args...),
	}
}

func CreateSSHExecuters(instanceIds []string, localCmd string, remoteCmd []string) []*Executer {
	var executers []*Executer
	for _, instanceId := range instanceIds {
		executer := NewSSHExecuter(instanceId, remoteCmd)
		executers = append(executers, executer)
	}
	return executers
}
