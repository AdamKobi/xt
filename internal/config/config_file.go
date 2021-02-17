package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"syscall"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

//DefaultDir returns config directory
func DefaultDir() string {
	dir, _ := homedir.Expand("~/.xt")
	return dir
}

//DefaultFile
func DefaultFile() string {
	return path.Join(DefaultDir(), "config.yaml")
}

func ParseDefaultConfig() (*Config, error) {
	return parseConfigFile(DefaultFile())
}

func ParseConfig(filename string) (*Config, error) {
	return parseConfigFile(filename)
}

func WriteDefaultConfigFile(data []byte) error {
	return WriteConfigFile(DefaultFile(), data)
}

func ReadConfigFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, pathError(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func WriteConfigFile(filename string, data []byte) error {
	err := os.MkdirAll(path.Dir(filename), 0771)
	if err != nil {
		return pathError(err)
	}

	cfgFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) // cargo coded from setup
	if err != nil {
		return err
	}
	defer cfgFile.Close()

	n, err := cfgFile.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}

	return err
}

func parseConfigFile(filename string) (*Config, error) {
	data, err := ReadConfigFile(filename)
	if err != nil {
		return nil, err
	}

	cfg, err := parseConfigData(data)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

func parseConfigData(data []byte) (*Config, error) {
	var cfg Config
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func pathError(err error) error {
	var pathError *os.PathError
	if errors.As(err, &pathError) && errors.Is(pathError.Err, syscall.ENOTDIR) {
		if p := findRegularFile(pathError.Path); p != "" {
			return fmt.Errorf("remove or rename regular file `%s` (must be a directory)", p)
		}

	}
	return err
}

func findRegularFile(p string) string {
	for {
		if s, err := os.Stat(p); err == nil && s.Mode().IsRegular() {
			return p
		}
		newPath := path.Dir(p)
		if newPath == p || newPath == "/" || newPath == "." {
			break
		}
		p = newPath
	}
	return ""
}
