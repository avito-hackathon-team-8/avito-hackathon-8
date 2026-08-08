package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/email"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvalidEmail = errors.New("enter a valid email address")
	ErrInvalidOTP   = errors.New("invalid or expired OTP")
	ErrUnauthorized = errors.New("authentication required")
)

const maxOTPAttempts = 3

var defaultInterests = `{"electronics":0.5,"home":0.5,"fashion":0.5,"auto":0.5}`

type Config struct {
	JWTSecret  string
	SessionTTL time.Duration
	OTPTTL     time.Duration
	OTPLength  int
}

type Service struct {
	db     *gorm.DB
	mailer email.Sender
	config Config
	now    func() time.Time
}

func NewService(db *gorm.DB, mailer email.Sender, config Config) *Service {
	return &Service{db: db, mailer: mailer, config: config, now: time.Now}
}

func (service *Service) RequestOTP(ctx context.Context, rawEmail string) error {
	normalizedEmail, err := normalizeEmail(rawEmail)

	if err != nil {
		return err
	}

	code, err := generateCode(service.config.OTPLength)

	if err != nil {
		return fmt.Errorf("generate OTP: %w", err)
	}

	var otp models.OTP

	err = service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		candidate := models.User{Email: normalizedEmail, Interests: defaultInterests}

		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&candidate).Error; err != nil {
			return fmt.Errorf("find or create user: %w", err)
		}

		var user models.User

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("email = ?", normalizedEmail).First(&user).Error; err != nil {
			return fmt.Errorf("lock user: %w", err)
		}

		if err := tx.Where("user_id = ?", user.ID).Delete(&models.OTP{}).Error; err != nil {
			return fmt.Errorf("delete previous OTP: %w", err)
		}

		otp = models.OTP{
			UserID:    user.ID,
			CodeHash:  service.hashOTP(user.ID, code),
			ExpiresAt: service.now().UTC().Add(service.config.OTPTTL),
		}

		if err := tx.Create(&otp).Error; err != nil {
			return fmt.Errorf("save OTP: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if err := service.mailer.SendOTP(normalizedEmail, code); err != nil {
		_ = service.db.WithContext(ctx).Delete(&otp).Error

		return fmt.Errorf("send OTP email: %w", err)
	}

	return nil
}

func (service *Service) VerifyOTP(ctx context.Context, rawEmail, code string) (models.User, string, error) {
	normalizedEmail, err := normalizeEmail(rawEmail)

	if err != nil {
		return models.User{}, "", ErrInvalidOTP
	}

	var user models.User

	valid := false

	err = service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("email = ?", normalizedEmail).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}

			return err
		}

		var otp models.OTP

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", user.ID).First(&otp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}

			return err
		}

		expectedHash := service.hashOTP(user.ID, code)
		expired := !service.now().UTC().Before(otp.ExpiresAt)
		matches := len(code) == service.config.OTPLength && hmac.Equal([]byte(expectedHash), []byte(otp.CodeHash))

		if expired {
			return tx.Delete(&otp).Error
		}

		if !matches {
			otp.FailedAttempts++

			if otp.FailedAttempts >= maxOTPAttempts {
				return tx.Delete(&otp).Error
			}

			return tx.Model(&otp).Update("failed_attempts", otp.FailedAttempts).Error
		}

		if err := tx.Delete(&otp).Error; err != nil {
			return err
		}

		valid = true

		if !user.Verified {
			user.Verified = true

			if err := tx.Save(&user).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return models.User{}, "", err
	}

	if !valid {
		return models.User{}, "", ErrInvalidOTP
	}

	token, err := service.createToken(user)

	if err != nil {
		return models.User{}, "", err
	}

	return user, token, nil
}

func (service *Service) Authenticate(ctx context.Context, rawToken string) (models.User, error) {
	tokenString := strings.TrimSpace(rawToken)

	if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
		tokenString = strings.TrimSpace(tokenString[7:])
	}

	if tokenString == "" {
		return models.User{}, ErrUnauthorized
	}

	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(service.config.JWTSecret), nil
	}, jwt.WithIssuer("avito-hackathon-api"), jwt.WithValidMethods([]string{"HS256"}))

	if err != nil || !token.Valid {
		return models.User{}, ErrUnauthorized
	}

	userID, err := uuid.Parse(claims.Subject)

	if err != nil {
		return models.User{}, ErrUnauthorized
	}

	var user models.User

	if err := service.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, ErrUnauthorized
		}

		return models.User{}, err
	}

	return user, nil
}

func (service *Service) createToken(user models.User) (string, error) {
	now := service.now().UTC()

	claims := jwt.RegisteredClaims{
		Issuer:    "avito-hackathon-api",
		Subject:   user.ID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(service.config.SessionTTL)),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(service.config.JWTSecret))
}

func (service *Service) hashOTP(userID uuid.UUID, code string) string {
	digest := hmac.New(sha256.New, []byte(service.config.JWTSecret))
	_, _ = digest.Write([]byte(userID.String() + ":" + code))

	return hex.EncodeToString(digest.Sum(nil))
}

func normalizeEmail(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(normalized)

	if err != nil || address.Address != normalized {
		return "", ErrInvalidEmail
	}

	return normalized, nil
}

func generateCode(length int) (string, error) {
	code := make([]byte, length)

	for index := range code {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))

		if err != nil {
			return "", err
		}

		code[index] = byte('0' + digit.Int64())
	}

	return string(code), nil
}
