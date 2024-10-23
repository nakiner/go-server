package server

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func loadLocalValuesYaml() {
	var localFile string
	for i, v := range os.Args {
		if v == "--config" && i+1 < len(os.Args) {
			localFile = os.Args[i+1]
			break
		}
		if strings.HasPrefix(v, "--config=") {
			parts := strings.SplitN(v, "=", 2)
			localFile = parts[1]
			break
		}
	}

	if localFile != "" {
		if err := loadEnvFromValuesFile(localFile); err != nil {
			panic(err)
		}
	}
}

func loadEnvFromValuesFile(file string) error {
	b, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return errors.Wrap(err, "read yaml file")
	}

	var config valuesYamlConfig
	if err := yaml.Unmarshal(b, &config); err != nil {
		return errors.Wrapf(err, "unmarshal %s", file)
	}

	for _, v := range config.Env {
		if err = os.Setenv(v.Name, v.Value); err != nil {
			return errors.Wrapf(err, "set env %s='%s'", v.Name, v.Value)
		}
		appLogger.With(
			zap.String("config_file", file),
		).Debugf("set env %s='%s'", v.Name, v.Value)
	}

	return nil
}

type valuesYamlConfig struct {
	Env []struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	} `yaml:"env"`
}
