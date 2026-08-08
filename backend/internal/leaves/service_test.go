package leaves

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/testutil"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestApplyLevelUps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		level, wantLevel     int
		balance, wantBalance int64
	}{
		{1, 1, 99, 99},
		{1, 2, 100, 0},
		{1, 4, 390, 0},
		{4, 5, 300, 110},
		{10, 10, 10_000, 10_000},
	}
	for _, test := range tests {
		level, balance := ApplyLevelUps(test.level, test.balance)
		if level != test.wantLevel || balance != test.wantBalance {
			t.Errorf("ApplyLevelUps(%d, %d) = (%d, %d), want (%d, %d)", test.level, test.balance, level, balance, test.wantLevel, test.wantBalance)
		}
	}
}

func TestCreditRecordsLedgerAndMultipleLevelUps(t *testing.T) {
	service, db, user := testLeavesService(t, true)
	result, err := service.Credit(context.Background(), Credit{
		UserID: user.ID, Amount: 390, Reason: models.LeafReasonTaskReward, OperationKey: "task:one",
	})
	if err != nil {
		t.Fatalf("Credit() error = %v", err)
	}
	if result.Level != 4 || result.Leaves != 0 || result.NextLevelTargetLeaves != 190 || result.ChestPrice != models.ChestOpeningLeavesCost || !result.LevelUp {
		t.Fatalf("progress = %+v, want level 4 with target 190 and chest price %d", result, models.ChestOpeningLeavesCost)
	}

	var transactions []models.LeafTransaction
	if err := db.Order("amount DESC").Find(&transactions).Error; err != nil {
		t.Fatalf("load ledger: %v", err)
	}
	if len(transactions) != 4 || transactions[0].Amount != 390 {
		t.Fatalf("transactions = %+v, want credit and three level-up debits", transactions)
	}
	var sum int64
	for _, transaction := range transactions {
		sum += transaction.Amount
	}
	if sum != 0 {
		t.Fatalf("ledger sum = %d, want 0", sum)
	}

	var state models.UserGameState
	if err := db.First(&state, "user_id = ?", user.ID).Error; err != nil {
		t.Fatalf("load game state: %v", err)
	}
	if state.PetLevel != 4 || state.LeafBalance != 0 {
		t.Fatalf("game state = %+v", state)
	}
}

func TestCreditRejectsDuplicateOperationWithoutChangingBalance(t *testing.T) {
	service, db, user := testLeavesService(t, true)
	credit := Credit{UserID: user.ID, Amount: 25, Reason: models.LeafReasonWeeklyLogin, OperationKey: "weekly-login:one"}
	if _, err := service.Credit(context.Background(), credit); err != nil {
		t.Fatalf("first Credit() error = %v", err)
	}
	if _, err := service.Credit(context.Background(), credit); !errors.Is(err, ErrDuplicateOperation) {
		t.Fatalf("second Credit() error = %v, want ErrDuplicateOperation", err)
	}
	var stored models.Pet
	if err := db.First(&stored, "user_id = ?", user.ID).Error; err != nil {
		t.Fatalf("load pet: %v", err)
	}
	if stored.Leaves != 25 {
		t.Fatalf("pet leaves = %d, want 25", stored.Leaves)
	}
}

func TestCreditDetectsOverflow(t *testing.T) {
	service, db, user := testLeavesService(t, true)
	if err := db.Model(&models.Pet{}).Where("user_id = ?", user.ID).Updates(map[string]any{"level": MaxPetLevel, "leaves": maxInt64}).Error; err != nil {
		t.Fatalf("seed max balance: %v", err)
	}
	if _, err := service.Credit(context.Background(), Credit{UserID: user.ID, Amount: 1, Reason: models.LeafReasonTaskReward, OperationKey: "task:overflow"}); !errors.Is(err, ErrLeavesOverflow) {
		t.Fatalf("Credit() error = %v, want ErrLeavesOverflow", err)
	}
	var count int64
	if err := db.Model(&models.LeafTransaction{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("ledger count after overflow = %d, error = %v", count, err)
	}
}

func TestCreditValidatesAmountReasonAndOperationKey(t *testing.T) {
	service, _, user := testLeavesService(t, true)
	if _, err := service.Credit(context.Background(), Credit{UserID: user.ID, Amount: 0, Reason: models.LeafReasonTaskReward, OperationKey: "task:zero"}); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("zero amount error = %v, want ErrInvalidAmount", err)
	}
	if _, err := service.Credit(context.Background(), Credit{UserID: user.ID, Amount: 1, Reason: models.LeafReasonLevelUp, OperationKey: "level:invalid-credit"}); !errors.Is(err, ErrInvalidOperation) {
		t.Fatalf("invalid credit reason error = %v, want ErrInvalidOperation", err)
	}
	if _, err := service.Credit(context.Background(), Credit{UserID: user.ID, Amount: 1, Reason: models.LeafReasonTaskReward, OperationKey: " "}); !errors.Is(err, ErrInvalidOperation) {
		t.Fatalf("empty operation key error = %v, want ErrInvalidOperation", err)
	}
}

func TestCreditRollsBackWhenStateSyncFails(t *testing.T) {
	service, db, user := testLeavesService(t, false)
	if _, err := service.Credit(context.Background(), Credit{UserID: user.ID, Amount: 10, Reason: models.LeafReasonTaskReward, OperationKey: "task:rollback"}); err == nil {
		t.Fatal("Credit() error = nil, want missing state table error")
	}
	var stored models.Pet
	if err := db.First(&stored, "user_id = ?", user.ID).Error; err != nil {
		t.Fatalf("load pet: %v", err)
	}
	if stored.Leaves != 0 {
		t.Fatalf("pet leaves = %d, want rollback to zero", stored.Leaves)
	}
	var count int64
	if err := db.Model(&models.LeafTransaction{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("ledger count after rollback = %d, error = %v", count, err)
	}
}

func testLeavesService(t *testing.T, withState bool) (*Service, *gorm.DB, models.User) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	modelsToMigrate := []any{&models.User{}, &models.Pet{}, &models.LeafTransaction{}}
	if withState {
		modelsToMigrate = append(modelsToMigrate, &models.UserGameState{})
	}
	if err := db.AutoMigrate(modelsToMigrate...); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	user := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString()), Verified: true}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&models.Pet{UserID: user.ID, Level: 1}).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	return NewService(db, testutil.DailyReportNotifierMock{}), db, user
}
