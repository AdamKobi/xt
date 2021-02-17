package cmdutil

import (
	"github.com/adamkobi/xt/internal/config"
	"github.com/adamkobi/xt/pkg/iostreams"
)

//Factory provides config func and iostreams for all commands
type Factory struct {
	IOStreams *iostreams.IOStreams
	Config    func() (*config.Config, error)
}
