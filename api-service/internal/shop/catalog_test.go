package shop

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
)

func TestLoadCatalog(t *testing.T) {
	catalog, err := LoadCatalog(writeCatalog(t, validCatalog))
	if err != nil {
		t.Fatalf("LoadCatalog() error = %v", err)
	}

	items := catalog.Items()
	if len(items) != 6 {
		t.Fatalf("items = %d, want 6", len(items))
	}
	item, ok := catalog.Item("cyber-bowl")
	if !ok || item.Type != models.ShopItemTypeCyberBowl || item.RequiredLevel != 7 || item.PriceLeaves != 150 ||
		item.ImageURL != "/api/v1/shop-images/bowl-cyberpunk.webp" {
		t.Fatalf("cyber-bowl = %+v, exists = %t", item, ok)
	}
}

func TestLoadCatalogRejectsInvalidCatalog(t *testing.T) {
	tests := []string{
		"items: []",
		"items:\n  - id: a\n    type: FASHIONABLE_BOWL",
	}
	for _, content := range tests {
		if _, err := LoadCatalog(writeCatalog(t, content)); err == nil {
			t.Fatalf("LoadCatalog(%q) error = nil", content)
		}
	}
}

func TestCatalogItemsReturnsCopyAndItemReportsMissingID(t *testing.T) {
	catalog, err := LoadCatalog(writeCatalog(t, validCatalog))
	if err != nil {
		t.Fatalf("LoadCatalog() error = %v", err)
	}

	items := catalog.Items()
	items[0].Title = "Changed"
	if fresh := catalog.Items(); fresh[0].Title == "Changed" {
		t.Fatal("Items() returned a mutable catalog slice")
	}
	if _, ok := catalog.Item("missing"); ok {
		t.Fatal("Item(missing) exists")
	}
}

func writeCatalog(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "shop_items.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write catalog: %v", err)
	}
	return path
}

const validCatalog = `items:
  - id: fashionable-bowl
    type: FASHIONABLE_BOWL
    title: Fashionable bowl
    description: Clothes cashback
    imageUrl: /api/v1/shop-images/bowl-fashionable.webp
    category: BOWL
    requiredLevel: 5
    priceLeaves: 100
    durationDays: 3
  - id: cyber-bowl
    type: CYBER_BOWL
    title: Cyber bowl
    description: Electronics cashback
    imageUrl: /api/v1/shop-images/bowl-cyberpunk.webp
    category: BOWL
    requiredLevel: 7
    priceLeaves: 150
    durationDays: 3
  - id: helper-bowl
    type: HELPER_BOWL
    title: Helper bowl
    description: Services cashback
    imageUrl: /api/v1/shop-images/bowl-helper.webp
    category: BOWL
    requiredLevel: 10
    priceLeaves: 200
    durationDays: 3
  - id: trader-bed
    type: TRADER_BED
    title: Trader bed
    description: Commission bonus
    imageUrl: /api/v1/shop-images/bed-sell.webp
    category: BED
    requiredLevel: 5
    priceLeaves: 300
    durationDays: 3
  - id: accident-free-bed
    type: ACCIDENT_FREE_BED
    title: Accident-free bed
    description: Vehicle report bonus
    imageUrl: /api/v1/shop-images/bed-car.webp
    category: BED
    requiredLevel: 7
    priceLeaves: 400
    durationDays: 3
  - id: pro-bed
    type: PRO_BED
    title: Pro bed
    description: Pro subscription bonus
    imageUrl: /api/v1/shop-images/bed-profi.webp
    category: BED
    requiredLevel: 10
    priceLeaves: 500
    durationDays: 3
`
