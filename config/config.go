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
    Topics []ConsumerTopicConfig `yaml:"topics"`
}

type ConsumerTopicConfig struct {
    Topic      string `yaml:"topic"`
    GroupID    string `yaml:"group_id"`
    OutputPath string `yaml:"output_path"`
}

type ProducerConfig struct {
    TopicFiles []TopicFile `yaml:"topic_files"`
}

type TopicFile struct {
    Topic string `yaml:"topic"`
    Path  string `yaml:"path"`
}

type TLSConfig struct {
	Enabled        bool   `yaml:"enabled"`
	ClientCertFile string `yaml:"client_cert_file"`
	ClientKeyFile  string `yaml:"client_key_file"`
	CACertFile     string `yaml:"ca_cert_file"`
}

type Config struct {
    Kafka struct {
        Brokers  []string       `yaml:"brokers"`
        Producer ProducerConfig `yaml:"producer"`
        Consumer ConsumerConfig `yaml:"consumer"`
    } `yaml:"kafka"`
    TLS TLSConfig `yaml:"tls"`
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
