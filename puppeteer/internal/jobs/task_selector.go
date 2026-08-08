package jobs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/puppeteer/internal/models"
)

type TaskSelector interface {
	Select(user models.User, definitions []models.DailyTaskDefinition, day time.Time) []models.DailyTaskDefinition
}

type SimilaritySelector struct{}

func (SimilaritySelector) Select(user models.User, definitions []models.DailyTaskDefinition, day time.Time) []models.DailyTaskDefinition {
	interests := decodeWeights(user.Interests)
	bySlot := make(map[int][]models.DailyTaskDefinition)

	for _, definition := range definitions {
		bySlot[definition.Slot] = append(bySlot[definition.Slot], definition)
	}

	result := make([]models.DailyTaskDefinition, 0, 4)

	for slot := 1; slot <= 4; slot++ {
		candidates := bySlot[slot]

		if len(candidates) == 0 {
			continue
		}

		sort.SliceStable(candidates, func(i, j int) bool {
			si := cosine(interests, decodeCategories(candidates[i].Categories))
			sj := cosine(interests, decodeCategories(candidates[j].Categories))

			if math.Abs(si-sj) > 1e-9 {
				return si > sj
			}

			return tieValue(user.ID, day, slot, candidates[i].Code) < tieValue(user.ID, day, slot, candidates[j].Code)
		})

		result = append(result, candidates[0])
	}

	return result
}

func decodeWeights(raw string) map[string]float64 {
	var result map[string]float64

	if raw == "" || json.Unmarshal([]byte(raw), &result) != nil {
		return map[string]float64{}
	}

	return result
}

func decodeCategories(raw string) map[string]float64 {
	var categories []string

	if raw == "" || json.Unmarshal([]byte(raw), &categories) != nil {
		return map[string]float64{}
	}

	result := make(map[string]float64, len(categories))

	for _, category := range categories {
		result[category] = 1
	}

	return result
}

func cosine(left, right map[string]float64) float64 {
	var dot, leftNorm, rightNorm float64

	for key, value := range left {
		leftNorm += value * value

		if other, ok := right[key]; ok {
			dot += value * other
		}
	}

	for _, value := range right {
		rightNorm += value * value
	}

	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}

	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}

func tieValue(userID interface{ String() string }, day time.Time, slot int, code string) uint64 {
	digest := sha256.Sum256([]byte(userID.String() + ":" + day.Format(time.DateOnly) + ":" + strconv.Itoa(slot) + ":" + code))

	return binary.BigEndian.Uint64(digest[:8])
}
