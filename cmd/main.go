package main

import (
	"github.com/adamkobi/xt/command"
	"github.com/adamkobi/xt/pkg/logging"
)

var logger = logging.GetLogger()

func main() {
	if err := command.RootCmd.Execute(); err != nil {
		logger.Error(err.Error())
	}
}
