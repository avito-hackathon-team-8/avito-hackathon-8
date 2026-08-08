package jobs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/puppeteer/internal/models"
	"gopkg.in/yaml.v3"
)

type taskCatalogFile struct {
	Tasks []taskDefinitionConfig `yaml:"tasks"`
}

type taskDefinitionConfig struct {
	Code         string   `yaml:"code"`
	Type         string   `yaml:"type"`
	Title        string   `yaml:"title"`
	Slot         int      `yaml:"slot"`
	TargetCount  int      `yaml:"targetCount"`
	RewardLeaves int      `yaml:"rewardLeaves"`
	UnlockLevel  int      `yaml:"unlockLevel"`
	Categories   []string `yaml:"categories"`
	Active       bool     `yaml:"active"`
}

type leaderboardRewardFile struct {
	Rewards []leaderboardRewardConfig `yaml:"rewards"`
}

type leaderboardRewardConfig struct {
	Rank     int    `yaml:"rank"`
	Title    string `yaml:"title"`
	Category string `yaml:"category"`
	ValidFor string `yaml:"validFor"`
}

type LeaderboardReward struct {
	Rank     int
	Title    string
	Category string
	ValidFor time.Duration
}

func LoadTaskDefinitions(path string) ([]models.DailyTaskDefinition, error) {
	content, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("read task definitions: %w", err)
	}

	var catalog taskCatalogFile

	if err := decodeYAML(content, &catalog); err != nil {
		return nil, fmt.Errorf("decode task definitions: %w", err)
	}

	if len(catalog.Tasks) == 0 {
		return nil, errors.New("task definitions are empty")
	}

	definitions := make([]models.DailyTaskDefinition, 0, len(catalog.Tasks))

	codes := make(map[string]struct{}, len(catalog.Tasks))
	slots := make(map[int]int, 4)

	typeSlots := make(map[string]int, len(catalog.Tasks))

	for _, item := range catalog.Tasks {
		item.Code = strings.TrimSpace(item.Code)
		item.Title = strings.TrimSpace(item.Title)

		if item.Code == "" || item.Title == "" || !models.IsKnownTaskType(item.Type) || item.Slot < 1 || item.Slot > 4 ||
			item.TargetCount < 1 || item.RewardLeaves < 1 || item.UnlockLevel < 1 || len(item.Categories) == 0 {

			return nil, fmt.Errorf("invalid task definition %q", item.Code)
		}

		if _, exists := codes[item.Code]; exists {
			return nil, fmt.Errorf("duplicate task code %q", item.Code)
		}

		if slot, exists := typeSlots[item.Type]; exists && slot != item.Slot {
			return nil, fmt.Errorf("task type %q is configured for multiple slots", item.Type)
		}

		codes[item.Code] = struct{}{}
		typeSlots[item.Type] = item.Slot

		if item.Active {
			slots[item.Slot]++
		}

		categories, _ := json.Marshal(item.Categories)

		definitions = append(definitions, models.DailyTaskDefinition{
			Code: item.Code, Type: item.Type, Title: item.Title, Slot: item.Slot,
			TargetCount: item.TargetCount, Reward: item.RewardLeaves, UnlockLevel: item.UnlockLevel,
			Categories: string(categories), Active: item.Active,
		})
	}

	for slot := 1; slot <= 4; slot++ {
		if slots[slot] == 0 {
			return nil, fmt.Errorf("task definitions do not contain slot %d", slot)
		}
	}

	return definitions, nil
}

func LoadLeaderboardRewards(path string) (map[int]LeaderboardReward, error) {
	content, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("read leaderboard rewards: %w", err)
	}

	var catalog leaderboardRewardFile

	if err := decodeYAML(content, &catalog); err != nil {
		return nil, fmt.Errorf("decode leaderboard rewards: %w", err)
	}

	if len(catalog.Rewards) != 3 {
		return nil, errors.New("leaderboard rewards must contain exactly ranks 1, 2, and 3")
	}

	rewards := make(map[int]LeaderboardReward, 3)

	for _, item := range catalog.Rewards {
		duration, err := time.ParseDuration(item.ValidFor)

		if err != nil || duration <= 0 || item.Rank < 1 || item.Rank > 3 || strings.TrimSpace(item.Title) == "" || !models.IsKnownRewardCategory(item.Category) {
			return nil, fmt.Errorf("invalid leaderboard reward for rank %d", item.Rank)
		}

		if _, exists := rewards[item.Rank]; exists {
			return nil, fmt.Errorf("duplicate leaderboard reward rank %d", item.Rank)
		}

		rewards[item.Rank] = LeaderboardReward{Rank: item.Rank, Title: strings.TrimSpace(item.Title), Category: item.Category, ValidFor: duration}
	}

	for rank := 1; rank <= 3; rank++ {
		if _, exists := rewards[rank]; !exists {
			return nil, fmt.Errorf("leaderboard reward rank %d is missing", rank)
		}
	}

	return rewards, nil
}

func decodeYAML(content []byte, target any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)

	if err := decoder.Decode(target); err != nil {
		return err
	}

	var extra any

	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("expected one YAML document")
		}

		return err
	}

	return nil
}
