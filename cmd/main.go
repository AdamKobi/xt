package main

import (
	"fmt"
	"os"
	"path"

	rootCmd "github.com/adamkobi/xt/command/root"

	"github.com/adamkobi/xt/command/factory"
	"github.com/adamkobi/xt/config"
	"github.com/adamkobi/xt/internal/api"
	"github.com/adamkobi/xt/internal/build"
	"github.com/adamkobi/xt/internal/update"
	"github.com/mgutz/ansi"
)

var updaterEnabled = ""

func main() {
	buildDate := build.Date
	buildVersion := build.Version

	updateMessageChan := make(chan *update.ReleaseInfo)
	go func() {
		rel, _ := checkForUpdate(buildVersion)
		updateMessageChan <- rel
	}()

	f := factory.New()
	log := f.Log()
	_, err := f.Config()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	rootCmd := rootCmd.NewCmdRoot(f, buildVersion, buildDate)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	newRelease := <-updateMessageChan
	if newRelease != nil {
		msg := fmt.Sprintf("%s %s → %s\n%s",
			ansi.Color("A new release of xt is available:", "yellow"),
			ansi.Color(buildVersion, "cyan"),
			ansi.Color(newRelease.Version, "cyan"),
			ansi.Color(newRelease.URL, "yellow"))

		fmt.Fprintf(os.Stdout, "\n\n%s\n\n", msg)
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
