package producer

import (
	"crypto/tls"
	"log"
	"os"
	"path/filepath"
	"time"
	"context"
	"sync"

	"github.com/IBM/sarama"
	"github.com/smutosey/kafka-client/config"
	"github.com/smutosey/kafka-client/pkg/metrics"
)

type Producer struct {
	producer sarama.SyncProducer
	config   *config.ProducerConfig
}

func NewProducer(brokers []string, producerConfig *config.ProducerConfig, tlsConfig *tls.Config) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	if tlsConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
		config:   producerConfig,
	}, nil
}

func (p *Producer) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			files, err := os.ReadDir(p.config.Path)
			if err != nil {
				log.Printf("Ошибка при чтении директории: %v", err)
				continue
			}

			for _, file := range files {
				if err := p.sendFile(filepath.Join(p.config.Path, file.Name())); err != nil {
					log.Printf("Ошибка при отправке файла: %v", err)
				}
			}
		}
	}
}

func (p *Producer) sendFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.config.Topic,
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Файл успешно отправлен: %s", filePath)
	metrics.ProducedMessages.WithLabelValues(p.config.Topic).Inc()

	return nil
}
