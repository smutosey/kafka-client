package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"github.com/smutosey/kafka-client/config"
)

// LoadTLSConfig загружает и возвращает TLS конфигурацию.
func LoadTLSConfig(tlsCfg config.TLSConfig) (*tls.Config, error) {
	if !tlsCfg.Enabled {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(tlsCfg.CertFile, tlsCfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load X509 key pair: %w", err)
	}

	caCert, err := os.ReadFile(tlsCfg.CaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA file: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}, nil
}
