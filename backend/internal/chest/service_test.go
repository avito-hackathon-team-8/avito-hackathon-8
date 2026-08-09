package chest

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/rewards"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/testutil"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpenSpendsLeavesAndIssuesReward(t *testing.T) {
	service, db, user, petService := testService(t)
	if err := db.Create(&models.Pet{
		UserID: user.ID,
		Level:  pet.MaxPetLevel,
		Leaves: models.ChestOpeningLeavesCost + 75,
	}).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}

	updates, unsubscribe := petService.Subscribe(user.ID)
	defer unsubscribe()

	reward, err := service.Open(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if reward.Source != models.RewardSourceChest || reward.ChestOpeningID == nil || reward.LevelRewardID != nil {
		t.Fatalf("reward = %+v, want chest reward linked only to chest opening", reward)
	}
	if reward.Title != "1000 бонусов Авито" || reward.Category != models.RewardCategoryAvitoBonus {
		t.Fatalf("reward = %+v, want configured chest reward", reward)
	}

	var storedPet models.Pet
	if err := db.Where("user_id = ?", user.ID).First(&storedPet).Error; err != nil {
		t.Fatalf("load pet: %v", err)
	}
	if storedPet.Leaves != 75 || storedPet.Level != pet.MaxPetLevel {
		t.Fatalf("stored pet = %+v, want level 10 with 75 leaves", storedPet)
	}

	var opening models.ChestOpening
	if err := db.Where("id = ?", *reward.ChestOpeningID).First(&opening).Error; err != nil {
		t.Fatalf("load chest opening: %v", err)
	}
	if opening.UserID != user.ID || opening.LeavesSpent != models.ChestOpeningLeavesCost {
		t.Fatalf("opening = %+v, want user opening with 200 leaves spent", opening)
	}

	var issuedCount int64
	if err := db.Model(&models.Reward{}).Where("chest_opening_id = ?", opening.ID).Count(&issuedCount).Error; err != nil {
		t.Fatalf("count chest rewards: %v", err)
	}
	if issuedCount != 1 {
		t.Fatalf("chest rewards = %d, want 1", issuedCount)
	}

	var purchase models.LeafTransaction
	if err := db.Where("operation_key = ?", fmt.Sprintf("chest:%s", opening.ID)).First(&purchase).Error; err != nil {
		t.Fatalf("load chest purchase transaction: %v", err)
	}
	if purchase.UserID != user.ID || purchase.Amount != -models.ChestOpeningLeavesCost || purchase.Reason != models.LeafReasonChestPurchase {
		t.Fatalf("chest purchase = %+v, want a 200-leaf CHEST_PURCHASE debit", purchase)
	}

	var state models.UserGameState
	if err := db.Where("user_id = ?", user.ID).First(&state).Error; err != nil {
		t.Fatalf("load game state: %v", err)
	}
	if state.PetLevel != pet.MaxPetLevel || state.LeafBalance != 75 {
		t.Fatalf("game state = %+v, want level 10 with 75 leaves", state)
	}

	update := <-updates
	if update.Progress.Level != pet.MaxPetLevel || update.Progress.Leaves != 75 || update.Progress.LevelUp {
		t.Fatalf("progress update = %+v, want level 10 with 75 leaves", update.Progress)
	}
}

func TestOpenRejectsUnavailableChestWithoutChangingBalance(t *testing.T) {
	tests := []struct {
		name   string
		level  int
		leaves int64
		want   error
	}{
		{name: "level is below 10", level: pet.MaxPetLevel - 1, leaves: models.ChestOpeningLeavesCost, want: ErrChestLevelRequired},
		{name: "not enough leaves", level: pet.MaxPetLevel, leaves: models.ChestOpeningLeavesCost - 1, want: ErrInsufficientLeaves},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, db, user, _ := testService(t)
			if err := db.Create(&models.Pet{UserID: user.ID, Level: test.level, Leaves: test.leaves}).Error; err != nil {
				t.Fatalf("create pet: %v", err)
			}

			if _, err := service.Open(context.Background(), user.ID); !errors.Is(err, test.want) {
				t.Fatalf("Open() error = %v, want %v", err, test.want)
			}

			var storedPet models.Pet
			if err := db.Where("user_id = ?", user.ID).First(&storedPet).Error; err != nil {
				t.Fatalf("load pet: %v", err)
			}
			if storedPet.Leaves != test.leaves {
				t.Fatalf("stored leaves = %d, want %d", storedPet.Leaves, test.leaves)
			}

			var openings int64
			if err := db.Model(&models.ChestOpening{}).Count(&openings).Error; err != nil {
				t.Fatalf("count chest openings: %v", err)
			}
			if openings != 0 {
				t.Fatalf("chest openings = %d, want 0", openings)
			}
		})
	}
}

func TestOpenRejectsMissingPet(t *testing.T) {
	service, _, user, _ := testService(t)

	if _, err := service.Open(context.Background(), user.ID); !errors.Is(err, ErrPetNotFound) {
		t.Fatalf("Open() error = %v, want ErrPetNotFound", err)
	}
}

func TestOpenRejectsNilUser(t *testing.T) {
	service, _, _, _ := testService(t)

	if _, err := service.Open(context.Background(), uuid.Nil); !errors.Is(err, ErrPetNotFound) {
		t.Fatalf("Open() error = %v, want ErrPetNotFound", err)
	}
}

func TestOpenRollsBackWhenRewardSelectionFails(t *testing.T) {
	service, db, user, _ := testService(t)
	if err := db.Create(&models.Pet{
		UserID: user.ID,
		Level:  pet.MaxPetLevel,
		Leaves: models.ChestOpeningLeavesCost,
	}).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}

	selectionError := errors.New("reward selection failed")
	service.selectReward = func() (RewardDefinition, error) {
		return RewardDefinition{}, selectionError
	}

	if _, err := service.Open(context.Background(), user.ID); !errors.Is(err, selectionError) {
		t.Fatalf("Open() error = %v, want reward selection error", err)
	}

	var storedPet models.Pet
	if err := db.Where("user_id = ?", user.ID).First(&storedPet).Error; err != nil {
		t.Fatalf("load pet: %v", err)
	}
	if storedPet.Leaves != models.ChestOpeningLeavesCost {
		t.Fatalf("stored leaves = %d, want %d after rollback", storedPet.Leaves, models.ChestOpeningLeavesCost)
	}

	var openings, issuedRewards int64
	if err := db.Model(&models.ChestOpening{}).Count(&openings).Error; err != nil {
		t.Fatalf("count chest openings: %v", err)
	}
	if err := db.Model(&models.Reward{}).Count(&issuedRewards).Error; err != nil {
		t.Fatalf("count rewards: %v", err)
	}
	if openings != 0 || issuedRewards != 0 {
		t.Fatalf("openings = %d and rewards = %d, want no persisted records", openings, issuedRewards)
	}
}

func TestRandomRewardReturnsConfiguredReward(t *testing.T) {
	reward, err := randomReward(testChestRewardDefinitions())()
	if err != nil {
		t.Fatalf("randomReward() error = %v", err)
	}

	definitions := testChestRewardDefinitions()
	for _, definition := range definitions {
		if reward == definition {
			return
		}
	}

	t.Fatalf("randomReward() = %+v, want configured reward", reward)
}

func testService(t *testing.T) (*Service, *gorm.DB, models.User, *pet.Service) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Pet{}, &models.ChestOpening{}, &models.Reward{}, &models.LeafTransaction{}, &models.UserGameState{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	user := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString())}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	petService := pet.NewService(db)
	notifier := testutil.DailyReportNotifierMock{}
	rewardService := rewards.NewService(db, notifier)
	service := NewService(db, notifier, petService, rewardService, testChestRewardDefinitions())
	service.selectReward = func() (RewardDefinition, error) {
		return testChestRewardDefinitions()[0], nil
	}

	return service, db, user, petService
}

func testChestRewardDefinitions() []RewardDefinition {
	return []RewardDefinition{
		{Title: "1000 бонусов Авито", Category: models.RewardCategoryAvitoBonus},
		{Title: "Бесплатная доставка для трёх заказов", Category: models.RewardCategoryFreeDelivery},
		{Title: "Бесплатное продвижение объявления на 7 дней", Category: models.RewardCategoryFreePromotion},
	}
}
