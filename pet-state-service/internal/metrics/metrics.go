package metrics

import (
	"database/sql"
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
	Registry        *prometheus.Registry
	HTTPRequests    *prometheus.CounterVec
	HTTPDuration    *prometheus.HistogramVec
	OutboxPublished *prometheus.CounterVec
	OutboxPending   prometheus.Gauge
	OutboxOldest    prometheus.Gauge
	SQLDuration     *prometheus.HistogramVec
	SQLErrors       *prometheus.CounterVec
}

func New(sqlDB *sql.DB) *Metrics {
	registry := prometheus.NewRegistry()

	metrics := &Metrics{
		Registry:        registry,
		HTTPRequests:    prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_http_requests_total", ConstLabels: prometheus.Labels{"service": "pet-state-service"}}, []string{"method", "route", "status"}),
		HTTPDuration:    prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "app_http_request_duration_seconds", ConstLabels: prometheus.Labels{"service": "pet-state-service"}}, []string{"method", "route"}),
		OutboxPublished: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_outbox_publish_total", ConstLabels: prometheus.Labels{"service": "pet-state-service"}}, []string{"result"}),
		OutboxPending:   prometheus.NewGauge(prometheus.GaugeOpts{Name: "app_outbox_pending", ConstLabels: prometheus.Labels{"service": "pet-state-service"}}),
		OutboxOldest:    prometheus.NewGauge(prometheus.GaugeOpts{Name: "app_outbox_oldest_age_seconds", ConstLabels: prometheus.Labels{"service": "pet-state-service"}}),
		SQLDuration:     prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "app_sql_query_duration_seconds", ConstLabels: prometheus.Labels{"service": "pet-state-service"}}, []string{"operation"}),
		SQLErrors:       prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_sql_errors_total", ConstLabels: prometheus.Labels{"service": "pet-state-service"}}, []string{"operation"}),
	}

	registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}), collectors.NewDBStatsCollector(sqlDB, "pet_state"), metrics.HTTPRequests, metrics.HTTPDuration, metrics.OutboxPublished, metrics.OutboxPending, metrics.OutboxOldest, metrics.SQLDuration, metrics.SQLErrors)

	return metrics
}

func (metrics *Metrics) InstrumentGORM(db *gorm.DB) error {
	before := func(tx *gorm.DB) { tx.InstanceSet("metrics_started", time.Now()) }
	after := func(operation string) func(*gorm.DB) {
		return func(tx *gorm.DB) {
			if value, ok := tx.InstanceGet("metrics_started"); ok {
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

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (writer *statusWriter) WriteHeader(status int) {
	writer.status = status
	writer.ResponseWriter.WriteHeader(status)
}
