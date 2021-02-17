package executer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adamkobi/xt/pkg/iostreams"
	"github.com/cli/safeexec"
	"github.com/creack/pty"
)

//SSH is a constant describing SSH executer
const SSH = "ssh"

//SCP is a constant describing SCP executer
const SCP = "scp"

//Cmd is basicly a wrapper around exec.Cmd
type Cmd struct {
	Hostname string
	Exec     *exec.Cmd
}

//Options describes all parameters Cmd can receive
type Options struct {
	IO         *iostreams.IOStreams
	Selected   string
	Hostnames  []string
	User       string
	Domain     string
	Binary     string
	Args       []string
	RemoteCmd  []string
	LocalPath  string
	RemotePath string
	Download   bool
}

//New creates a new executer for the required binary
func New(options *Options) (*Cmd, error) {
	switch options.Binary {
	case SSH:
		return newSSH(options)
	case SCP:
		return newSCP(options)
	default:
		return nil, fmt.Errorf("binary %s not suppoerted", options.Binary)
	}
}

//newSSH creates a new SSH executer with required args
func newSSH(options *Options) (*Cmd, error) {
	if err := validate(options); err != nil {
		return nil, err
	}

	sshExe, err := safeexec.LookPath("ssh")
	if err != nil {
		return nil, err
	}

	connStr := fmt.Sprintf("%s@%s%s", options.User, options.Selected, options.Domain)
	options.Args = append(options.Args, connStr)

	if options.RemoteCmd != nil {
		options.Args = append(options.Args, options.RemoteCmd...)
	}

	return &Cmd{
		Exec:     exec.Command(sshExe, options.Args...),
		Hostname: options.Selected,
	}, nil
}

//newSCP creates a new SCP executer with required args
func newSCP(options *Options) (*Cmd, error) {
	if err := validate(options); err != nil {
		return nil, err
	}

	scpExe, err := safeexec.LookPath("scp")
	if err != nil {
		return nil, err
	}

	connStr := fmt.Sprintf("%s@%s%s:%s", options.User, options.Selected, options.Domain, options.RemotePath)
	if options.Download {
		options.Args = append(options.Args, connStr, options.LocalPath)
	} else {
		options.Args = append(options.Args, options.LocalPath, connStr)
	}

	return &Cmd{
		Exec:     exec.Command(scpExe, options.Args...),
		Hostname: options.Selected,
	}, nil
}

func validate(o *Options) error {
	if o.Selected == "" {
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

//Output will run a single command and return stdout or stderr
func (c Cmd) Output() ([]byte, error) {
	if os.Getenv("DEBUG") != "" {
		_ = printArgs(os.Stderr, c.Exec.Args)
	}
	if c.Exec.Stderr != nil {
		return c.Exec.Output()
	}
	errStream := &bytes.Buffer{}
	c.Exec.Stderr = errStream
	out, err := c.Exec.Output()
	if err != nil {
		err = &CmdError{errStream, c.Hostname, c.Exec.Args, err}
	}
	return out, err
}

//Start is a wrapper around pty.Start()
func (c *Cmd) Start() (*os.File, error) {
	return pty.Start(c.Exec)
}

//Connect will run command and request for TTY
func (c *Cmd) Connect() error {
	if os.Getenv("DEBUG") != "" {
		_ = printArgs(os.Stderr, c.Exec.Args)
	}
	c.Exec.Stdout = os.Stdout
	c.Exec.Stderr = os.Stderr
	c.Exec.Stdin = os.Stdin
	return c.Exec.Run()
}

func printOutput(io *iostreams.IOStreams, f *os.File, wg *sync.WaitGroup, c *Cmd) {
	out := io.Out
	cs := io.ColorScheme()
	defer wg.Done()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(out, "%s | %s\n", cs.Green(c.Hostname), line)
	}
	fmt.Println()
}

//RunCommands takes Cmd slice and executes all commands using go routines then prints output accordingly
func RunCommands(io *iostreams.IOStreams, executers []*Cmd) {
	out := io.Out
	errOut := io.ErrOut
	cs := io.ColorScheme()
	var wg sync.WaitGroup
	for _, c := range executers {
		wg.Add(1)
		f, err := c.Start()
		if err != nil {
			fmt.Fprintf(errOut, "%s error starting command\n%s\n", cs.WarningIcon(), err)
			os.Exit(1)
		}
		fmt.Fprintf(out, "running command on %s\n", cs.Bold(c.Hostname))
		go printOutput(io, f, &wg, c)
	}
	wg.Wait()
}

//CreateAll creates executers for all hostnames listed
func CreateAll(opts *Options) ([]*Cmd, error) {
	var executers []*Cmd
	var err error
	for _, host := range opts.Hostnames {
		hostOpts := *opts
		hostOpts.Selected = host
		if hostOpts.Download && hostOpts.LocalPath != "" {
			hostOpts.LocalPath, err = prepareDownloads(hostOpts.LocalPath, host)
			if err != nil {
				return nil, err
			}
		}
		executer, err := New(&hostOpts)
		if err != nil {
			return nil, err
		}
		executers = append(executers, executer)
	}
	return executers, nil
}

func prepareDownloads(localPath, host string) (string, error) {
	fi, err := os.Stat(localPath)
	if err != nil {
		return "", err
	}

	if !fi.IsDir() {
		return "", fmt.Errorf("local path must be a directory when downloading from multiple servers")
	}
	baseDir := path.Dir(localPath)
	hostPath := path.Join(baseDir, host)
	if err := os.MkdirAll(hostPath, 0700); err != nil {
		return "", err
	}
	return hostPath, nil
}

// CmdError provides more visibility into why an exec.Cmd had failed
type CmdError struct {
	Stderr   *bytes.Buffer
	Hostname string
	Args     []string
	Err      error
}

func (e CmdError) Error() string {
	msg := e.Stderr.String()
	if msg != "" && !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	return fmt.Sprintf("%s%s: %s", msg, e.Args[0], e.Err)
}

func printArgs(w io.Writer, args []string) error {
	if len(args) > 0 {
		// print commands, but omit the full path to an executable
		args = append([]string{filepath.Base(args[0])}, args[1:]...)
	}
	_, err := fmt.Fprintf(w, "%v\n", args)
	return err
}
