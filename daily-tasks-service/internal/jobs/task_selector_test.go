package jobs

import (
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/daily-tasks-service/internal/models"
	"github.com/google/uuid"
)

func TestSimilaritySelectorIsDeterministicAndSelectsEachSlot(t *testing.T) {
	user := models.User{
		ID:        uuid.New(),
		Interests: `{"electronics": 1, "home": 0.5}`,
	}

	definitions := make([]models.DailyTaskDefinition, 0, 8)

	for slot := 1; slot <= 4; slot++ {
		definitions = append(definitions,
			models.DailyTaskDefinition{
				Code:       uuid.NewString(),
				Slot:       slot,
				Categories: `["home"]`,
			},
			models.DailyTaskDefinition{
				Code:       uuid.NewString(),
				Slot:       slot,
				Categories: `["electronics"]`,
			},
		)
	}

	day := time.Date(2026, time.August, 9, 0, 0, 0, 0, time.UTC)
	selector := SimilaritySelector{}
	first := selector.Select(user, definitions, day)
	second := selector.Select(user, definitions, day)

	if len(first) != 4 || len(second) != 4 {
		t.Fatalf("selected task counts = %d and %d, want 4", len(first), len(second))
	}

	for index := range first {
		if first[index].Slot != index+1 {
			t.Fatalf("task %d slot = %d, want %d", index, first[index].Slot, index+1)
		}
		if first[index].Code != second[index].Code {
			t.Fatalf("selection is not deterministic for slot %d", index+1)
		}
	}
}
