package shop

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
	"gopkg.in/yaml.v3"
)

type Catalog struct {
	items map[string]models.ShopItem
	list  []models.ShopItem
}

type catalogFile struct {
	Items []shopItemConfig `yaml:"items"`
}

type shopItemConfig struct {
	ID            string                `yaml:"id"`
	Type          models.ShopItemType   `yaml:"type"`
	Title         string                `yaml:"title"`
	Description   string                `yaml:"description"`
	ImageURL      string                `yaml:"imageUrl"`
	Category      models.RewardCategory `yaml:"category"`
	RequiredLevel int                   `yaml:"requiredLevel"`
	PriceLeaves   int64                 `yaml:"priceLeaves"`
	DurationDays  int                   `yaml:"durationDays"`
}

func LoadCatalog(path string) (Catalog, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Catalog{}, fmt.Errorf("read shop catalog: %w", err)
	}

	var file catalogFile
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(&file); err != nil {
		return Catalog{}, fmt.Errorf("decode shop catalog: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Catalog{}, errors.New("shop catalog must contain one YAML document")
		}
		return Catalog{}, fmt.Errorf("decode shop catalog: %w", err)
	}

	if len(file.Items) != 6 {
		return Catalog{}, errors.New("shop catalog must contain exactly 6 items")
	}

	catalog := Catalog{
		items: make(map[string]models.ShopItem, len(file.Items)),
		list:  make([]models.ShopItem, 0, len(file.Items)),
	}
	types := make(map[models.ShopItemType]struct{}, len(file.Items))
	for _, config := range file.Items {
		item := models.ShopItem{
			ID:            strings.TrimSpace(config.ID),
			Type:          config.Type,
			Title:         strings.TrimSpace(config.Title),
			Description:   strings.TrimSpace(config.Description),
			ImageURL:      strings.TrimSpace(config.ImageURL),
			Category:      config.Category,
			RequiredLevel: config.RequiredLevel,
			PriceLeaves:   config.PriceLeaves,
			DurationDays:  config.DurationDays,
		}
		if !validItem(item) {
			return Catalog{}, fmt.Errorf("invalid shop item %q", config.ID)
		}
		if _, exists := catalog.items[item.ID]; exists {
			return Catalog{}, fmt.Errorf("duplicate shop item %q", item.ID)
		}
		if _, exists := types[item.Type]; exists {
			return Catalog{}, fmt.Errorf("duplicate shop item type %q", item.Type)
		}
		catalog.items[item.ID] = item
		catalog.list = append(catalog.list, item)
		types[item.Type] = struct{}{}
	}

	return catalog, nil
}

func (catalog Catalog) Items() []models.ShopItem {
	return append([]models.ShopItem(nil), catalog.list...)
}

func (catalog Catalog) Item(id string) (models.ShopItem, bool) {
	item, ok := catalog.items[id]
	return item, ok
}

func validItem(item models.ShopItem) bool {
	if item.ID == "" || item.Title == "" || item.Description == "" || item.ImageURL == "" || item.RequiredLevel < 1 ||
		item.RequiredLevel > 10 || item.PriceLeaves <= 0 || item.DurationDays <= 0 {
		return false
	}

	switch item.Type {
	case models.ShopItemTypeFashionableBowl, models.ShopItemTypeCyberBowl, models.ShopItemTypeHelperBowl:
		return item.Category == models.RewardCategoryBowl
	case models.ShopItemTypeTraderBed, models.ShopItemTypeAccidentFreeBed, models.ShopItemTypeProBed:
		return item.Category == models.RewardCategoryBed
	default:
		return false
	}
}
