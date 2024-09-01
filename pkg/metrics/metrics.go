package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	ConsumedMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_consumed_messages_total",
			Help: "Общее количество потребленных сообщений Kafka.",
		},
		[]string{"topic"},
	)

	ProducedMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_produced_messages_total",
			Help: "Общее количество отправленных сообщений Kafka.",
		},
		[]string{"topic"},
	)

	FileProcessingTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_processing_duration_seconds",
			Help:    "Продолжительность обработки файла в секундах.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(ConsumedMessages)
	prometheus.MustRegister(ProducedMessages)
	prometheus.MustRegister(FileProcessingTime)

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)
}
