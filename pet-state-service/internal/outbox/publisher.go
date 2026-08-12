package outbox

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	appmetrics "github.com/avito-hackathon-team-8/avito-hackathon-8/pet-state-service/internal/metrics"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/pet-state-service/internal/state"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Publisher struct {
	db            *gorm.DB
	period        time.Duration
	leaseDuration time.Duration
	retention     time.Duration
	cleanupPeriod time.Duration
	nextCleanup   time.Time
	nextBacklog   time.Time
	produce       func(context.Context, *kgo.Record) error
	metrics       *appmetrics.Metrics
}

func NewPublisher(db *gorm.DB, kafka *kgo.Client, observed ...*appmetrics.Metrics) *Publisher {
	var metrics *appmetrics.Metrics

	if len(observed) > 0 {
		metrics = observed[0]
	}

	return &Publisher{
		db: db, period: 500 * time.Millisecond,
		metrics:       metrics,
		leaseDuration: 30 * time.Second, retention: 7 * 24 * time.Hour, cleanupPeriod: time.Hour,
		produce: func(ctx context.Context, record *kgo.Record) error {
			return kafka.ProduceSync(ctx, record).FirstErr()
		},
	}
}

func (publisher *Publisher) Run(ctx context.Context) {
	ticker := time.NewTicker(publisher.period)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return
		}

		if err := publisher.publishBatch(ctx); err != nil && ctx.Err() == nil {
			log.Printf("service=pet-state job=outbox-publisher error_type=publish_failed: %v", err)
		}

		if ctx.Err() != nil {
			return
		}

		if publisher.nextCleanup.IsZero() || time.Now().After(publisher.nextCleanup) {
			if err := publisher.cleanup(ctx); err != nil && ctx.Err() == nil {
				log.Printf("service=pet-state job=outbox-cleanup error_type=cleanup_failed: %v", err)
			}

			publisher.nextCleanup = time.Now().Add(publisher.cleanupPeriod)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (publisher *Publisher) publishBatch(ctx context.Context) error {
	if publisher.metrics != nil && (publisher.nextBacklog.IsZero() || time.Now().After(publisher.nextBacklog)) {
		publisher.observeBacklog(ctx)
		publisher.nextBacklog = time.Now().Add(10 * time.Second)
	}

	events, err := publisher.claimBatch(ctx, 100)

	if err != nil {
		return err
	}

	if len(events) == 100 {
		log.Printf("pet state outbox backlog has at least %d ready events", len(events))
	}

	if len(events) == 0 {
		return nil
	}

	if err := publisher.renewLease(ctx, events); err != nil {
		return err
	}

	for _, event := range events {
		if err := publisher.produce(ctx, &kgo.Record{Topic: event.Topic, Key: []byte(event.EventKey), Value: event.Payload}); err != nil {
			if publisher.metrics != nil {
				publisher.metrics.OutboxPublished.WithLabelValues("failure").Inc()
			}

			releaseCtx := ctx
			releaseCancel := func() {}

			if ctx.Err() != nil {
				releaseCtx, releaseCancel = context.WithTimeout(context.Background(), time.Second)
			}

			for _, claimed := range events {
				publisher.releaseLease(releaseCtx, claimed)
			}

			releaseCancel()

			return fmt.Errorf("produce %s: %w", event.EventID, err)
		}

		now := time.Now().UTC()

		result := publisher.db.WithContext(ctx).Model(&state.OutboxEvent{}).
			Where("event_id = ? AND published_at IS NULL AND lease_token = ?", event.EventID, event.LeaseToken).
			Updates(map[string]any{"published_at": now, "lease_token": nil, "lease_until": nil})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected != 1 {
			return fmt.Errorf("lost lease for outbox event %s after publication", event.EventID)
		}

		if publisher.metrics != nil {
			publisher.metrics.OutboxPublished.WithLabelValues("success").Inc()
		}
	}

	return nil
}

func (publisher *Publisher) observeBacklog(ctx context.Context) {
	if publisher.metrics == nil {
		return
	}

	var stats struct {
		Count  int64
		Oldest *time.Time
	}

	if err := publisher.db.WithContext(ctx).Model(&state.OutboxEvent{}).Select("COUNT(*) AS count, MIN(created_at) AS oldest").Where("published_at IS NULL").Scan(&stats).Error; err != nil {
		return
	}

	publisher.metrics.OutboxPending.Set(float64(stats.Count))

	age := 0.0

	if stats.Oldest != nil {
		age = time.Since(*stats.Oldest).Seconds()

		if age < 0 {
			age = 0
		}
	}

	publisher.metrics.OutboxOldest.Set(age)
}

func (publisher *Publisher) renewLease(ctx context.Context, events []state.OutboxEvent) error {
	if len(events) == 0 || events[0].LeaseToken == nil {
		return errors.New("outbox lease token is missing")
	}

	ids := make([]uuid.UUID, 0, len(events))

	for _, event := range events {
		ids = append(ids, event.EventID)
	}

	result := publisher.db.WithContext(ctx).Model(&state.OutboxEvent{}).
		Where("event_id IN ? AND published_at IS NULL AND lease_token = ?", ids, events[0].LeaseToken).
		Update("lease_until", time.Now().UTC().Add(publisher.leaseDuration))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected != int64(len(events)) {
		return fmt.Errorf("lost outbox lease: renewed %d of %d events", result.RowsAffected, len(events))
	}

	return nil
}

func (publisher *Publisher) claimBatch(ctx context.Context, limit int) ([]state.OutboxEvent, error) {
	now := time.Now().UTC()
	leaseToken := uuid.New()

	var events []state.OutboxEvent

	err := publisher.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("published_at IS NULL AND (lease_until IS NULL OR lease_until < ?)", now).
			Order("created_at").Limit(limit).Find(&events).Error; err != nil {
			return err
		}

		if len(events) == 0 {
			return nil
		}

		ids := make([]uuid.UUID, 0, len(events))
		leaseUntil := now.Add(publisher.leaseDuration)

		for index := range events {
			ids = append(ids, events[index].EventID)
			events[index].LeaseToken = &leaseToken
			events[index].LeaseUntil = &leaseUntil
		}

		return tx.Model(&state.OutboxEvent{}).Where("event_id IN ?", ids).
			Updates(map[string]any{"lease_token": leaseToken, "lease_until": leaseUntil}).Error
	})

	return events, err
}

func (publisher *Publisher) releaseLease(ctx context.Context, event state.OutboxEvent) {
	if event.LeaseToken == nil {
		return
	}

	if err := publisher.db.WithContext(ctx).Model(&state.OutboxEvent{}).
		Where("event_id = ? AND published_at IS NULL AND lease_token = ?", event.EventID, event.LeaseToken).
		Updates(map[string]any{"lease_token": nil, "lease_until": nil}).Error; err != nil && ctx.Err() == nil {
		log.Printf("release outbox lease event_id=%s: %v", event.EventID, err)
	}
}

func (publisher *Publisher) cleanup(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-publisher.retention)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var ids []uuid.UUID

		if err := publisher.db.WithContext(ctx).Model(&state.OutboxEvent{}).
			Where("published_at IS NOT NULL AND published_at < ?", cutoff).
			Order("published_at").Limit(500).Pluck("event_id", &ids).Error; err != nil {
			return err
		}

		if len(ids) == 0 {
			break
		}

		if err := publisher.db.WithContext(ctx).Where("event_id IN ?", ids).Delete(&state.OutboxEvent{}).Error; err != nil {
			return err
		}
	}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var ids []struct {
			UserID         uuid.UUID
			IdempotencyKey string
		}

		if err := publisher.db.WithContext(ctx).Model(&state.CareIdempotency{}).
			Where("created_at < ?", cutoff).Order("created_at").Limit(500).
			Select("user_id, idempotency_key").Scan(&ids).Error; err != nil {
			return err
		}

		if len(ids) == 0 {
			return nil
		}

		for _, id := range ids {
			if err := publisher.db.WithContext(ctx).Where("user_id = ? AND idempotency_key = ?", id.UserID, id.IdempotencyKey).Delete(&state.CareIdempotency{}).Error; err != nil {
				return err
			}
		}
	}
}
