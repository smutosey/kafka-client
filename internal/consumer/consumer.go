package consumer

import (
	"context"
	"log"
	"sync"
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"path/filepath"

	"github.com/IBM/sarama"
	"github.com/smutosey/kafka-client/config"
)
type Consumer struct {
    consumers map[string]sarama.ConsumerGroup
    topicConfigs []config.ConsumerTopicConfig
    tlsConfig *tls.Config
}

func NewConsumer(brokers []string, topicConfigs []config.ConsumerTopicConfig, tlsConfig *tls.Config) (*Consumer, error) {
    config := sarama.NewConfig()
    if tlsConfig != nil {
        config.Net.TLS.Enable = true
        config.Net.TLS.Config = tlsConfig
    }

    consumers := make(map[string]sarama.ConsumerGroup)
    for _, tc := range topicConfigs {
        consumer, err := sarama.NewConsumerGroup(brokers, tc.GroupID, config)
        if err != nil {
            return nil, err
        }
        consumers[tc.Topic] = consumer
    }

    return &Consumer{consumers: consumers, topicConfigs: topicConfigs, tlsConfig: tlsConfig}, nil
}

func (c *Consumer) Close() error {
    var err error
    for _, consumer := range c.consumers {
        if e := consumer.Close(); e != nil {
            err = e
        }
    }
    return err
}

func (c *Consumer) Start(ctx context.Context, wg *sync.WaitGroup) {
    defer wg.Done()
    for _, tc := range c.topicConfigs {
        go c.consumeMessages(ctx, tc)
    }
}

func (c *Consumer) consumeMessages(ctx context.Context, tc config.ConsumerTopicConfig) {
    handler := &messageHandler{
        outputPath: tc.OutputPath,
    }

    for {
        select {
        case <-ctx.Done():
            return
        default:
            log.Printf("Подключение к топику: %s, группа: %s", tc.Topic, tc.GroupID)
            consumer, exists := c.consumers[tc.Topic]
            if !exists {
                log.Printf("Консьюмер для топика %s не найден", tc.Topic)
                return
            }

            if err := consumer.Consume(ctx, []string{tc.Topic}, handler); err != nil {
                log.Printf("Ошибка при потреблении сообщений из топика %s: %v", tc.Topic, err)
            }
        }
    }
}

type messageHandler struct {
    outputPath string
}

func (h *messageHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *messageHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *messageHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    fileData := make(map[string][]byte)
    for msg := range claim.Messages() {
        fileName := fmt.Sprintf("%s/%d-%s", h.outputPath, msg.Partition, strconv.Itoa(int(msg.Offset)))
        fileData[fileName] = append(fileData[fileName], msg.Value...)

        log.Printf("Получено сообщение из топика %s: %s", msg.Topic, string(msg.Value))
        session.MarkMessage(msg, "")
    }

    for fileName, data := range fileData {
        if err := h.saveToFile(fileName, data); err != nil {
            log.Printf("Ошибка при сохранении файла %s: %v", fileName, err)
        } else {
            log.Printf("Файл сохранен: %s", fileName)
        }
    }
    return nil
}

func (h *messageHandler) saveToFile(fileName string, data []byte) error {
    filePath := filepath.Join(h.outputPath, fileName)
    if err := os.WriteFile(filePath, data, 0644); err != nil {
        return err
    }

    // Дополнительная проверка размера
    origSize, err := getOriginalFileSize(filePath) // Эту функцию нужно реализовать
    if err != nil {
        return err
    }
    savedSize := int64(len(data))
    if savedSize != origSize {
        return fmt.Errorf("размеры файлов не совпадают: ожидалось %d, получено %d", origSize, savedSize)
    }

    return nil
}

func getOriginalFileSize(filePath string) (int64, error) {
    // Реализуйте функцию для получения размера оригинального файла
    return 0, nil
}