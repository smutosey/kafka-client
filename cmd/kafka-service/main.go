package main

import (
	"context"
	"log"
	"os"
	"os/signal"
    "io"
	"sync"
	"syscall"

	"github.com/smutosey/kafka-client/config"
	"github.com/smutosey/kafka-client/internal/consumer"
	"github.com/smutosey/kafka-client/internal/producer"
	"github.com/smutosey/kafka-client/internal/tls"
	"github.com/smutosey/kafka-client/pkg/metrics"
)

func main() {
	cfg, err := config.LoadConfig("application.yml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	logFile, err := os.OpenFile(cfg.Logging.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Ошибка открытия файла логов: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	tlsConfig, err := tls.LoadTLSConfig(cfg.TLSConfig)
	if err != nil {
		log.Fatalf("Ошибка загрузки TLS конфигурации: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	for _, consumerCfg := range cfg.Kafka.Consumers {
		c, err := consumer.NewConsumer(cfg.Kafka.Brokers, &consumerCfg, tlsConfig)
		if err != nil {
			log.Fatalf("Ошибка создания консьюмера: %v", err)
		}
		wg.Add(1)
		go c.Start(ctx, wg)
	}

	for _, producerCfg := range cfg.Kafka.Producers {
		p, err := producer.NewProducer(cfg.Kafka.Brokers, &producerCfg, tlsConfig)
		if err != nil {
			log.Fatalf("Ошибка создания продюсера: %v", err)
		}
		wg.Add(1)
		go p.Start(ctx, wg)
	}

	metrics.RegisterMetrics()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cancel()
	wg.Wait()
}
