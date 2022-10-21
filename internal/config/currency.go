package config

import (
	"time"
)

type CurrencyConfig struct {
	Available       []Currency    `yaml:"available"`
	Base            string        `yaml:"base"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
}

type Currency struct {
	Code string `yaml:"code"`
	Flag string `yaml:"flag"`
}
