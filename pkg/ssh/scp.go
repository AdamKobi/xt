package ssh

import (
	"fmt"
	"os/exec"
	"path"

	"github.com/adamkobi/xt/config"
)

//NewSCPExecuter creates a new SCP executer with required args
func NewSCPExecuter(instance string, localFilePath, remoteFilePath string, upload bool) *Executer {
	args := getSCPArgs(instance, localFilePath, remoteFilePath, upload)
	return &Executer{
		Hostname:  instance,
		RemoteCmd: []string{localFilePath, remoteFilePath},
		Cmd:       exec.Command(scpBinary, args...),
	}
}

func getSCPArgs(host, localFilePath, remoteFilePath string, upload bool) []string {
	p := config.GetProfile()
	domain := p.SSH.Domain
	sshUser := p.SSH.User
	connStr := fmt.Sprintf("%s@%s%s:%s", sshUser, host, domain, remoteFilePath)
	args := p.SSH.Options
	if upload {
		args = append(args, localFilePath, connStr)
		return args
	}
	args = append(args, connStr, localFilePath)
	return args
}

//CreateSCPExecuters creates multiple SCP executers with same local and remote file paths
func CreateSCPExecuters(instanceIds []string, localFilePath, remoteFilePath string, upload bool) []*Executer {
	localDir := path.Dir(localFilePath)
	localFile := path.Base(localFilePath)

	//In case we copy same file name to local dir we need to have unique names
	if localFile == localDir {
		localFile = path.Base(remoteFilePath)
	}

	var executers []*Executer
	for _, instanceID := range instanceIds {
		if !upload && len(instanceIds) > 1 {
			localFilePath = path.Join(localDir, instanceID+"_"+localFile)
		}
		executers = append(executers, NewSCPExecuter(instanceID, localFilePath, remoteFilePath, upload))
	}
	return executers
}
