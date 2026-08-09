package reward_catalog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"gopkg.in/yaml.v3"
)

type LevelRewardDefinition struct {
	Level       int
	Title       string
	Description string
	Category    models.RewardCategory
}

type ChestRewardDefinition struct {
	Title    string
	Category models.RewardCategory
}

type Catalog struct {
	levelRewards []LevelRewardDefinition
	chestRewards []ChestRewardDefinition
}

type catalogFile struct {
	Rewards      []levelRewardConfig `yaml:"rewards"`
	ChestRewards []chestRewardConfig `yaml:"chestRewards"`
}

type levelRewardConfig struct {
	Level       int    `yaml:"level"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`
}

type chestRewardConfig struct {
	Title    string `yaml:"title"`
	Category string `yaml:"category"`
}

func Load(path string) (Catalog, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Catalog{}, fmt.Errorf("read reward catalog: %w", err)
	}

	var file catalogFile
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(&file); err != nil {
		return Catalog{}, fmt.Errorf("decode reward catalog: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Catalog{}, errors.New("reward catalog must contain one YAML document")
		}
		return Catalog{}, fmt.Errorf("decode reward catalog: %w", err)
	}

	if len(file.Rewards) != 10 {
		return Catalog{}, errors.New("reward catalog must contain exactly 10 level rewards")
	}
	if len(file.ChestRewards) == 0 {
		return Catalog{}, errors.New("reward catalog must contain at least one chest reward")
	}

	catalog := Catalog{
		levelRewards: make([]LevelRewardDefinition, 0, len(file.Rewards)),
		chestRewards: make([]ChestRewardDefinition, 0, len(file.ChestRewards)),
	}
	levels := make(map[int]struct{}, len(file.Rewards))
	for _, item := range file.Rewards {
		definition := LevelRewardDefinition{
			Level:       item.Level,
			Title:       strings.TrimSpace(item.Title),
			Description: strings.TrimSpace(item.Description),
			Category:    models.RewardCategory(strings.TrimSpace(item.Category)),
		}
		if definition.Level < 1 || definition.Level > 10 || definition.Title == "" ||
			definition.Description == "" || !validCategory(definition.Category) {
			return Catalog{}, fmt.Errorf("invalid level reward for level %d", item.Level)
		}
		if _, exists := levels[definition.Level]; exists {
			return Catalog{}, fmt.Errorf("duplicate level reward %d", definition.Level)
		}
		levels[definition.Level] = struct{}{}
		catalog.levelRewards = append(catalog.levelRewards, definition)
	}
	for level := 1; level <= 10; level++ {
		if _, exists := levels[level]; !exists {
			return Catalog{}, fmt.Errorf("level reward %d is missing", level)
		}
	}

	for _, item := range file.ChestRewards {
		definition := ChestRewardDefinition{
			Title:    strings.TrimSpace(item.Title),
			Category: models.RewardCategory(strings.TrimSpace(item.Category)),
		}
		if definition.Title == "" || !validCategory(definition.Category) {
			return Catalog{}, fmt.Errorf("invalid chest reward %q", item.Title)
		}
		catalog.chestRewards = append(catalog.chestRewards, definition)
	}

	return catalog, nil
}

func (catalog Catalog) LevelRewards() []LevelRewardDefinition {
	return append([]LevelRewardDefinition(nil), catalog.levelRewards...)
}

func (catalog Catalog) ChestRewards() []ChestRewardDefinition {
	return append([]ChestRewardDefinition(nil), catalog.chestRewards...)
}

func validCategory(category models.RewardCategory) bool {
	switch category {
	case models.RewardCategoryAvitoBonus,
		models.RewardCategoryFreeDelivery,
		models.RewardCategoryFreePromotion,
		models.RewardCategoryPromotionDiscount,
		models.RewardCategoryDeliveryDiscount:
		return true
	default:
		return false
	}
}
