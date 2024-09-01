package config

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"gopkg.in/yaml.v2"
)

type KafkaConfig struct {
	Brokers  []string      `yaml:"brokers"`
	Consumer ConsumerConfig `yaml:"consumer"`
	Producer ProducerConfig `yaml:"producer"`
}

type ConsumerConfig struct {
	GroupID string   `yaml:"group_id"`
	Topics  []string `yaml:"topics"`
}

type ProducerConfig struct {
	Topic string `yaml:"topic"`
}

type TLSConfig struct {
	Enabled        bool   `yaml:"enabled"`
	ClientCertFile string `yaml:"client_cert_file"`
	ClientKeyFile  string `yaml:"client_key_file"`
	CACertFile     string `yaml:"ca_cert_file"`
}

type Config struct {
	Kafka KafkaConfig `yaml:"kafka"`
	TLS   TLSConfig   `yaml:"tls"`
}

func LoadConfig(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	// Загрузка клиентского сертификата
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return nil, err
	}

	// Загрузка CA сертификата
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}, nil
}
