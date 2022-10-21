package config

type ClientConfig struct {
	Telegram Telegram `yaml:"tg"`
}

type Telegram struct {
	Token string `yaml:"token"`
}
