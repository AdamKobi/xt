package factory

import (
	"github.com/adamkobi/xt/config"
	"github.com/sirupsen/logrus"
)

type CmdConfig struct {
	Config func() (*config.Config, error)
	Log    func() *logrus.Logger
	Debug  func()
}

func New() *CmdConfig {
	var cachedConfig *config.Config
	var cachedLog *logrus.Logger
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

	logFunc := func() *logrus.Logger {
		if cachedLog != nil {
			return cachedLog
		}
		cachedLog = logrus.New()
		cachedLog.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: true,
		})
		return cachedLog
	}

	debugFunc := func() {
		cachedLog.SetLevel(logrus.DebugLevel)
	}

	return &CmdConfig{
		Config: configFunc,
		Log:    logFunc,
		Debug:  debugFunc,
	}
}
