package ssh

import (
	"fmt"
	"os/exec"
	"path"

	"github.com/adamkobi/xt/config"
	"github.com/spf13/viper"
)

//NewExecuter creates a new SSH executer with required args
func NewSCPExecuter(instance string, localFilePath, remoteFilePath string, upload bool) *Executer {
	args := getSCPArgs(instance, localFilePath, remoteFilePath, upload)
	return &Executer{
		Host:      instance,
		RemoteCmd: localFilePath + "|" + remoteFilePath,
		Cmd:       exec.Command(SCPBinary, args...),
	}
}

func getSCPArgs(host, localFilePath, remoteFilePath string, upload bool) []string {
	current := config.GetProfile()
	domain := current.GetString("ssh.domain")
	sshUser := current.GetString("ssh.user")
	connStr := fmt.Sprintf("%s@%s%s:%s", sshUser, host, domain, remoteFilePath)
	args := viper.GetStringSlice("ssh.args")
	if upload {
		args = append(args, localFilePath, connStr)
		return args
	}
	args = append(args, connStr, localFilePath)
	return args
}

func CreateSCPExecuters(instanceIds []string, localFilePath, remoteFilePath string, upload bool) []*Executer {
	localDir := path.Dir(localFilePath)
	localFile := path.Base(localFilePath)

	//In case we copy same file name to local dir we need to have unique names
	if localFile == localDir {
		localFile = path.Base(remoteFilePath)
	}

	var executers []*Executer
	for _, instanceId := range instanceIds {
		if !upload && len(instanceIds) > 1 {
			localFilePath = path.Join(localDir, instanceId+"_"+localFile)
		}
		executers = append(executers, NewSCPExecuter(instanceId, localFilePath, remoteFilePath, upload))
	}
	return executers
}
