package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config содержит конфигурацию всего приложения.
type Config struct {
	Kafka     KafkaConfig  `yaml:"kafka"`
	Logging   LoggingConfig `yaml:"logging"`
	TLSConfig TLSConfig    `yaml:"tls"`
}

// KafkaConfig содержит конфигурацию Kafka.
type KafkaConfig struct {
	Brokers  []string       `yaml:"brokers"`
	Consumers []ConsumerConfig `yaml:"consumers"`
	Producers []ProducerConfig `yaml:"producers"`
}

// ConsumerConfig содержит конфигурацию для каждого консьюмера.
type ConsumerConfig struct {
	Topic   string `yaml:"topic"`
	GroupID string `yaml:"group"`
	Path    string `yaml:"path"`
}

// ProducerConfig содержит конфигурацию для каждого продюсера.
type ProducerConfig struct {
	Topic string `yaml:"topic"`
	Path  string `yaml:"path"`
}

// LoggingConfig содержит конфигурацию логирования.
type LoggingConfig struct {
	FilePath string `yaml:"file_path"`
}

// TLSConfig содержит конфигурацию TLS.
type TLSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	CaFile     string `yaml:"ca_file"`
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
	Passphrase string `yaml:"passphrase"`
}

// LoadConfig загружает конфигурацию из файла.
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
