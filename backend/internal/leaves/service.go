package leaves

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	MinPetLevel = 1
	MaxPetLevel = 10
	maxInt64    = int64(1<<63 - 1)
)

var (
	ErrPetNotFound        = errors.New("pet not found")
	ErrInvalidAmount      = errors.New("leaves amount must be positive")
	ErrInvalidOperation   = errors.New("leaf operation is invalid")
	ErrDuplicateOperation = errors.New("leaf operation already applied")
	ErrLeavesOverflow     = errors.New("leaves amount overflows int64")
)

var levelTargets = [...]int64{0, 0, 100, 230, 390, 580, 810, 1090, 1430, 1850, 2400}

type Credit struct {
	UserID       uuid.UUID
	Amount       int64
	Reason       models.LeafTransactionReason
	OperationKey string
	OccurredAt   time.Time
}

type Progress struct {
	Name                  string
	Level                 int
	Leaves                int64
	NextLevelTargetLeaves int64
	ChestPrice            int64
	LevelUp               bool
}

type DailyReportNotifier interface {
	Notify(userID uuid.UUID)
}

type Service struct {
	db          *gorm.DB
	now         func() time.Time
	dailyReport DailyReportNotifier
}

func NewService(db *gorm.DB, dailyReport DailyReportNotifier) *Service {
	return &Service{db: db, now: time.Now, dailyReport: dailyReport}
}

func (service *Service) Credit(ctx context.Context, credit Credit) (Progress, error) {
	var progress Progress

	err := service.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error

		progress, err = service.CreditTx(tx, credit)

		return err
	})

	if err != nil {
		return Progress{}, err
	}

	service.dailyReport.Notify(credit.UserID)

	return progress, nil
}

func (service *Service) CreditTx(tx *gorm.DB, credit Credit) (Progress, error) {
	if credit.Amount <= 0 {
		return Progress{}, ErrInvalidAmount
	}

	credit.OperationKey = strings.TrimSpace(credit.OperationKey)

	if credit.UserID == uuid.Nil || credit.OperationKey == "" || len(credit.OperationKey) > 160 || !validCreditReason(credit.Reason) {
		return Progress{}, ErrInvalidOperation
	}

	if credit.OccurredAt.IsZero() {
		credit.OccurredAt = service.now().UTC()
	} else {
		credit.OccurredAt = credit.OccurredAt.UTC()
	}

	result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "operation_key"}}, DoNothing: true}).Create(&models.LeafTransaction{
		UserID: credit.UserID, Amount: credit.Amount, Reason: credit.Reason,
		OperationKey: credit.OperationKey, OccurredAt: credit.OccurredAt,
	})

	if result.Error != nil {
		return Progress{}, fmt.Errorf("record leaf credit: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return Progress{}, ErrDuplicateOperation
	}

	if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}}, DoNothing: true}).Create(&models.Pet{UserID: credit.UserID, Level: MinPetLevel}).Error; err != nil {
		return Progress{}, fmt.Errorf("ensure pet: %w", err)
	}

	var pet models.Pet

	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", credit.UserID).First(&pet).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Progress{}, ErrPetNotFound
	}

	if err != nil {
		return Progress{}, fmt.Errorf("lock pet: %w", err)
	}

	if pet.Leaves < 0 || pet.Leaves > maxInt64-credit.Amount {
		return Progress{}, ErrLeavesOverflow
	}

	oldLevel := pet.Level
	newLevel, remainingLeaves := ApplyLevelUps(pet.Level, pet.Leaves+credit.Amount)
	pet.Level = newLevel
	pet.Leaves = remainingLeaves

	if err := tx.Model(&pet).Updates(map[string]any{"level": pet.Level, "leaves": pet.Leaves}).Error; err != nil {
		return Progress{}, fmt.Errorf("save pet progress: %w", err)
	}

	for level := oldLevel; level < newLevel; level++ {
		cost := LevelCost(level)

		if err := tx.Create(&models.LeafTransaction{
			UserID: credit.UserID, Amount: -cost, Reason: models.LeafReasonLevelUp,
			OperationKey: fmt.Sprintf("%s:level:%d", credit.OperationKey, level+1), OccurredAt: credit.OccurredAt,
		}).Error; err != nil {
			return Progress{}, fmt.Errorf("record level spending: %w", err)
		}
	}

	if err := tx.Exec(`
		INSERT INTO user_game_states (user_id, pet_level, leaf_balance, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET pet_level = EXCLUDED.pet_level,
		leaf_balance = EXCLUDED.leaf_balance, updated_at = EXCLUDED.updated_at`,
		credit.UserID, pet.Level, pet.Leaves, service.now().UTC()).Error; err != nil {
		return Progress{}, fmt.Errorf("sync game state: %w", err)
	}

	return ProgressForPet(pet, newLevel > oldLevel), nil
}

func validCreditReason(reason models.LeafTransactionReason) bool {
	return reason == models.LeafReasonTaskReward || reason == models.LeafReasonWeeklyLogin
}

func LevelCost(level int) int64 {
	if level < MinPetLevel || level >= MaxPetLevel {
		return 0
	}

	return levelTargets[level+1] - levelTargets[level]
}

func ApplyLevelUps(level int, balance int64) (int, int64) {
	for level < MaxPetLevel {
		cost := LevelCost(level)

		if balance < cost {
			break
		}

		balance -= cost
		level++
	}

	return level, balance
}

func ProgressForPet(pet models.Pet, levelUp bool) Progress {
	progress := Progress{
		Name:       pet.Name,
		Level:      pet.Level,
		Leaves:     pet.Leaves,
		ChestPrice: models.ChestOpeningLeavesCost,
		LevelUp:    levelUp,
	}

	if pet.Level < MaxPetLevel {
		progress.NextLevelTargetLeaves = LevelCost(pet.Level)
	}

	return progress
}
