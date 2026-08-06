package tasks

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"gopkg.in/yaml.v3"
)

//go:embed task_definitions.yaml
var defaultDefinitionsYAML []byte

type Definition struct {
	Slot          int             `yaml:"slot"`
	Type          models.TaskType `yaml:"type"`
	Description   string          `yaml:"description"`
	TargetCount   int             `yaml:"targetCount"`
	RewardLeaves  int             `yaml:"rewardLeaves"`
	RequiredLevel int             `yaml:"requiredLevel"`
}

type definitionsFile struct {
	Tasks []Definition `yaml:"tasks"`
}

type slotRule struct {
	RewardLeaves  int
	RequiredLevel int
}

var dailyTaskSlots = []int{1, 2, 3, 4}

var slotRules = map[int]slotRule{
	1: {RewardLeaves: 45, RequiredLevel: 1},
	2: {RewardLeaves: 45, RequiredLevel: 1},
	3: {RewardLeaves: 50, RequiredLevel: 5},
	4: {RewardLeaves: 60, RequiredLevel: 10},
}

var knownTaskTypes = []models.TaskType{
	models.ViewListingsTaskType,
	models.AddToFavoritesTaskType,
	models.PublishListingTaskType,
	models.BoostListingTaskType,
	models.LeaveReviewTaskType,
	models.CompleteDealTaskType,
	models.OrderWithDeliveryTaskType,
}

func LoadDefaultDefinitions() ([]Definition, error) {
	return loadDefinitions(defaultDefinitionsYAML)
}

func IsKnownTaskType(taskType models.TaskType) bool {
	for _, knownType := range knownTaskTypes {
		if taskType == knownType {
			return true
		}
	}

	return false
}

func loadDefinitions(content []byte) ([]Definition, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)

	var config definitionsFile
	err := decoder.Decode(&config)

	if err != nil {
		return nil, fmt.Errorf("decode task definitions: %w", err)
	}

	var extra any
	err = decoder.Decode(&extra)

	if !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, fmt.Errorf("decode task definitions: expected one YAML document")
		}

		return nil, fmt.Errorf("decode task definitions: %w", err)
	}

	err = validateDefinitions(config.Tasks)

	if err != nil {
		return nil, err
	}

	return config.Tasks, nil
}

func validateDefinitions(definitions []Definition) error {
	if len(definitions) == 0 {
		return fmt.Errorf("task definitions are empty")
	}

	taskCountBySlot := make(map[int]int, TotalDailyTasks)
	taskTypeExists := make(map[models.TaskType]bool, len(knownTaskTypes))
	uniqueDefinitions := make(map[string]bool, len(definitions))

	for _, definition := range definitions {
		rule, ok := slotRules[definition.Slot]
		if !ok {
			return fmt.Errorf("task definition has unknown slot %d", definition.Slot)
		}

		if !IsKnownTaskType(definition.Type) {
			return fmt.Errorf("task definition has unknown type %q", definition.Type)
		}

		if strings.TrimSpace(definition.Description) == "" {
			return fmt.Errorf("task definition %d/%s has empty description", definition.Slot, definition.Type)
		}

		if definition.TargetCount < 1 {
			return fmt.Errorf("task definition %d/%s has invalid target count", definition.Slot, definition.Type)
		}

		if definition.RewardLeaves != rule.RewardLeaves {
			return fmt.Errorf("task definition %d/%s has reward %d, want %d", definition.Slot, definition.Type, definition.RewardLeaves, rule.RewardLeaves)
		}

		if definition.RequiredLevel != rule.RequiredLevel {
			return fmt.Errorf("task definition %d/%s has required level %d, want %d", definition.Slot, definition.Type, definition.RequiredLevel, rule.RequiredLevel)
		}

		key := fmt.Sprintf("%d:%s:%d", definition.Slot, definition.Type, definition.TargetCount)
		if uniqueDefinitions[key] {
			return fmt.Errorf("duplicate task definition %s", key)
		}

		uniqueDefinitions[key] = true
		taskCountBySlot[definition.Slot]++
		taskTypeExists[definition.Type] = true
	}

	for _, slot := range dailyTaskSlots {
		if taskCountBySlot[slot] == 0 {
			return fmt.Errorf("task definitions do not contain slot %d", slot)
		}
	}

	for _, taskType := range knownTaskTypes {
		if !taskTypeExists[taskType] {
			return fmt.Errorf("task definitions do not contain type %s", taskType)
		}
	}

	return nil
}
