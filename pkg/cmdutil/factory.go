package cmdutil

import (
	"github.com/adamkobi/xt/internal/config"
	"github.com/adamkobi/xt/pkg/iostreams"
	"github.com/sirupsen/logrus"
)

type Factory struct {
	IOStreams *iostreams.IOStreams
	Config    func() (*config.Config, error)
	Log       func() *logrus.Logger
	Debug     func()
}
