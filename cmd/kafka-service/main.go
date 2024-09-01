package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/myuser/kafka-service/config"
	"github.com/myuser/kafka-service/internal/consumer"
	"github.com/myuser/kafka-service/internal/producer"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("configs/application.yml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создаем TLS конфигурацию
	tlsConfig, err := config.NewTLSConfig(cfg.TLS.ClientCertFile, cfg.TLS.ClientKeyFile, cfg.TLS.CACertFile)
	if err != nil {
		log.Fatalf("Ошибка настройки TLS: %v", err)
	}

	// Настройка контекста для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализируем продюсера
	prod, err := producer.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Producer.Topic, tlsConfig)
	if err != nil {
		log.Fatalf("Ошибка создания продюсера: %v", err)
	}
	defer prod.Close()

	// Инициализируем консьюмера
	cons, err := consumer.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Consumer.GroupID, cfg.Kafka.Consumer.Topics, tlsConfig)
	if err != nil {
		log.Fatalf("Ошибка создания консьюмера: %v", err)
	}
	defer cons.Close()

	// Запуск консьюмера и продюсера
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go cons.Start(ctx, wg)

	go func() {
		for {
			prod.SendMessage("Пример сообщения")
			time.Sleep(2 * time.Second)
		}
	}()

	// Обработка сигнала завершения
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
	log.Println("Завершение работы...")
	cancel()
	wg.Wait()
}
