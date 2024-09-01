package producer

import (
	"crypto/tls"
	"log"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(brokers []string, topic string, tlsConfig *tls.Config) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Net.TLS.Enable = tlsConfig != nil
	config.Net.TLS.Config = tlsConfig

	prod, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: prod,
		topic:    topic,
	}, nil
}

func (p *Producer) SendMessage(message string) {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	} else {
		log.Printf("Сообщение отправлено в партицию %d на смещение %d", partition, offset)
	}
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
