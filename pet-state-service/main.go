package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/pet-state-service/internal/config"
	appmetrics "github.com/avito-hackathon-team-8/avito-hackathon-8/pet-state-service/internal/metrics"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/pet-state-service/internal/outbox"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/pet-state-service/internal/state"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type careRequest struct {
	Type state.CareType `json:"type"`
}

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})

	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	sqlDB, err := db.DB()

	if err != nil {
		log.Fatalf("get database connection: %v", err)
	}

	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime)

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)

	if err := sqlDB.PingContext(pingCtx); err != nil {
		pingCancel()

		log.Fatalf("ping database: %v", err)
	}

	pingCancel()

	kafka, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.KafkaBrokers...),
		kgo.RecordDeliveryTimeout(10*time.Second),
	)

	if err != nil {
		log.Fatalf("create kafka client: %v", err)
	}

	startupContext, startupCancel := context.WithTimeout(context.Background(), 30*time.Second)

	if err := ensureTopic(startupContext, kafka, cfg.KafkaTopic); err != nil {
		startupCancel()

		log.Fatalf("ensure kafka topic: %v", err)
	}

	startupCancel()

	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("close database connection: %v", err)
		}
	}()

	defer kafka.Close()

	service := state.NewService(db, cfg.KafkaTopic)
	mux := http.NewServeMux()
	metrics := appmetrics.New(sqlDB)

	if err := metrics.InstrumentGORM(db); err != nil {
		log.Printf("instrument database metrics: %v", err)

		return
	}

	mux.Handle("GET /metrics", metrics.Handler())

	mux.HandleFunc("GET /health/live", liveHandler)
	mux.HandleFunc("GET /health/ready", readyHandler(sqlDB.PingContext, kafka.Ping))
	mux.HandleFunc("GET /health", readyHandler(sqlDB.PingContext, kafka.Ping))

	mux.HandleFunc("GET /internal/v1/users/{userId}/pet-state", authenticated(cfg.InternalServiceToken, func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(r.PathValue("userId"))

		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user id", nil)

			return
		}

		result, err := service.Get(r.Context(), userID)

		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not load pet state", nil)

			return
		}

		writeJSON(w, http.StatusOK, result)
	}))

	mux.HandleFunc("POST /internal/v1/users/{userId}/pet-state/care", authenticated(cfg.InternalServiceToken, func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(r.PathValue("userId"))

		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user id", nil)

			return
		}

		var body careRequest

		if json.NewDecoder(r.Body).Decode(&body) != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", nil)

			return
		}

		result, err := service.Care(r.Context(), userID, body.Type, r.Header.Get("Idempotency-Key"))

		var cooldown *state.CooldownError

		switch {
		case errors.As(err, &cooldown):
			writeError(w, http.StatusConflict, "PET_CARE_COOLDOWN", err.Error(), &cooldown.NextAvailableAt)
		case errors.Is(err, state.ErrHappinessFull):
			writeError(w, http.StatusConflict, "PET_HAPPINESS_FULL", err.Error(), nil)
		case errors.Is(err, state.ErrInvalidCareType):
			writeError(w, http.StatusBadRequest, "INVALID_CARE_TYPE", err.Error(), nil)
		case errors.Is(err, state.ErrInvalidIdempotencyKey):
			writeError(w, http.StatusBadRequest, "INVALID_IDEMPOTENCY_KEY", err.Error(), nil)
		case errors.Is(err, state.ErrIdempotencyConflict):
			writeError(w, http.StatusConflict, "IDEMPOTENCY_KEY_CONFLICT", err.Error(), nil)
		case err != nil:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not care for pet", nil)
		default:
			writeJSON(w, http.StatusOK, result)
		}
	}))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	publisher := outbox.NewPublisher(db, kafka, metrics)
	publisherDone := make(chan struct{})
	go func() {
		defer close(publisherDone)
		publisher.Run(ctx)
	}()

	server := &http.Server{Addr: cfg.HTTPAddress, Handler: metrics.Middleware(mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("serve: %v", err)

			stop()
		}
	}()

	<-ctx.Done()

	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdown); err != nil {
		log.Printf("shutdown pet-state http server: %v", err)
	}
	select {
	case <-publisherDone:
	case <-shutdown.Done():
		log.Printf("service=pet-state job=outbox-shutdown error_type=timeout")
	}
}

func liveHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func readyHandler(dbPing, kafkaPing func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := dbPing(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "dependency": "postgres"})

			return
		}

		if err := kafkaPing(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "dependency": "kafka"})

			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}

func ensureTopic(ctx context.Context, kafka *kgo.Client, topic string) error {
	request := kmsg.NewPtrCreateTopicsRequest()
	request.Topics = []kmsg.CreateTopicsRequestTopic{{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}}

	response, err := request.RequestWith(ctx, kafka)

	if err != nil {
		return err
	}

	for _, result := range response.Topics {
		if result.ErrorCode != 0 && result.ErrorCode != 36 {
			message := ""

			if result.ErrorMessage != nil {
				message = *result.ErrorMessage
			}

			return fmt.Errorf("create %s: kafka error %d: %s", result.Topic, result.ErrorCode, message)
		}
	}

	return nil
}

func authenticated(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := sha256.Sum256([]byte(r.Header.Get("X-Service-Token")))
		b := sha256.Sum256([]byte(token))

		if subtle.ConstantTimeCompare(a[:], b[:]) != 1 {
			writeError(w, http.StatusUnauthorized, "INVALID_SERVICE_TOKEN", "Invalid service token", nil)

			return
		}

		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string, next *time.Time) {
	payload := map[string]any{"code": code, "message": message}

	if next != nil {
		payload["nextAvailableAt"] = next.UTC()
	}

	writeJSON(w, status, payload)
}
