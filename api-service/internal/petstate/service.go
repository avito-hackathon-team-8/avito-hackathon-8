package petstate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/httpx"
	appmetrics "github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/metrics"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type CareType string

const (
	CareStroke CareType = "STROKE"
	CareFeed   CareType = "FEED"
)

var (
	ErrUnavailable           = errors.New("pet state service is unavailable")
	ErrCooldown              = errors.New("pet care action is on cooldown")
	ErrFull                  = errors.New("pet happiness is already full")
	ErrInvalidType           = errors.New("invalid pet care type")
	ErrIdempotencyConflict   = errors.New("idempotency key was already used for another operation")
	ErrInvalidIdempotencyKey = errors.New("idempotency key is invalid")
)

type Snapshot struct {
	Happiness             float64    `json:"happiness"`
	CalculatedAt          time.Time  `json:"calculatedAt"`
	DecaysToZeroAt        time.Time  `json:"decaysToZeroAt"`
	StrokeNextAvailableAt *time.Time `json:"strokeNextAvailableAt"`
	FeedNextAvailableAt   *time.Time `json:"feedNextAvailableAt"`
	HappinessMultiplier   float64    `json:"happinessMultiplier"`
}

type Update struct {
	UserID     uuid.UUID
	OccurredAt time.Time
	State      Snapshot
}

type remoteError struct {
	Code            string     `json:"code"`
	Message         string     `json:"message"`
	NextAvailableAt *time.Time `json:"nextAvailableAt"`
}

type CooldownError struct{ NextAvailableAt time.Time }

func (err *CooldownError) Error() string { return ErrCooldown.Error() }
func (err *CooldownError) Unwrap() error { return ErrCooldown }

type Service struct {
	baseURL    string
	token      string
	httpClient *http.Client
	kafka      *kgo.Client
	metrics    *appmetrics.Metrics

	subscribersMu sync.Mutex
	subscribers   map[uuid.UUID]map[chan Update]struct{}
}

func NewService(baseURL, token string, brokers []string, topic, instanceID string, metrics *appmetrics.Metrics) (*Service, error) {
	instanceID = sanitizeGroupPart(instanceID)

	kafka, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topic),
		kgo.ConsumerGroup("api-service-pet-state-ws-v1-"+instanceID),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
	)

	if err != nil {
		return nil, fmt.Errorf("create pet state kafka consumer: %w", err)
	}

	return &Service{
		baseURL: strings.TrimRight(baseURL, "/"), token: token,
		httpClient: &http.Client{Timeout: 5 * time.Second}, kafka: kafka,
		metrics:     metrics,
		subscribers: make(map[uuid.UUID]map[chan Update]struct{}),
	}, nil
}

func sanitizeGroupPart(value string) string {
	value = strings.TrimSpace(value)

	if value == "" {
		return "local"
	}

	return strings.NewReplacer(" ", "-", "/", "-", "\\", "-").Replace(value)
}

func (service *Service) Close() { service.kafka.Close() }

func (service *Service) Get(ctx context.Context, userID uuid.UUID) (Snapshot, error) {
	return service.request(ctx, http.MethodGet, userID, nil)
}

func (service *Service) Care(ctx context.Context, userID uuid.UUID, careType CareType, idempotencyKey string) (Snapshot, error) {
	if careType != CareStroke && careType != CareFeed {
		return Snapshot{}, ErrInvalidType
	}

	return service.request(ctx, http.MethodPost, userID, map[string]CareType{"type": careType}, idempotencyKey)
}

func (service *Service) request(ctx context.Context, method string, userID uuid.UUID, body any, idempotencyKey ...string) (Snapshot, error) {
	var reader io.Reader

	if body != nil {
		payload, err := json.Marshal(body)

		if err != nil {
			return Snapshot{}, err
		}

		reader = bytes.NewReader(payload)
	}

	path := fmt.Sprintf("%s/internal/v1/users/%s/pet-state", service.baseURL, userID)

	if method == http.MethodPost {
		path += "/care"
	}

	requestCtx, cancel := httpx.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, method, path, reader)

	if err != nil {
		return Snapshot{}, fmt.Errorf("create pet state request: %w", err)
	}

	request.Header.Set("X-Service-Token", service.token)
	request.Header.Set("Content-Type", "application/json")

	if len(idempotencyKey) > 0 && idempotencyKey[0] != "" {
		request.Header.Set("Idempotency-Key", idempotencyKey[0])
	}

	started := time.Now()
	response, err := service.httpClient.Do(request)
	resultLabel := "success"

	if err != nil {
		resultLabel = "error"

		if errors.Is(err, context.DeadlineExceeded) {
			resultLabel = "timeout"
		}
	} else if response.StatusCode >= 500 {
		resultLabel = "server_error"
	}

	if service.metrics != nil {
		service.metrics.ExternalCalls.WithLabelValues("pet-state", method, resultLabel).Inc()
		service.metrics.ExternalTime.WithLabelValues("pet-state", method).Observe(time.Since(started).Seconds())
	}

	if err != nil {
		return Snapshot{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("close pet state response body: %v", err)
		}
	}()

	if response.StatusCode != http.StatusOK {
		var remote remoteError

		_ = json.NewDecoder(response.Body).Decode(&remote)

		switch remote.Code {
		case "PET_CARE_COOLDOWN":
			if remote.NextAvailableAt != nil {
				return Snapshot{}, &CooldownError{NextAvailableAt: *remote.NextAvailableAt}
			}

			return Snapshot{}, ErrCooldown
		case "PET_HAPPINESS_FULL":
			return Snapshot{}, ErrFull
		case "INVALID_CARE_TYPE":
			return Snapshot{}, ErrInvalidType
		case "IDEMPOTENCY_KEY_CONFLICT":
			return Snapshot{}, ErrIdempotencyConflict
		case "INVALID_IDEMPOTENCY_KEY":
			return Snapshot{}, ErrInvalidIdempotencyKey
		default:
			return Snapshot{}, fmt.Errorf("%w: status %d", ErrUnavailable, response.StatusCode)
		}
	}

	var result Snapshot

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return Snapshot{}, fmt.Errorf("%w: decode response: %w", ErrUnavailable, err)
	}

	return result, nil
}

func (service *Service) Ready(ctx context.Context) error {
	requestCtx, cancel := httpx.WithTimeout(ctx, 750*time.Millisecond)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, service.baseURL+"/health/ready", nil)

	if err != nil {
		return err
	}

	response, err := service.httpClient.Do(request)

	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("close pet state readiness response body: %v", err)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: readiness status %d", ErrUnavailable, response.StatusCode)
	}

	return nil
}

func (service *Service) Subscribe(userID uuid.UUID) (<-chan Update, func()) {
	updates := make(chan Update, 8)

	service.subscribersMu.Lock()

	if service.subscribers[userID] == nil {
		service.subscribers[userID] = make(map[chan Update]struct{})
	}

	service.subscribers[userID][updates] = struct{}{}
	service.subscribersMu.Unlock()

	var once sync.Once

	return updates, func() {
		once.Do(func() {
			service.subscribersMu.Lock()

			delete(service.subscribers[userID], updates)

			if len(service.subscribers[userID]) == 0 {
				delete(service.subscribers, userID)
			}

			close(updates)

			service.subscribersMu.Unlock()
		})
	}
}

func (service *Service) Run(ctx context.Context) {
	seen := make(map[uuid.UUID]struct{}, 1024)
	order := make([]uuid.UUID, 0, 1024)

	for ctx.Err() == nil {
		fetches := service.kafka.PollFetches(ctx)

		if fetches.IsClientClosed() || ctx.Err() != nil {
			return
		}

		fetches.EachRecord(func(record *kgo.Record) {
			if service.metrics != nil {
				service.metrics.KafkaRecords.Inc()

				if !record.Timestamp.IsZero() {
					lag := time.Since(record.Timestamp).Seconds()

					if lag >= 0 {
						service.metrics.KafkaLag.Set(lag)
					}
				}
			}

			var event struct {
				EventID    uuid.UUID `json:"eventId"`
				Type       string    `json:"type"`
				UserID     uuid.UUID `json:"userId"`
				OccurredAt time.Time `json:"occurredAt"`
				Data       Snapshot  `json:"data"`
			}

			if err := json.Unmarshal(record.Value, &event); err != nil {
				if service.metrics != nil {
					service.metrics.KafkaErrors.WithLabelValues("decode").Inc()
				}

				log.Printf("decode pet state event topic=%s partition=%d offset=%d: %v", record.Topic, record.Partition, record.Offset, err)

				return
			}

			if event.Type != "PET_STATE_CHANGED" {
				log.Printf("ignore pet state event with unexpected type %q", event.Type)

				return
			}

			if event.EventID != uuid.Nil {
				if _, duplicate := seen[event.EventID]; duplicate {
					return
				}

				seen[event.EventID] = struct{}{}
				order = append(order, event.EventID)

				if len(order) > 4096 {
					delete(seen, order[0])
					order = order[1:]
				}
			}

			service.publish(Update{UserID: event.UserID, OccurredAt: event.OccurredAt, State: event.Data})
		})

		if err := service.kafka.CommitUncommittedOffsets(ctx); err != nil && ctx.Err() == nil {
			if service.metrics != nil {
				service.metrics.KafkaErrors.WithLabelValues("commit").Inc()
			}

			log.Printf("commit pet state kafka offsets: %v", err)

			timer := time.NewTimer(time.Second)

			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}
}

func (service *Service) publish(update Update) {
	service.subscribersMu.Lock()
	defer service.subscribersMu.Unlock()

	for channel := range service.subscribers[update.UserID] {
		select {
		case channel <- update:
		default:
			draining := true
			for draining {
				select {
				case <-channel:
				default:
					draining = false
				}
			}
			select {
			case channel <- update:
				log.Printf("replace buffered pet state websocket updates user_id=%s", update.UserID)
			default:
				log.Printf("drop pet state websocket update after buffer replacement user_id=%s", update.UserID)
			}
		}
	}
}
