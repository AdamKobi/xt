package logging

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

//Main is the global logger
var Main log.Logger = log.Logger{Handler: cli.Default, Level: log.DebugLevel}
