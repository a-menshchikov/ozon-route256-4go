package config

type cacheDriver string

const (
	RedisDriver cacheDriver = "redis"
)

type CacheConfig struct {
	Reporter CacheSectionConfig `yaml:"reporter"`
	Rates    CacheSectionConfig `yaml:"rates"`
}

type CacheSectionConfig struct {
	Driver cacheDriver `yaml:"driver"`
	Dsn    string      `yaml:"dsn"`
}
