package logging

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

//Logger is the main logger for XT
var Logger log.Logger = log.Logger{Handler: cli.Default, Level: log.DebugLevel}

//GetLogger returns a pointer to Logger
func GetLogger() *log.Logger {
	return &Logger
}
