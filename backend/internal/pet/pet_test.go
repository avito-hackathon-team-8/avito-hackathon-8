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

func TestApplyLevelUps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level      int
		leaves     int64
		wantLevel  int
		wantLeaves int64
	}{
		{level: 1, leaves: 0, wantLevel: 1, wantLeaves: 0},
		{level: 1, leaves: 99, wantLevel: 1, wantLeaves: 99},
		{level: 1, leaves: 100, wantLevel: 2, wantLeaves: 0},
		{level: 1, leaves: 229, wantLevel: 2, wantLeaves: 129},
		{level: 1, leaves: 390, wantLevel: 4, wantLeaves: 0},
		{level: 4, leaves: 300, wantLevel: 5, wantLeaves: 110},
		{level: 10, leaves: 10000, wantLevel: 10, wantLeaves: 10000},
	}

	for _, test := range tests {
		gotLevel, gotLeaves := applyLevelUps(test.level, test.leaves)
		if gotLevel != test.wantLevel || gotLeaves != test.wantLeaves {
			t.Errorf("applyLevelUps(%d, %d) = (%d, %d), want (%d, %d)", test.level, test.leaves, gotLevel, gotLeaves, test.wantLevel, test.wantLeaves)
		}
	}
}

func TestAddLeavesCrossesLevelsInOneTransaction(t *testing.T) {
	service, user := testService(t)
	pet, err := service.GetOrCreate(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	result, err := service.AddLeaves(context.Background(), user.ID, 390)
	if err != nil {
		t.Fatalf("AddLeaves() error = %v", err)
	}

	if result.Level != 4 || result.Leaves != 0 || !result.LevelUp {
		t.Fatalf("AddLeaves() progress = %+v, want level 4, 0 leaves, and level up", result)
	}

	stored, err := service.GetOrCreate(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetOrCreate() after AddLeaves() error = %v", err)
	}

	if stored.ID != pet.ID || stored.Level != 4 || stored.Leaves != 0 {
		t.Fatalf("stored pet = %+v, want level 4 and 0 leaves", stored)
	}
}

func TestAddLeavesRejectsInvalidAmountAndOverflow(t *testing.T) {
	service, user := testService(t)
	if _, err := service.GetOrCreate(context.Background(), user.ID); err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	if _, err := service.AddLeaves(context.Background(), user.ID, 0); !errors.Is(err, ErrInvalidLeaves) {
		t.Fatalf("AddLeaves(0) error = %v, want ErrInvalidLeaves", err)
	}

	if err := service.db.Model(&models.Pet{}).Where("user_id = ?", user.ID).Updates(map[string]any{
		"leaves": maxInt64,
		"level":  MaxPetLevel,
	}).Error; err != nil {
		t.Fatalf("seed max leaves: %v", err)
	}

	if _, err := service.AddLeaves(context.Background(), user.ID, 1); !errors.Is(err, ErrLeavesOverflow) {
		t.Fatalf("AddLeaves() error = %v, want ErrLeavesOverflow", err)
	}
}

func TestUpdatesContainCurrentPetProgress(t *testing.T) {
	service, user := testService(t)
	if _, err := service.GetOrCreate(context.Background(), user.ID); err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	updates, unsubscribe := service.Subscribe(user.ID)
	defer unsubscribe()

	if _, err := service.AddLeaves(context.Background(), user.ID, 100); err != nil {
		t.Fatalf("AddLeaves() error = %v", err)
	}

	update := <-updates
	if update.Progress.Name != "" || update.Progress.Level != 2 || update.Progress.Leaves != 0 ||
		update.Progress.NextLevelTargetLeaves != 130 || !update.Progress.LevelUp {
		t.Fatalf("update = %+v, want empty name, level 2, 0 leaves, target 130, and level up", update.Progress)
	}

	if _, err := service.UpdateName(context.Background(), user.ID, "Листик"); err != nil {
		t.Fatalf("UpdateName() error = %v", err)
	}
}

func TestUpdateNameTrimsAndValidates(t *testing.T) {
	service, user := testService(t)
	if _, err := service.GetOrCreate(context.Background(), user.ID); err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	pet, err := service.UpdateName(context.Background(), user.ID, "  Листик  ")
	if err != nil {
		t.Fatalf("UpdateName() error = %v", err)
	}

	if pet.Name != "Листик" {
		t.Fatalf("pet name = %q, want Листик", pet.Name)
	}

	if _, err := service.UpdateName(context.Background(), user.ID, "   "); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("UpdateName(empty) error = %v, want ErrInvalidName", err)
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

	service := NewService(db)

	return service, user
}
