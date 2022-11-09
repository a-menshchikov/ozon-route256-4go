package config

type cacheDriver string

const (
	RedisDriver cacheDriver = "redis"
)

type CacheConfig struct {
	Report CacheSectionConfig `yaml:"report"`
	Rates  CacheSectionConfig `yaml:"rates"`
}

type CacheSectionConfig struct {
	Driver cacheDriver `yaml:"driver"`
	Dsn    string      `yaml:"dsn"`
}
