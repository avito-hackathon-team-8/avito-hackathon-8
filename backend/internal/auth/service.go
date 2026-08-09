package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidEmail = errors.New("enter a valid email address")
	ErrInvalidOTP   = errors.New("invalid or expired OTP")
	ErrUnauthorized = errors.New("authentication required")
)

var defaultInterests = `{"electronics":0.5,"home":0.5,"fashion":0.5,"auto":0.5}`

type Config struct {
	JWTSecret  string
	SessionTTL time.Duration
}

type Service struct {
	db     *gorm.DB
	config Config
	now    func() time.Time
}

func NewService(db *gorm.DB, config Config) *Service {
	return &Service{db: db, config: config, now: time.Now}
}

func (service *Service) RequestOTP(ctx context.Context, rawEmail string) error {
	normalizedEmail, err := normalizeEmail(rawEmail)

	if err != nil {
		return err
	}

	return service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		candidate := models.User{Email: normalizedEmail, Interests: defaultInterests}

		if err := tx.Where("email = ?", normalizedEmail).FirstOrCreate(&candidate).Error; err != nil {
			return err
		}

		return nil
	})
}

func (service *Service) VerifyOTP(ctx context.Context, rawEmail, code string) (models.User, string, error) {
	normalizedEmail, err := normalizeEmail(rawEmail)

	if err != nil {
		return models.User{}, "", ErrInvalidOTP
	}

	if strings.TrimSpace(code) == "" {
		return models.User{}, "", ErrInvalidOTP
	}

	var user models.User

	err = service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("email = ?", normalizedEmail).First(&user).Error; err != nil {
			return err
		}

		if !user.Verified {
			user.Verified = true

			if err := tx.Save(&user).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, "", ErrInvalidOTP
		}

		return models.User{}, "", err
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

func normalizeEmail(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(normalized)

	if err != nil || address.Address != normalized {
		return "", ErrInvalidEmail
	}

	return normalized, nil
}
