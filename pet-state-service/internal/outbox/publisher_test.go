package outbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/pet-state-service/internal/state"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPublisherTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})

	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&state.OutboxEvent{}, &state.CareIdempotency{}); err != nil {
		t.Fatal(err)
	}

	return db
}

func TestClaimBatchLeasesEventsAcrossPublishers(t *testing.T) {
	db := newPublisherTestDB(t)
	event := state.OutboxEvent{EventID: uuid.New(), Topic: "events", EventKey: "user", Payload: []byte(`{}`), CreatedAt: time.Now().UTC()}

	if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}

	first := NewPublisher(db, nil)
	second := NewPublisher(db, nil)
	claimed, err := first.claimBatch(context.Background(), 100)

	if err != nil || len(claimed) != 1 {
		t.Fatalf("first claim = %d, err = %v", len(claimed), err)
	}

	claimed, err = second.claimBatch(context.Background(), 100)

	if err != nil || len(claimed) != 0 {
		t.Fatalf("second claim = %d, err = %v", len(claimed), err)
	}
}

func TestPublishFailureReleasesLease(t *testing.T) {
	db := newPublisherTestDB(t)
	event := state.OutboxEvent{EventID: uuid.New(), Topic: "events", EventKey: "user", Payload: []byte(`{}`), CreatedAt: time.Now().UTC()}

	if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}

	publisher := NewPublisher(db, nil)
	publisher.produce = func(context.Context, *kgo.Record) error { return errors.New("kafka unavailable") }

	if err := publisher.publishBatch(context.Background()); err == nil {
		t.Fatal("publishBatch() error = nil")
	}

	var stored state.OutboxEvent

	if err := db.First(&stored, "event_id = ?", event.EventID).Error; err != nil {
		t.Fatal(err)
	}

	if stored.LeaseToken != nil || stored.LeaseUntil != nil || stored.PublishedAt != nil {
		t.Fatalf("event remained claimed: %+v", stored)
	}
}

func TestPublishMarksEventAndDoesNotReplayIt(t *testing.T) {
	db := newPublisherTestDB(t)
	event := state.OutboxEvent{EventID: uuid.New(), Topic: "events", EventKey: "user", Payload: []byte(`{}`), CreatedAt: time.Now().UTC()}

	if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}

	publisher := NewPublisher(db, nil)

	var calls int

	publisher.produce = func(context.Context, *kgo.Record) error { calls++; return nil }

	if err := publisher.publishBatch(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := publisher.publishBatch(context.Background()); err != nil {
		t.Fatal(err)
	}

	if calls != 1 {
		t.Fatalf("produce calls = %d, want 1", calls)
	}
}

func TestCleanupRemovesOnlyExpiredPublishedEvents(t *testing.T) {
	db := newPublisherTestDB(t)
	now := time.Now().UTC()
	old := now.Add(-8 * 24 * time.Hour)
	recent := now.Add(-time.Hour)

	events := []state.OutboxEvent{
		{EventID: uuid.New(), Topic: "events", EventKey: "old", Payload: []byte(`{}`), CreatedAt: old, PublishedAt: &old},
		{EventID: uuid.New(), Topic: "events", EventKey: "recent", Payload: []byte(`{}`), CreatedAt: recent, PublishedAt: &recent},
		{EventID: uuid.New(), Topic: "events", EventKey: "pending", Payload: []byte(`{}`), CreatedAt: old},
	}

	if err := db.Create(&events).Error; err != nil {
		t.Fatal(err)
	}

	publisher := NewPublisher(db, nil)

	if err := publisher.cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}

	var count int64

	if err := db.Model(&state.OutboxEvent{}).Count(&count).Error; err != nil || count != 2 {
		t.Fatalf("remaining count = %d, err = %v", count, err)
	}
}
