package main

import (
    "context"
    "log"
    "sync"
    "time"

    "github.com/smutosey/kafka-client/internal/consumer"
    "github.com/smutosey/kafka-client/internal/metrics"
    "github.com/smutosey/kafka-client/internal/producer"
    "github.com/smutosey/kafka-client/config"
)

func main() {
    // Загрузка конфигурации
    cfg, err := config.LoadConfig("configs/application.yml")
    if err != nil {
        log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
    }

    // Настройка TLS
    tlsConfig, err := config.NewTLSConfig(cfg.TLS.ClientCertFile, cfg.TLS.ClientKeyFile, cfg.TLS.CACertFile)
    if err != nil && cfg.TLS.Enabled {
        log.Fatalf("Ошибка при создании TLS-конфигурации: %v", err)
    }

    // Инициализация метрик
    metrics.StartHTTPServer(":2112")

    // Создание продюсера
    prod, err := producer.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Producer.TopicFiles, tlsConfig)
    if err != nil {
        log.Fatalf("Ошибка при создании продюсера: %v", err)
    }
    defer prod.Close()

    // Запуск мониторинга каталогов
    prod.MonitorAndSend()

    // Создание консьюмера
    cons, err := consumer.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Consumer.Topics, tlsConfig)
    if err != nil {
        log.Fatalf("Ошибка при создании консьюмера: %v", err)
    }
    defer cons.Close()

    // Запуск консьюмера
    ctx, cancel := context.WithCancel(context.Background())
    var wg sync.WaitGroup
    wg.Add(1)
    go cons.Start(ctx, &wg)

    // Ожидание завершения работы
    time.Sleep(10 * time.Minute)
    cancel()
    wg.Wait()
}
