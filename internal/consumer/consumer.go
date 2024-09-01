package consumer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	// "strconv"
	"sync"
	"time"
    "crypto/tls"

	"github.com/IBM/sarama"
	// "github.com/prometheus/client_golang/prometheus"
	"github.com/smutosey/kafka-client/config"
	"github.com/smutosey/kafka-client/pkg/metrics"
)

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	config        *config.ConsumerConfig
}

func NewConsumer(brokers []string, consumerConfig *config.ConsumerConfig, tlsConfig *tls.Config) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	if tlsConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	consumerGroup, err := sarama.NewConsumerGroup(brokers, consumerConfig.GroupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		config:        consumerConfig,
	}, nil
}

func (c *Consumer) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	handler := consumerGroupHandler{config: c.config}

	for {
		if err := c.consumerGroup.Consume(ctx, []string{c.config.Topic}, &handler); err != nil {
			log.Printf("Ошибка при обработке сообщений: %v", err)
		}

		if ctx.Err() != nil {
			return
		}
	}
}

type consumerGroupHandler struct {
	config *config.ConsumerConfig
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	fileBuffers := make(map[string]*bytes.Buffer)

	for msg := range claim.Messages() {
		filename := fmt.Sprintf("%s-%d-%d", msg.Topic, msg.Partition, msg.Offset)
		buffer, exists := fileBuffers[filename]
		if !exists {
			buffer = new(bytes.Buffer)
			fileBuffers[filename] = buffer
		}

		buffer.Write(msg.Value)
		session.MarkMessage(msg, "")
		metrics.ConsumedMessages.WithLabelValues(msg.Topic).Inc()
	}

	for filename, buffer := range fileBuffers {
		filePath := filepath.Join(h.config.Path, filename)
		if err := os.WriteFile(filePath, buffer.Bytes(), 0644); err != nil {
			log.Printf("Ошибка при записи файла %s: %v", filePath, err)
			continue
		}

		log.Printf("Файл успешно записан: %s", filePath)
		metrics.FileProcessingTime.WithLabelValues(h.config.Topic).Observe(float64(time.Now().UnixNano()))
	}

	return nil
}
