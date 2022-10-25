package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Client   ClientConfig   `yaml:"client"`
	Storage  StorageConfig  `yaml:"storage"`
	Currency CurrencyConfig `yaml:"currency"`
}

func New(configPath string) (*Config, error) {
	c := &Config{}

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
