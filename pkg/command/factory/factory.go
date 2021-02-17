package factory

import (
	"github.com/adamkobi/xt/internal/config"
	"github.com/adamkobi/xt/pkg/cmdutil"
	"github.com/adamkobi/xt/pkg/iostreams"
)

func New() *cmdutil.Factory {
	io := iostreams.System()

	var cachedConfig *config.Config
	var err error
	configFunc := func() (*config.Config, error) {
		if cachedConfig != nil {
			return cachedConfig, nil
		}
		cachedConfig, err = config.ParseDefaultConfig()
		if err != nil {
			return nil, err
		}
		// if errors.Is(configError, os.ErrNotExist) {
		// 	cachedConfig = config.NewBlankConfig()
		// 	configError = nil
		// }
		// cachedConfig = config.InheritEnv(cachedConfig)
		return cachedConfig, nil
	}

	return &cmdutil.Factory{
		IOStreams: io,
		Config:    configFunc,
	}
}
