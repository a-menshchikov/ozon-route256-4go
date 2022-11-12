package config

type (
	clientConfig struct {
		Telegram telegram `yaml:"tg"`
	}

	telegram struct {
		Token string `yaml:"token"`
	}
)
