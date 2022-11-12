package config

type cacheDriver string

const (
	RedisDriver cacheDriver = "redis"
)

type (
	cacheConfig struct {
		Reporter CacheSectionConfig `yaml:"reporter"`
		Rates    CacheSectionConfig `yaml:"rates"`
	}

	CacheSectionConfig struct {
		Driver cacheDriver `yaml:"driver"`
		Dsn    string      `yaml:"dsn"`
	}
)
