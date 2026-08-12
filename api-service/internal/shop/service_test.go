package shop

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/rewards"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/testutil"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPurchaseSpendsLeavesAndIssuesShopReward(t *testing.T) {
	service, db, user, petService := testService(t)
	if err := db.Create(&models.Pet{UserID: user.ID, Level: 5, Leaves: 350}).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}

	updates, unsubscribe := petService.Subscribe(user.ID)
	defer unsubscribe()

	if err := service.Purchase(context.Background(), user.ID, Purchase{ItemID: "fashionable-bowl"}); err != nil {
		t.Fatalf("Purchase() error = %v", err)
	}

	var reward models.Reward
	if err := db.Where("user_id = ? AND source = ?", user.ID, models.RewardSourceShop).First(&reward).Error; err != nil {
		t.Fatalf("load shop reward: %v", err)
	}
	if reward.Source != models.RewardSourceShop || reward.ItemType == nil || *reward.ItemType != models.ShopItemTypeFashionableBowl {
		t.Fatalf("reward = %+v, want fashionable bowl shop reward", reward)
	}
	if reward.Category != models.RewardCategoryBowl || reward.ExpiresAt != testNow.AddDate(0, 0, 3) {
		t.Fatalf("reward = %+v, want bowl that expires in three days", reward)
	}
	var transaction models.LeafTransaction
	if err := db.Where("reason = ?", models.LeafReasonShopPurchase).First(&transaction).Error; err != nil {
		t.Fatalf("load shop transaction: %v", err)
	}
	if transaction.Amount != -100 {
		t.Fatalf("transaction amount = %d, want -100", transaction.Amount)
	}

	update := <-updates
	if update.Progress.Leaves != 250 {
		t.Fatalf("published progress = %+v, want 250 leaves", update.Progress)
	}
}

func TestPurchaseRejectsUnavailableItemsWithoutSpendingLeaves(t *testing.T) {
	tests := []struct {
		name   string
		level  int
		leaves int64
		itemID string
		want   error
	}{
		{name: "unknown item", level: 10, leaves: 500, itemID: "missing", want: ErrItemNotFound},
		{name: "level required", level: 4, leaves: 500, itemID: "fashionable-bowl", want: ErrLevelRequired},
		{name: "insufficient leaves", level: 5, leaves: 99, itemID: "fashionable-bowl", want: ErrInsufficientLeaves},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, db, user, _ := testService(t)
			if err := db.Create(&models.Pet{UserID: user.ID, Level: test.level, Leaves: test.leaves}).Error; err != nil {
				t.Fatalf("create pet: %v", err)
			}

			if err := service.Purchase(context.Background(), user.ID, Purchase{ItemID: test.itemID}); !errors.Is(err, test.want) {
				t.Fatalf("Purchase() error = %v, want %v", err, test.want)
			}

			var stored models.Pet
			if err := db.Where("user_id = ?", user.ID).First(&stored).Error; err != nil {
				t.Fatalf("load pet: %v", err)
			}
			if stored.Leaves != test.leaves {
				t.Fatalf("leaves = %d, want %d", stored.Leaves, test.leaves)
			}
		})
	}
}

func TestPurchaseRejectsMissingPetAndNilUser(t *testing.T) {
	service, _, user, _ := testService(t)

	if err := service.Purchase(context.Background(), uuid.Nil, Purchase{ItemID: "fashionable-bowl"}); !errors.Is(err, ErrPetNotFound) {
		t.Fatalf("Purchase(nil user) error = %v, want %v", err, ErrPetNotFound)
	}
	if err := service.Purchase(context.Background(), user.ID, Purchase{ItemID: "fashionable-bowl"}); !errors.Is(err, ErrPetNotFound) {
		t.Fatalf("Purchase(missing pet) error = %v, want %v", err, ErrPetNotFound)
	}
}

func TestListCalculatesItemStatusesForCurrentUser(t *testing.T) {
	service, db, user, _ := testService(t)
	if err := db.Create(&models.Pet{UserID: user.ID, Level: 7, Leaves: 1000}).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}

	if err := service.Purchase(context.Background(), user.ID, Purchase{ItemID: "fashionable-bowl"}); err != nil {
		t.Fatalf("Purchase() error = %v", err)
	}

	items, err := service.List(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	statuses := make(map[string]ItemStatus, len(items))
	for _, item := range items {
		statuses[item.ID] = item.Status
	}

	want := map[string]ItemStatus{
		"fashionable-bowl":  ItemStatusActive,
		"cyber-bowl":        ItemStatusAvailable,
		"helper-bowl":       ItemStatusLocked,
		"trader-bed":        ItemStatusAvailable,
		"accident-free-bed": ItemStatusAvailable,
		"pro-bed":           ItemStatusLocked,
	}
	for itemID, wantStatus := range want {
		if statuses[itemID] != wantStatus {
			t.Errorf("status of %s = %q, want %q", itemID, statuses[itemID], wantStatus)
		}
	}
}

func TestActiveImageURLsReturnsOnlyActiveBowlAndBedImages(t *testing.T) {
	service, db, user, _ := testService(t)
	if err := db.Create(&models.Pet{UserID: user.ID, Level: 10, Leaves: 1000}).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}

	empty, err := service.ActiveImageURLs(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("ActiveImageURLs() without purchases error = %v", err)
	}
	if empty.Bowl != nil || empty.Bed != nil {
		t.Fatalf("ActiveImageURLs() without purchases = %+v, want nil URLs", empty)
	}

	if err := service.Purchase(context.Background(), user.ID, Purchase{ItemID: "fashionable-bowl"}); err != nil {
		t.Fatalf("purchase bowl: %v", err)
	}
	if err := service.Purchase(context.Background(), user.ID, Purchase{ItemID: "trader-bed"}); err != nil {
		t.Fatalf("purchase bed: %v", err)
	}

	images, err := service.ActiveImageURLs(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("ActiveImageURLs() error = %v", err)
	}
	if images.Bowl == nil || *images.Bowl != "/api/v1/shop-images/bowl-fashionable.webp" {
		t.Fatalf("bowl image = %v, want fashionable bowl image URL", images.Bowl)
	}
	if images.Bed == nil || *images.Bed != "/api/v1/shop-images/bed-sell.webp" {
		t.Fatalf("bed image = %v, want trader bed image URL", images.Bed)
	}
}

func TestPurchaseExtendsSameItemAndRequiresReplacementConfirmation(t *testing.T) {
	service, db, user, _ := testService(t)
	if err := db.Create(&models.Pet{UserID: user.ID, Level: 7, Leaves: 600}).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}

	if err := service.Purchase(context.Background(), user.ID, Purchase{ItemID: "fashionable-bowl"}); err != nil {
		t.Fatalf("first Purchase() error = %v", err)
	}
	var first models.Reward
	if err := db.Where("user_id = ? AND source = ?", user.ID, models.RewardSourceShop).First(&first).Error; err != nil {
		t.Fatalf("load first reward: %v", err)
	}

	if err := service.Purchase(context.Background(), user.ID, Purchase{ItemID: "fashionable-bowl"}); err != nil {
		t.Fatalf("second Purchase() error = %v", err)
	}
	var second models.Reward
	if err := db.Where("id = ?", first.ID).First(&second).Error; err != nil {
		t.Fatalf("load extended reward: %v", err)
	}
	if second.ID != first.ID || second.ExpiresAt != testNow.AddDate(0, 0, 6) {
		t.Fatalf("extended reward = %+v, want original reward extended to six days", second)
	}

	if err := service.Purchase(context.Background(), user.ID, Purchase{ItemID: "cyber-bowl"}); !errors.Is(err, ErrReplacementConfirmation) {
		t.Fatalf("replacement Purchase() error = %v, want %v", err, ErrReplacementConfirmation)
	}
	if err := service.Purchase(context.Background(), user.ID, Purchase{ItemID: "cyber-bowl", ConfirmReplacement: true}); err != nil {
		t.Fatalf("confirmed replacement error = %v", err)
	}
	var replacement models.Reward
	if err := db.Where("user_id = ? AND item_type = ?", user.ID, models.ShopItemTypeCyberBowl).First(&replacement).Error; err != nil {
		t.Fatalf("load replacement reward: %v", err)
	}
	if replacement.ItemType == nil || *replacement.ItemType != models.ShopItemTypeCyberBowl {
		t.Fatalf("replacement = %+v, want cyber bowl", replacement)
	}

	var replaced models.Reward
	if err := db.Where("id = ?", first.ID).First(&replaced).Error; err != nil {
		t.Fatalf("load replaced reward: %v", err)
	}
	if !replaced.ExpiresAt.Equal(testNow) {
		t.Fatalf("replaced reward expires at %s, want %s", replaced.ExpiresAt, testNow)
	}
}

const testNowText = "2030-08-05T12:00:00Z"

var testNow, _ = time.Parse(time.RFC3339, testNowText)

func testService(t *testing.T) (*Service, *gorm.DB, models.User, *pet.Service) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Pet{}, &models.Reward{}, &models.LeafTransaction{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	user := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString())}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	catalog, err := LoadCatalog(writeCatalog(t, validCatalog))
	if err != nil {
		t.Fatalf("LoadCatalog() error = %v", err)
	}
	petService := pet.NewService(db)
	notifier := testutil.DailyReportNotifierMock{}
	rewardService := rewards.NewService(db, notifier)
	service := NewService(db, notifier, petService, rewardService, catalog)
	service.now = func() time.Time { return testNow }

	return service, db, user, petService
}
