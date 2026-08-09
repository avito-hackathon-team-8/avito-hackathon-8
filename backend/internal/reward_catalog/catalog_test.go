package reward_catalog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
)

func TestLoadReadsLevelAndChestRewards(t *testing.T) {
	path := writeCatalog(t, `
rewards:
  - level: 1
    title: Level 1
    description: Level reward 1
    category: AVITO_BONUS
  - level: 2
    title: Level 2
    description: Level reward 2
    category: AVITO_BONUS
  - level: 3
    title: Level 3
    description: Level reward 3
    category: AVITO_BONUS
  - level: 4
    title: Level 4
    description: Level reward 4
    category: AVITO_BONUS
  - level: 5
    title: Level 5
    description: Level reward 5
    category: AVITO_BONUS
  - level: 6
    title: Level 6
    description: Level reward 6
    category: AVITO_BONUS
  - level: 7
    title: Level 7
    description: Level reward 7
    category: AVITO_BONUS
  - level: 8
    title: Level 8
    description: Level reward 8
    category: AVITO_BONUS
  - level: 9
    title: Level 9
    description: Level reward 9
    category: AVITO_BONUS
  - level: 10
    title: Level 10
    description: Level reward 10
    category: AVITO_BONUS
chestRewards:
  - title: Chest reward
    category: FREE_DELIVERY
`)

	catalog, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	levels := catalog.LevelRewards()
	if len(levels) != 10 || levels[0].Level != 1 || levels[9].Level != 10 {
		t.Fatalf("level rewards = %+v, want levels 1 through 10", levels)
	}
	chests := catalog.ChestRewards()
	if len(chests) != 1 || chests[0].Title != "Chest reward" || chests[0].Category != models.RewardCategoryFreeDelivery {
		t.Fatalf("chest rewards = %+v, want configured chest reward", chests)
	}
}

func TestLoadRejectsIncompleteLevelRewards(t *testing.T) {
	path := writeCatalog(t, `
rewards:
  - level: 1
    title: Level 1
    description: Level reward 1
    category: AVITO_BONUS
chestRewards:
  - title: Chest reward
    category: FREE_DELIVERY
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want incomplete level rewards error")
	}
}

func writeCatalog(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "level_rewards.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write catalog: %v", err)
	}

	return path
}
