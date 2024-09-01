package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "net/http"
)

var (
    fileProcessingDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "file_processing_duration_seconds",
            Help:    "Duration of file processing in seconds.",
            Buckets: prometheus.DefBuckets,
        },
        []string{"operation"},
    )
    processedFiles = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "processed_files_total",
            Help: "Total number of processed files.",
        },
        []string{"status"},
    )
)

func init() {
    // Register metrics with Prometheus
    prometheus.MustRegister(fileProcessingDuration)
    prometheus.MustRegister(processedFiles)
}

// StartHTTPServer starts the HTTP server to expose Prometheus metrics.
func StartHTTPServer(addr string) {
    http.Handle("/metrics", promhttp.Handler())
    go func() {
        if err := http.ListenAndServe(addr, nil); err != nil {
            panic(err)
        }
    }()
}

// ObserveFileProcessingDuration records the duration of file processing.
func ObserveFileProcessingDuration(operation string, duration float64) {
    fileProcessingDuration.WithLabelValues(operation).Observe(duration)
}

// IncProcessedFiles increases the count of processed files.
func IncProcessedFiles(status string) {
    processedFiles.WithLabelValues(status).Inc()
}
