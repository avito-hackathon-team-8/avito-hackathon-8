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
	Registry       *prometheus.Registry
	JobRuns        *prometheus.CounterVec
	JobDuration    *prometheus.HistogramVec
	JobRunning     *prometheus.GaugeVec
	JobLastSuccess *prometheus.GaugeVec
	JobUsers       *prometheus.CounterVec
	JobRows        *prometheus.CounterVec
	HTTPRequests   *prometheus.CounterVec
	HTTPDuration   *prometheus.HistogramVec
	SQLDuration    *prometheus.HistogramVec
	SQLErrors      *prometheus.CounterVec
}

func New(sqlDB *sql.DB) *Metrics {
	r := prometheus.NewRegistry()
	m := &Metrics{Registry: r, JobRuns: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_job_runs_total", ConstLabels: prometheus.Labels{"service": "daily-tasks-service"}}, []string{"job", "result"}), JobDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "app_job_duration_seconds", ConstLabels: prometheus.Labels{"service": "daily-tasks-service"}}, []string{"job"}), JobRunning: prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "app_job_running", ConstLabels: prometheus.Labels{"service": "daily-tasks-service"}}, []string{"job"}), JobLastSuccess: prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "app_job_last_success_timestamp_seconds", ConstLabels: prometheus.Labels{"service": "daily-tasks-service"}}, []string{"job"}), JobUsers: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_job_users_total", ConstLabels: prometheus.Labels{"service": "daily-tasks-service"}}, []string{"job"}), JobRows: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_job_rows_total", ConstLabels: prometheus.Labels{"service": "daily-tasks-service"}}, []string{"job"}), HTTPRequests: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_http_requests_total", ConstLabels: prometheus.Labels{"service": "daily-tasks-service"}}, []string{"method", "route", "status"}), HTTPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "app_http_request_duration_seconds", ConstLabels: prometheus.Labels{"service": "daily-tasks-service"}}, []string{"method", "route"}), SQLDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "app_sql_query_duration_seconds", ConstLabels: prometheus.Labels{"service": "daily-tasks-service"}}, []string{"operation"}), SQLErrors: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "app_sql_errors_total", ConstLabels: prometheus.Labels{"service": "daily-tasks-service"}}, []string{"operation"})}

	r.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}), collectors.NewDBStatsCollector(sqlDB, "daily-tasks-service"), m.JobRuns, m.JobDuration, m.JobRunning, m.JobLastSuccess, m.JobUsers, m.JobRows, m.HTTPRequests, m.HTTPDuration, m.SQLDuration, m.SQLErrors)

	return m
}

func (m *Metrics) InstrumentGORM(db *gorm.DB) error {
	before := func(tx *gorm.DB) { tx.InstanceSet("metrics_started", time.Now()) }
	after := func(operation string) func(*gorm.DB) {
		return func(tx *gorm.DB) {
			if value, ok := tx.InstanceGet("metrics_started"); ok {
				m.SQLDuration.WithLabelValues(operation).Observe(time.Since(value.(time.Time)).Seconds())
			}

			if tx.Error != nil {
				m.SQLErrors.WithLabelValues(operation).Inc()
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
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")

		if requestID == "" {
			requestID = uuid.NewString()
		}

		w.Header().Set("X-Request-ID", requestID)

		started := time.Now()
		writer := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(writer, r)
		route := r.Pattern

		if route == "" {
			route = "unmatched"
		}

		m.HTTPRequests.WithLabelValues(r.Method, route, strconv.Itoa(writer.status)).Inc()
		m.HTTPDuration.WithLabelValues(r.Method, route).Observe(time.Since(started).Seconds())
	})
}

func (m *Metrics) AddUsers(job string, count int) {
	m.JobUsers.WithLabelValues(job).Add(float64(count))
}

func (m *Metrics) AddRows(job string, count int) { m.JobRows.WithLabelValues(job).Add(float64(count)) }

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (writer *statusWriter) WriteHeader(status int) {
	writer.status = status
	writer.ResponseWriter.WriteHeader(status)
}

func (m *Metrics) Record(name string, at time.Time, err error) {
	result := "success"

	if err != nil {
		result = "failure"
	} else {
		m.JobLastSuccess.WithLabelValues(name).Set(float64(at.Unix()))
	}

	m.JobRuns.WithLabelValues(name, result).Inc()
}

func (m *Metrics) StartJob(name string) func() {
	started := time.Now()

	m.JobRunning.WithLabelValues(name).Inc()

	return func() {
		m.JobRunning.WithLabelValues(name).Dec()
		m.JobDuration.WithLabelValues(name).Observe(time.Since(started).Seconds())
	}
}
