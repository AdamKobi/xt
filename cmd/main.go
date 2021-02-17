package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"runtime"
	"strings"

	surveyCore "github.com/AlecAivazis/survey/v2/core"
	"github.com/adamkobi/xt/internal/api"
	"github.com/adamkobi/xt/internal/build"
	"github.com/adamkobi/xt/internal/config"
	"github.com/adamkobi/xt/internal/update"
	"github.com/adamkobi/xt/pkg/cmdutil"
	"github.com/adamkobi/xt/pkg/command/factory"
	"github.com/adamkobi/xt/pkg/command/root"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
)

var updaterEnabled = ""

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	buildDate := build.Date
	buildVersion := build.Version

	updateMessageChan := make(chan *update.ReleaseInfo)
	go func() {
		rel, _ := checkForUpdate(buildVersion)
		updateMessageChan <- rel
	}()

	hasDebug := os.Getenv("DEBUG") != ""

	cmdFactory := factory.New()
	stderr := cmdFactory.IOStreams.ErrOut
	if !cmdFactory.IOStreams.ColorEnabled() {
		surveyCore.DisableColor = true
	} else {
		// override survey's poor choice of color
		surveyCore.TemplateFuncsWithColor["color"] = func(style string) string {
			switch style {
			case "white":
				if cmdFactory.IOStreams.ColorSupport256() {
					return fmt.Sprintf("\x1b[%d;5;%dm", 38, 242)
				}
				return ansi.ColorCode("default")
			default:
				return ansi.ColorCode(style)
			}
		}
	}

	rootCmd := root.NewCmd(cmdFactory, buildVersion, buildDate)

	_, err := cmdFactory.Config()
	if err != nil {
		fmt.Fprintf(stderr, "failed to read configuration:  %s\n", err)
		os.Exit(2)
	}

	if cmd, err := rootCmd.ExecuteC(); err != nil {
		printError(stderr, err, cmd, hasDebug)
		os.Exit(1)
	}

	if root.HasFailed() {
		os.Exit(1)
	}

	newRelease := <-updateMessageChan
	if newRelease != nil {
		installCmd := "curl -s https://raw.githubusercontent.com/AdamKobi/xt/master/scripts/installer.sh | bash -s"
		fmt.Fprintf(stderr, "\n\n%s %s → %s\n",
			ansi.Color("A new release of xt is available:", "yellow"),
			ansi.Color(buildVersion, "cyan"),
			ansi.Color(newRelease.Version, "cyan"))
		fmt.Fprintf(stderr, "%s\n\n",
			ansi.Color(newRelease.URL, "yellow"))
		fmt.Fprintf(stderr, "install with %s", installCmd)
	}
}

func listenForInterrupt(stopScan chan os.Signal) {
	<-stopScan
	fmt.Fprintf(os.Stderr, "interupt received, exiting")
	os.Exit(1)
}

func checkForUpdate(currentVersion string) (*update.ReleaseInfo, error) {
	client := basicClient(currentVersion)
	repo := "adamkobi/xt"
	stateFilePath := path.Join(config.DefaultDir(), "state.yml")
	return update.CheckForUpdate(client, stateFilePath, repo, currentVersion)
}

// BasicClient returns an API client for github.com only that borrows from but
// does not depend on user configuration
func basicClient(currentVersion string) *api.Client {
	var opts []api.ClientOption

	opts = append(opts, api.AddHeader("User-Agent", fmt.Sprintf("Xt CLI %s", currentVersion)))
	token := os.Getenv("GITHUB_TOKEN")

	if token != "" {
		opts = append(opts, api.AddHeader("Authorization", fmt.Sprintf("token %s", token)))
	}
	return api.NewClient(opts...)
}

func printError(out io.Writer, err error, cmd *cobra.Command, debug bool) {
	if err == cmdutil.ErrSilent {
		return
	}

	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		fmt.Fprintf(out, "error connecting to %s\n", dnsError.Name)
		if debug {
			fmt.Fprintln(out, dnsError)
		}
		fmt.Fprintln(out, "check your internet connection or githubstatus.com")
		return
	}

	fmt.Fprintln(out, err)

	var flagError *cmdutil.FlagError
	if errors.As(err, &flagError) || strings.HasPrefix(err.Error(), "unknown command ") {
		if !strings.HasSuffix(err.Error(), "\n") {
			fmt.Fprintln(out)
		}
		fmt.Fprintln(out, cmd.UsageString())
	}
}
