package config

import (
	"time"
)

type storageDriver string

const (
	InMemoryDriver   storageDriver = "in_memory"
	PostgreSQLDriver storageDriver = "postgresql"
)

type StorageConfig struct {
	Driver      storageDriver `yaml:"driver"`
	Dsn         string        `yaml:"dsn"`
	WaitTimeout time.Duration `yaml:"wait_timeout"`
}
