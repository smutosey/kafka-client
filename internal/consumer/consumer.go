package consumer

import (
	"context"
	"log"
	"sync"
	"crypto/tls"

	"github.com/IBM/sarama"
)

type Consumer struct {
	group    sarama.ConsumerGroup
	topics   []string
}

func NewConsumer(brokers []string, groupID string, topics []string, tlsConfig *tls.Config) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Net.TLS.Enable = tlsConfig != nil
	config.Net.TLS.Config = tlsConfig
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		group:  group,
		topics: topics,
	}, nil
}

func (c *Consumer) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	handler := &consumerGroupHandler{}

	for {
		if err := c.group.Consume(ctx, c.topics, handler); err != nil {
			log.Printf("Ошибка при потреблении: %v", err)
			continue
		}

		if ctx.Err() != nil {
			return
		}
	}
}

func (c *Consumer) Close() error {
	return c.group.Close()
}

type consumerGroupHandler struct{}

func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Получено сообщение: %s", string(message.Value))
		session.MarkMessage(message, "")
	}
	return nil
}
