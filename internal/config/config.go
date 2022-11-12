package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type config struct {
	Client   clientConfig   `yaml:"client"`
	Storage  StorageConfig  `yaml:"storage"`
	Cache    cacheConfig    `yaml:"cache"`
	Currency CurrencyConfig `yaml:"currency"`
	Reports  ReportsConfig  `yaml:"reports"`
}

func NewConfig(configPath string) (*config, error) {
	c := &config{}

	rawYAML, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}

	err = yaml.Unmarshal(rawYAML, &c)
	if err != nil {
		return nil, errors.Wrap(err, "parsing yaml")
	}

	return c, nil
}
