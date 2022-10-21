package config

type storageDriver string

const (
	InMemoryDriver storageDriver = "in_memory"
)

type StorageConfig struct {
	Driver storageDriver `yaml:"driver"`
}
