package main

import (
	"github.com/adamkobi/xt/command"
	"github.com/adamkobi/xt/pkg/logging"
)

func main() {
	if err := command.RootCmd.Execute(); err != nil {
		logging.Main.Error(err.Error())
	}
}
