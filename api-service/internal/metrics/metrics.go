package metrics

import (
	"bufio"
	"database/sql"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

type Metrics struct {
	Registry      *prometheus.Registry
	HTTPRequests  *prometheus.CounterVec
	HTTPDuration  *prometheus.HistogramVec
	SQLDuration   *prometheus.HistogramVec
	SQLErrors     *prometheus.CounterVec
	WebSockets    prometheus.Gauge
	ExternalCalls *prometheus.CounterVec
	ExternalTime  *prometheus.HistogramVec
	Fallbacks     *prometheus.CounterVec
	KafkaRecords  prometheus.Counter
	KafkaErrors   *prometheus.CounterVec
	KafkaLag      prometheus.Gauge
}

func New(service string, sqlDB *sql.DB) *Metrics {
	registry := prometheus.NewRegistry()

	metrics := &Metrics{
		Registry:      registry,
		HTTPRequests:  prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_http_requests_total", ConstLabels: prometheus.Labels{"service": service}}, []string{"method", "route", "status"}),
		HTTPDuration:  prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "app_http_request_duration_seconds", ConstLabels: prometheus.Labels{"service": service}}, []string{"method", "route"}),
		SQLDuration:   prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "app_sql_query_duration_seconds", ConstLabels: prometheus.Labels{"service": service}}, []string{"operation"}),
		SQLErrors:     prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_sql_errors_total", ConstLabels: prometheus.Labels{"service": service}}, []string{"operation"}),
		WebSockets:    prometheus.NewGauge(prometheus.GaugeOpts{Name: "app_websocket_connections", ConstLabels: prometheus.Labels{"service": service}}),
		ExternalCalls: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_external_http_requests_total", ConstLabels: prometheus.Labels{"service": service}}, []string{"dependency", "method", "result"}),
		ExternalTime:  prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "app_external_http_duration_seconds", ConstLabels: prometheus.Labels{"service": service}}, []string{"dependency", "method"}),
		Fallbacks:     prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_fallbacks_total", ConstLabels: prometheus.Labels{"service": service}}, []string{"dependency", "operation"}),
		KafkaRecords:  prometheus.NewCounter(prometheus.CounterOpts{Name: "app_kafka_records_total", ConstLabels: prometheus.Labels{"service": service}}),
		KafkaErrors:   prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_kafka_errors_total", ConstLabels: prometheus.Labels{"service": service}}, []string{"operation"}),
		KafkaLag:      prometheus.NewGauge(prometheus.GaugeOpts{Name: "app_kafka_record_lag_seconds", ConstLabels: prometheus.Labels{"service": service}}),
	}

	registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}), collectors.NewDBStatsCollector(sqlDB, service), metrics.HTTPRequests, metrics.HTTPDuration, metrics.SQLDuration, metrics.SQLErrors, metrics.WebSockets, metrics.ExternalCalls, metrics.ExternalTime, metrics.Fallbacks, metrics.KafkaRecords, metrics.KafkaErrors, metrics.KafkaLag)

	return metrics
}

func (metrics *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{})
}

func (metrics *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestID := request.Header.Get("X-Request-ID")

		if requestID == "" {
			requestID = uuid.NewString()
		}

		response.Header().Set("X-Request-ID", requestID)

		started := time.Now()

		writer := &statusWriter{ResponseWriter: response, status: http.StatusOK}
		next.ServeHTTP(writer, request)

		route := request.Pattern

		if route == "" {
			route = "unmatched"
		}

		metrics.HTTPRequests.WithLabelValues(request.Method, route, strconv.Itoa(writer.status)).Inc()
		metrics.HTTPDuration.WithLabelValues(request.Method, route).Observe(time.Since(started).Seconds())
	})
}

func (metrics *Metrics) InstrumentGORM(db *gorm.DB) error {
	before := func(tx *gorm.DB) { tx.InstanceSet("metrics_started", time.Now()) }
	after := func(operation string) func(*gorm.DB) {
		return func(tx *gorm.DB) {
			value, ok := tx.InstanceGet("metrics_started")

			if ok {
				metrics.SQLDuration.WithLabelValues(operation).Observe(time.Since(value.(time.Time)).Seconds())
			}

			if tx.Error != nil {
				metrics.SQLErrors.WithLabelValues(operation).Inc()
			}
		}
	}

	if err := db.Callback().Create().Before("gorm:create").Register("metrics:before_create", before); err != nil {
		return err
	}

	if err := db.Callback().Create().After("gorm:create").Register("metrics:after_create", after("create")); err != nil {
		return err
	}

	if err := db.Callback().Query().Before("gorm:query").Register("metrics:before_query", before); err != nil {
		return err
	}

	if err := db.Callback().Query().After("gorm:query").Register("metrics:after_query", after("query")); err != nil {
		return err
	}

	if err := db.Callback().Update().Before("gorm:update").Register("metrics:before_update", before); err != nil {
		return err
	}

	if err := db.Callback().Update().After("gorm:update").Register("metrics:after_update", after("update")); err != nil {
		return err
	}

	if err := db.Callback().Delete().Before("gorm:delete").Register("metrics:before_delete", before); err != nil {
		return err
	}

	return db.Callback().Delete().After("gorm:delete").Register("metrics:after_delete", after("delete"))
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (writer *statusWriter) WriteHeader(status int) {
	writer.status = status
	writer.ResponseWriter.WriteHeader(status)
}

func (writer *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return writer.ResponseWriter.(http.Hijacker).Hijack()
}

func (writer *statusWriter) Flush() { writer.ResponseWriter.(http.Flusher).Flush() }
