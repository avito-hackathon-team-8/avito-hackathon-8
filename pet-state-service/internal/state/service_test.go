package state

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestService(t *testing.T) (*Service, *gorm.DB, uuid.UUID, time.Time) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})

	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&PetState{}, &PetStateAction{}, &OutboxEvent{}, &CareIdempotency{}); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC)
	service := NewService(db, KafkaTopic)
	service.now = func() time.Time { return now }

	return service, db, uuid.New(), now
}

func TestDecayedHappiness(t *testing.T) {
	start := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		elapsed time.Duration
		want    float64
	}{
		{"initial", 0, 100}, {"half", 36 * time.Hour, 50}, {"empty", 72 * time.Hour, 0}, {"clamped", 96 * time.Hour, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := DecayedHappiness(100, start, start.Add(test.elapsed))

			if math.Abs(got-test.want) > 0.001 {
				t.Fatalf("got %.3f, want %.3f", got, test.want)
			}
		})
	}
}

func TestCareAppliesEffectsCooldownsAndOutbox(t *testing.T) {
	service, db, userID, now := newTestService(t)

	if err := db.Create(&PetState{UserID: userID, Happiness: 40, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}

	stroke, err := service.Care(context.Background(), userID, CareStroke, "")

	if err != nil || stroke.Happiness != 60 {
		t.Fatalf("stroke = %+v, err = %v", stroke, err)
	}

	if stroke.StrokeNextAvailableAt == nil || !stroke.StrokeNextAvailableAt.Equal(now.Add(CareCooldown)) {
		t.Fatalf("stroke cooldown = %v", stroke.StrokeNextAvailableAt)
	}

	if _, err := service.Care(context.Background(), userID, CareStroke, ""); !errors.Is(err, ErrCareCooldown) {
		t.Fatalf("second stroke error = %v", err)
	}

	feed, err := service.Care(context.Background(), userID, CareFeed, "")

	if err != nil || feed.Happiness != 95 {
		t.Fatalf("feed = %+v, err = %v", feed, err)
	}

	var outboxCount int64

	if err := db.Model(&OutboxEvent{}).Count(&outboxCount).Error; err != nil || outboxCount != 2 {
		t.Fatalf("outbox count = %d, err = %v", outboxCount, err)
	}
}

func TestCareRejectsFullWithoutCooldown(t *testing.T) {
	service, db, userID, _ := newTestService(t)

	if _, err := service.Get(context.Background(), userID); err != nil {
		t.Fatal(err)
	}

	if _, err := service.Care(context.Background(), userID, CareFeed, ""); !errors.Is(err, ErrHappinessFull) {
		t.Fatalf("error = %v", err)
	}

	var count int64

	if err := db.Model(&PetStateAction{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("action count = %d, err = %v", count, err)
	}
}

func TestCareIdempotencyReturnsOriginalResult(t *testing.T) {
	service, db, userID, now := newTestService(t)
	if err := db.Create(&PetState{UserID: userID, Happiness: 40, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}

	first, err := service.Care(context.Background(), userID, CareStroke, "request-1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Care(context.Background(), userID, CareStroke, "request-1")
	if err != nil || second.Happiness != first.Happiness {
		t.Fatalf("second = %+v, err = %v, want %+v", second, err, first)
	}

	var outboxCount int64
	if err := db.Model(&OutboxEvent{}).Count(&outboxCount).Error; err != nil || outboxCount != 1 {
		t.Fatalf("outbox count = %d, err = %v", outboxCount, err)
	}
	if _, err := service.Care(context.Background(), userID, CareFeed, "request-1"); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("conflicting care error = %v", err)
	}
}
