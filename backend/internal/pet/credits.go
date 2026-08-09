package pet

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

const maxInt64 = int64(1<<63 - 1)

var (
	ErrInvalidAmount      = errors.New("leaves amount must be positive")
	ErrInvalidOperation   = errors.New("leaf operation is invalid")
	ErrDuplicateOperation = errors.New("leaf operation already applied")
	ErrLeavesOverflow     = errors.New("leaves amount overflows int64")
	ErrInsufficientLeaves = errors.New("insufficient leaves")
)

var levelTargets = [...]int64{0, 0, 100, 230, 390, 580, 810, 1090, 1430, 1850, 2400}

type Credit struct {
	UserID       uuid.UUID
	Amount       int64
	Reason       models.LeafTransactionReason
	OperationKey string
	OccurredAt   time.Time
}

type Debit struct {
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

	if service.dailyReport != nil {
		service.dailyReport.Notify(credit.UserID)
	}

	service.publish(Update{UserID: credit.UserID, Progress: progress})

	return progress, nil
}

func (service *Service) CreditTx(tx *gorm.DB, credit Credit) (Progress, error) {
	if credit.Amount <= 0 {
		return Progress{}, ErrInvalidAmount
	}

	credit.OperationKey = strings.TrimSpace(credit.OperationKey)

	if credit.UserID == uuid.Nil ||
		credit.OperationKey == "" ||
		len(credit.OperationKey) > 160 ||
		!validCreditReason(credit.Reason) {
		return Progress{}, ErrInvalidOperation
	}

	if credit.OccurredAt.IsZero() {
		credit.OccurredAt = service.now().UTC()
	} else {
		credit.OccurredAt = credit.OccurredAt.UTC()
	}

	result := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "operation_key"}},
		DoNothing: true,
	}).Create(&models.LeafTransaction{
		UserID:       credit.UserID,
		Amount:       credit.Amount,
		Reason:       credit.Reason,
		OperationKey: credit.OperationKey,
		OccurredAt:   credit.OccurredAt,
	})

	if result.Error != nil {
		return Progress{}, fmt.Errorf("record leaf credit: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return Progress{}, ErrDuplicateOperation
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoNothing: true,
	}).Create(&models.Pet{
		UserID: credit.UserID,
		Level:  InitialPetLevel,
		Leaves: InitialPetLeaves,
	}).Error; err != nil {
		return Progress{}, fmt.Errorf("ensure pet: %w", err)
	}

	var userPet models.Pet

	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", credit.UserID).
		First(&userPet).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Progress{}, ErrPetNotFound
	}

	if err != nil {
		return Progress{}, fmt.Errorf("lock pet: %w", err)
	}

	if userPet.Leaves < 0 || userPet.Leaves > maxInt64-credit.Amount {
		return Progress{}, ErrLeavesOverflow
	}

	oldLevel := userPet.Level
	newLevel, remainingLeaves := ApplyLevelUps(userPet.Level, userPet.Leaves+credit.Amount)
	userPet.Level = newLevel
	userPet.Leaves = remainingLeaves

	if err := tx.Model(&userPet).Updates(map[string]any{
		"level":  userPet.Level,
		"leaves": userPet.Leaves,
	}).Error; err != nil {
		return Progress{}, fmt.Errorf("save pet progress: %w", err)
	}

	for level := oldLevel; level < newLevel; level++ {
		cost := LevelCost(level)

		if err := tx.Create(&models.LeafTransaction{
			UserID:       credit.UserID,
			Amount:       -cost,
			Reason:       models.LeafReasonLevelUp,
			OperationKey: fmt.Sprintf("%s:level:%d", credit.OperationKey, level+1),
			OccurredAt:   credit.OccurredAt,
		}).Error; err != nil {
			return Progress{}, fmt.Errorf("record level spending: %w", err)
		}
	}

	if service.levelClaims != nil && newLevel > oldLevel {
		if err := service.levelClaims.openReachedRewards(tx, credit.UserID, newLevel); err != nil {
			return Progress{}, fmt.Errorf("open reached level rewards: %w", err)
		}
	}

	if err := tx.Exec(`
		INSERT INTO user_game_states (user_id, pet_level, leaf_balance, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET pet_level = EXCLUDED.pet_level,
		leaf_balance = EXCLUDED.leaf_balance, updated_at = EXCLUDED.updated_at`,
		credit.UserID,
		userPet.Level,
		userPet.Leaves,
		service.now().UTC(),
	).Error; err != nil {
		return Progress{}, fmt.Errorf("sync game state: %w", err)
	}

	return ProgressForPet(userPet, newLevel > oldLevel), nil
}

func (service *Service) DebitTx(tx *gorm.DB, debit Debit) (Progress, error) {
	if debit.Amount <= 0 {
		return Progress{}, ErrInvalidAmount
	}

	debit.OperationKey = strings.TrimSpace(debit.OperationKey)
	if debit.UserID == uuid.Nil || debit.OperationKey == "" || len(debit.OperationKey) > 160 ||
		!validDebitReason(debit.Reason) {
		return Progress{}, ErrInvalidOperation
	}

	if debit.OccurredAt.IsZero() {
		debit.OccurredAt = service.now().UTC()
	} else {
		debit.OccurredAt = debit.OccurredAt.UTC()
	}

	var userPet models.Pet
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", debit.UserID).
		First(&userPet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Progress{}, ErrPetNotFound
	}
	if err != nil {
		return Progress{}, fmt.Errorf("lock pet: %w", err)
	}
	if userPet.Leaves < debit.Amount {
		return Progress{}, ErrInsufficientLeaves
	}

	transaction := models.LeafTransaction{
		UserID:       debit.UserID,
		Amount:       -debit.Amount,
		Reason:       debit.Reason,
		OperationKey: debit.OperationKey,
		OccurredAt:   debit.OccurredAt,
	}
	result := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "operation_key"}},
		DoNothing: true,
	}).Create(&transaction)
	if result.Error != nil {
		return Progress{}, fmt.Errorf("record leaf debit: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return Progress{}, ErrDuplicateOperation
	}

	userPet.Leaves -= debit.Amount
	if err := tx.Model(&userPet).Update("leaves", userPet.Leaves).Error; err != nil {
		return Progress{}, fmt.Errorf("save pet balance: %w", err)
	}
	if err := tx.Exec(`
		INSERT INTO user_game_states (user_id, pet_level, leaf_balance, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET pet_level = EXCLUDED.pet_level,
		leaf_balance = EXCLUDED.leaf_balance, updated_at = EXCLUDED.updated_at`,
		debit.UserID,
		userPet.Level,
		userPet.Leaves,
		service.now().UTC(),
	).Error; err != nil {
		return Progress{}, fmt.Errorf("sync game state: %w", err)
	}

	return ProgressForPet(userPet, false), nil
}

func validCreditReason(reason models.LeafTransactionReason) bool {
	return reason == models.LeafReasonTaskReward || reason == models.LeafReasonWeeklyLogin
}

func validDebitReason(reason models.LeafTransactionReason) bool {
	return reason == models.LeafReasonChestPurchase
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

func ProgressForPet(userPet models.Pet, levelUp bool) Progress {
	progress := Progress{
		Name:       userPet.Name,
		Level:      userPet.Level,
		Leaves:     userPet.Leaves,
		ChestPrice: models.ChestOpeningLeavesCost,
		LevelUp:    levelUp,
	}

	if userPet.Level < MaxPetLevel {
		progress.NextLevelTargetLeaves = LevelCost(userPet.Level)
	}

	return progress
}
