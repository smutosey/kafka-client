package producer

import (
    "crypto/tls"
    // "fmt"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/IBM/sarama"
    "github.com/smutosey/kafka-client/internal/metrics"
	"github.com/smutosey/kafka-client/config"
)

type Producer struct {
    producer  sarama.SyncProducer
    topicFiles []config.TopicFile
    tlsConfig  *tls.Config
}

func NewProducer(brokers []string, topicFiles []config.TopicFile, tlsConfig *tls.Config) (*Producer, error) {
    config := sarama.NewConfig()
    if tlsConfig != nil {
        config.Net.TLS.Enable = true
        config.Net.TLS.Config = tlsConfig
    }
    prod, err := sarama.NewSyncProducer(brokers, config)
    if err != nil {
        return nil, err
    }
    return &Producer{producer: prod, topicFiles: topicFiles, tlsConfig: tlsConfig}, nil
}

func (p *Producer) Close() error {
    return p.producer.Close()
}

func (p *Producer) MonitorAndSend() {
    for _, tf := range p.topicFiles {
        go p.monitorAndSendFiles(tf.Topic, tf.Path)
    }
}

func (p *Producer) monitorAndSendFiles(topic, path string) {
    log.Printf("Начало мониторинга каталога: %s для топика: %s", path, topic)

    for {
        files, err := os.ReadDir(path)
        if err != nil {
            log.Printf("Ошибка при чтении каталога: %v", err)
            time.Sleep(10 * time.Second)
            continue
        }

        for _, file := range files {
            if file.IsDir() {
                continue
            }
            filePath := filepath.Join(path, file.Name())
            log.Printf("Обработка файла: %s", filePath)

            content, err := os.ReadFile(filePath)
            if err != nil {
                log.Printf("Ошибка при чтении файла: %v", err)
                continue
            }

            message := &sarama.ProducerMessage{
                Topic: topic,
                Value: sarama.ByteEncoder(content),
            }

            start := time.Now()
            _, _, err = p.producer.SendMessage(message)
            duration := time.Since(start).Seconds()

            if err != nil {
                log.Printf("Ошибка при отправке сообщения: %v", err)
                metrics.IncProcessedFiles("failure")
            } else {
                log.Printf("Сообщение отправлено в топик %s", topic)
                metrics.IncProcessedFiles("success")
            }

            metrics.ObserveFileProcessingDuration("send_message", duration)
            err = os.Remove(filePath)
            if err != nil {
                log.Printf("Ошибка при удалении файла: %v", err)
            }
        }

        time.Sleep(10 * time.Second)
    }
}
