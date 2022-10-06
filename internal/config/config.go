package config

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Token    string   `yaml:"token"`
	Currency Currency `yaml:"currency"`
}

type Currency struct {
	Available       []CurrencyConfig `yaml:"available"`
	Base            string           `yaml:"base"`
	RefreshInterval time.Duration    `yaml:"refresh_interval"`
}

type CurrencyConfig struct {
	Code string `yaml:"code"`
	Flag string `yaml:"flag"`
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
