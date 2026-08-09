package auth

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/pet"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type testMailer struct{}

func (testMailer) SendOTP(_, _ string) error { return nil }

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	email, err := normalizeEmail("  USER@example.com ")

	if err != nil {
		t.Fatalf("normalizeEmail returned an error: %v", err)
	}

	if email != "user@example.com" {
		t.Fatalf("normalizeEmail = %q, want user@example.com", email)
	}
}

func TestNormalizeEmailRejectsDisplayName(t *testing.T) {
	t.Parallel()

	if _, err := normalizeEmail("User <user@example.com>"); err == nil {
		t.Fatal("normalizeEmail accepted a display name")
	}
}

func TestGenerateCode(t *testing.T) {
	t.Parallel()

	code, err := generateCode(8)

	if err != nil {
		t.Fatalf("generateCode returned an error: %v", err)
	}

	if !regexp.MustCompile(`^[0-9]{8}$`).MatchString(code) {
		t.Fatalf("generateCode = %q, want 8 digits", code)
	}
}

func TestRequestOTPCreatesInitialPet(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Pet{}, &models.OTP{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	service := NewService(db, testMailer{}, Config{
		JWTSecret:  "test-secret",
		OTPTTL:     time.Minute,
		OTPLength:  8,
		SessionTTL: time.Hour,
	})
	if err := service.RequestOTP(context.Background(), "user@example.com"); err != nil {
		t.Fatalf("RequestOTP() error = %v", err)
	}

	var user models.User
	if err := db.Where("email = ?", "user@example.com").First(&user).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	var initialPet models.Pet
	if err := db.Where("user_id = ?", user.ID).First(&initialPet).Error; err != nil {
		t.Fatalf("load initial pet: %v", err)
	}
	if initialPet.Level != pet.InitialPetLevel || initialPet.Leaves != pet.InitialPetLeaves {
		t.Fatalf("initial pet = %+v, want level %d with %d leaves", initialPet, pet.InitialPetLevel, pet.InitialPetLeaves)
	}
}
