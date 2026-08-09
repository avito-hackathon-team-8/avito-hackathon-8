package pet

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGetOrCreateAndUpdateName(t *testing.T) {
	service, user := testService(t)

	created, err := service.GetOrCreate(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}
	if created.UserID != user.ID || created.Level != InitialPetLevel || created.Leaves != InitialPetLeaves {
		t.Fatalf("created pet = %+v, want level %d with %d leaves", created, InitialPetLevel, InitialPetLeaves)
	}

	updated, err := service.UpdateName(context.Background(), user.ID, "  Листик  ")
	if err != nil {
		t.Fatalf("UpdateName() error = %v", err)
	}
	if updated.Name != "Листик" {
		t.Fatalf("pet name = %q, want Листик", updated.Name)
	}
	if _, err := service.UpdateName(context.Background(), user.ID, "   "); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("UpdateName(empty) error = %v, want ErrInvalidName", err)
	}
}

func TestGetOrCreateRejectsNilUser(t *testing.T) {
	service, _ := testService(t)
	if _, err := service.GetOrCreate(context.Background(), uuid.Nil); !errors.Is(err, ErrPetNotFound) {
		t.Fatalf("GetOrCreate(nil) error = %v, want ErrPetNotFound", err)
	}
}

func TestPublishProgressNotifiesSubscriber(t *testing.T) {
	service, user := testService(t)
	updates, unsubscribe := service.Subscribe(user.ID)
	defer unsubscribe()

	want := Progress{Name: "Листик", Level: 2, Leaves: 15, NextLevelTargetLeaves: 130, LevelUp: true}
	service.PublishProgress(user.ID, want)

	if got := (<-updates).Progress; got != want {
		t.Fatalf("published progress = %+v, want %+v", got, want)
	}
}

func testService(t *testing.T) (*Service, models.User) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Pet{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	user := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString())}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return NewService(db), user
}
