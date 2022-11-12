package config

import (
	"time"
)

type (
	ReportsConfig struct {
		Grpc  ReportsGrpcConfig  `yaml:"grpc"`
		Kafka ReportsKafkaConfig `yaml:"kafka"`
	}

	ReportsGrpcConfig struct {
		ServerPort    uint16 `yaml:"server_port"`
		ClientAddress string `yaml:"client_addr"`
	}

	ReportsKafkaConfig struct {
		Brokers       []string      `yaml:"brokers"`
		Assignor      string        `yaml:"assignor"`
		Timeout       time.Duration `yaml:"timeout"`
		Topic         string        `yaml:"topic"`
		ConsumerGroup string        `yaml:"cg"`
	}
)
