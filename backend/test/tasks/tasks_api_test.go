package tasks_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type testConfig struct {
	repoRoot     string
	apiURL       string
	postgresDB   string
	postgresUser string
	jwtSecret    string
}

type apiResult struct {
	status int
	body   []byte
	json   map[string]any
}

type tasksResponse struct {
	Tasks []taskResponse `json:"tasks"`
}

type taskResponse struct {
	TaskID        string `json:"taskId"`
	Slot          int    `json:"slot"`
	Type          string `json:"type"`
	Description   string `json:"description"`
	CurrentCount  int    `json:"currentCount"`
	TargetCount   int    `json:"targetCount"`
	RewardLeaves  int    `json:"rewardLeaves"`
	RequiredLevel int    `json:"requiredLevel"`
	Status        string `json:"status"`
}

type progressResponse struct {
	CompletedCount int `json:"completedCount"`
	TotalCount     int `json:"totalCount"`
}

var (
	cfgOnce sync.Once
	cfg     testConfig
	cfgErr  error
)

func TestTasksUnauthorized(t *testing.T) {
	cfg := getConfig(t)

	result := request(t, cfg, "", http.MethodGet, "/api/v1/tasks", nil)

	if result.status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401, body = %s", result.status, result.body)
	}
	if result.json["code"] != "UNAUTHORIZED" {
		t.Fatalf("code = %v, want UNAUTHORIZED", result.json["code"])
	}
}

func TestTasksHappyPath(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)

	tasks := getTasks(t, cfg, token, "/api/v1/tasks")
	if len(tasks.Tasks) != 4 {
		t.Fatalf("got %d tasks, want 4", len(tasks.Tasks))
	}

	slotOne := findTaskBySlot(t, tasks.Tasks, 1)
	if slotOne.RewardLeaves != 45 ||
		slotOne.RequiredLevel != 1 ||
		slotOne.Status != "COMPLETED" || slotOne.CurrentCount != slotOne.TargetCount {
		t.Fatalf("unexpected slot 1 task: %+v", slotOne)
	}

	slotTwo := findTaskBySlot(t, tasks.Tasks, 2)
	if slotTwo.RewardLeaves != 45 ||
		slotTwo.RequiredLevel != 1 ||
		slotTwo.Status != "COMPLETED" || slotTwo.CurrentCount != slotTwo.TargetCount {
		t.Fatalf("unexpected slot 2 task: %+v", slotTwo)
	}

	slotThree := findTaskBySlot(t, tasks.Tasks, 3)
	slotFour := findTaskBySlot(t, tasks.Tasks, 4)
	if slotThree.RewardLeaves != 50 || slotThree.RequiredLevel != 5 ||
		slotFour.RewardLeaves != 60 || slotFour.RequiredLevel != 10 ||
		slotThree.Status != "IN_PROGRESS" || slotFour.Status != "IN_PROGRESS" {
		t.Fatalf("unexpected tasks available to a level-ten pet: %+v / %+v", slotThree, slotFour)
	}

	recordBody := map[string]any{
		"events": []map[string]any{
			{"taskId": slotThree.TaskID, "type": slotThree.Type, "count": slotThree.TargetCount},
		},
	}
	record := request(t, cfg, token, http.MethodPost, "/api/v1/tasks/record", recordBody)
	if record.status != http.StatusNoContent {
		t.Fatalf("record status = %d, want 204, body = %s", record.status, record.body)
	}

	tasks = getTasks(t, cfg, token, "/api/v1/tasks")
	slotThree = findTaskBySlot(t, tasks.Tasks, 3)

	if slotThree.CurrentCount != slotThree.TargetCount || slotThree.Status != "COMPLETED" {
		t.Fatalf("slot 3 after record = %+v, want completed", slotThree)
	}

	progress := getProgress(t, cfg, token)
	if progress.CompletedCount != 3 || progress.TotalCount != 4 {
		t.Fatalf("progress = %+v, want 3/4", progress)
	}

	claim := request(t, cfg, token, http.MethodPost, "/api/v1/tasks/"+slotThree.TaskID+"/claim", nil)
	if claim.status != http.StatusOK {
		t.Fatalf("claim status = %d, want 200, body = %s", claim.status, claim.body)
	}
	if claim.json["taskId"] != slotThree.TaskID ||
		claim.json["rewardLeaves"] != float64(slotThree.RewardLeaves) ||
		claim.json["status"] != "CLAIMED" {
		t.Fatalf("unexpected claim response: %s", claim.body)
	}

	tasks = getTasks(t, cfg, token, "/api/v1/tasks")
	slotThree = findTaskBySlot(t, tasks.Tasks, 3)
	if slotThree.Status != "CLAIMED" {
		t.Fatalf("slot 3 status = %s, want CLAIMED", slotThree.Status)
	}

	secondClaim := request(t, cfg, token, http.MethodPost, "/api/v1/tasks/"+slotThree.TaskID+"/claim", nil)
	if secondClaim.status != http.StatusConflict {
		t.Fatalf("second claim status = %d, want 409, body = %s", secondClaim.status, secondClaim.body)
	}
	if secondClaim.json["code"] != "TASK_REWARD_ALREADY_CLAIMED" {
		t.Fatalf("second claim code = %v, want TASK_REWARD_ALREADY_CLAIMED", secondClaim.json["code"])
	}
}

func TestTasksRecordErrors(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)

	invalidType := request(t, cfg, token, http.MethodPost, "/api/v1/tasks/record", map[string]any{
		"events": []map[string]any{
			{"type": "UNKNOWN_TASK_TYPE", "count": 1},
		},
	})
	if invalidType.status != http.StatusBadRequest || invalidType.json["code"] != "INVALID_TASK_TYPE" {
		t.Fatalf("invalid type response = %d %s", invalidType.status, invalidType.body)
	}

	tasks := getTasks(t, cfg, token, "/api/v1/tasks")
	inProgress := findTaskBySlot(t, tasks.Tasks, 3)
	invalidBatch := request(t, cfg, token, http.MethodPost, "/api/v1/tasks/record", map[string]any{
		"events": []map[string]any{
			{"taskId": inProgress.TaskID, "type": inProgress.Type, "count": inProgress.TargetCount},
			{"type": "UNKNOWN_TASK_TYPE", "count": 1},
		},
	})
	if invalidBatch.status != http.StatusBadRequest || invalidBatch.json["code"] != "INVALID_TASK_TYPE" {
		t.Fatalf("invalid batch response = %d %s", invalidBatch.status, invalidBatch.body)
	}

	tasks = getTasks(t, cfg, token, "/api/v1/tasks")
	inProgress = findTaskBySlot(t, tasks.Tasks, 3)
	if inProgress.CurrentCount != 0 || inProgress.Status != "IN_PROGRESS" {
		t.Fatalf("slot 3 after failed batch = %+v, want rollback", inProgress)
	}
}

func TestTasksClaimErrors(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)

	tasks := getTasks(t, cfg, token, "/api/v1/tasks")
	incompleteTask := findTaskBySlot(t, tasks.Tasks, 3)

	badID := request(t, cfg, token, http.MethodPost, "/api/v1/tasks/not-a-uuid/claim", nil)
	if badID.status != http.StatusNotFound || badID.json["code"] != "TASK_NOT_FOUND" {
		t.Fatalf("bad id response = %d %s", badID.status, badID.body)
	}

	incomplete := request(t, cfg, token, http.MethodPost, "/api/v1/tasks/"+incompleteTask.TaskID+"/claim", nil)
	if incomplete.status != http.StatusConflict || incomplete.json["code"] != "TASK_NOT_COMPLETED" {
		t.Fatalf("incomplete claim response = %d %s", incomplete.status, incomplete.body)
	}
}

func TestTasksRecordNoOpAndBadJSON(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)

	noEvents := request(t, cfg, token, http.MethodPost, "/api/v1/tasks/record", map[string]any{})
	if noEvents.status != http.StatusNoContent {
		t.Fatalf("no events status = %d, want 204, body = %s", noEvents.status, noEvents.body)
	}

	nullEvents := request(t, cfg, token, http.MethodPost, "/api/v1/tasks/record", map[string]any{
		"events": nil,
	})
	if nullEvents.status != http.StatusNoContent {
		t.Fatalf("null events status = %d, want 204, body = %s", nullEvents.status, nullEvents.body)
	}

	progress := getProgress(t, cfg, token)
	if progress.CompletedCount != 2 || progress.TotalCount != 4 {
		t.Fatalf("progress = %+v, want 2/4 after no-op records and demo completion", progress)
	}

	badJSON := rawRequest(t, cfg, token, http.MethodPost, "/api/v1/tasks/record", "{")
	if badJSON.status != http.StatusBadRequest || badJSON.json["code"] != "INVALID_REQUEST" {
		t.Fatalf("bad json response = %d %s", badJSON.status, badJSON.body)
	}
}

func getConfig(t *testing.T) testConfig {
	t.Helper()
	if os.Getenv("RUN_BACKEND_E2E") != "1" {
		t.Skip("set RUN_BACKEND_E2E=1 to run backend e2e tests")
	}

	cfgOnce.Do(func() {
		cfg, cfgErr = prepareConfig()
	})

	if cfgErr != nil {
		t.Skipf("tasks API tests need running backend and compose postgres: %v", cfgErr)
	}

	return cfg
}

func prepareConfig() (testConfig, error) {
	repoRoot, err := filepath.Abs("../../..")
	if err != nil {
		return testConfig{}, err
	}

	dotenv, err := readEnvFile(filepath.Join(repoRoot, ".env"), filepath.Join(repoRoot, ".env.example"))
	if err != nil {
		return testConfig{}, err
	}

	cfg := testConfig{
		repoRoot:     repoRoot,
		apiURL:       getEnv(dotenv, "TASKS_API_BASE_URL", "http://127.0.0.1:8090"),
		postgresDB:   getEnv(dotenv, "POSTGRES_DB", "hackathon"),
		postgresUser: getEnv(dotenv, "POSTGRES_USER", "hackathon"),
		jwtSecret:    getEnv(dotenv, "JWT_SECRET", ""),
	}
	cfg.apiURL = strings.TrimRight(cfg.apiURL, "/")

	if len(cfg.jwtSecret) < 32 {
		return testConfig{}, errors.New("JWT_SECRET must be set and must be at least 32 characters")
	}

	if err := waitForBackend(cfg.apiURL); err != nil {
		return testConfig{}, err
	}

	if err := runSQL(cfg, "SELECT 1;"); err != nil {
		return testConfig{}, err
	}

	return cfg, nil
}

func waitForBackend(apiURL string) error {
	deadline := time.Now().Add(20 * time.Second)
	var lastErr error

	for time.Now().Before(deadline) {
		response, err := http.Get(apiURL + "/api/health")
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
			lastErr = fmt.Errorf("health returned %d", response.StatusCode)
		} else {
			lastErr = err
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("backend is not ready, run `make up` first: %w", lastErr)
}

func createUser(t *testing.T, cfg testConfig) uuid.UUID {
	t.Helper()

	userID := uuid.New()
	email := "tasks-api-" + randomString(t) + "@example.com"

	sql := fmt.Sprintf(`
INSERT INTO users (id, email, verified, created_at, updated_at)
VALUES (%s, %s, true, NOW(), NOW());
`, sqlUUID(userID), sqlString(email))
	if err := runSQL(cfg, sql); err != nil {
		t.Fatalf("create test user: %v", err)
	}

	t.Cleanup(func() {
		sql := fmt.Sprintf(`
DELETE FROM external_events WHERE user_id = %s;
DELETE FROM user_daily_tasks WHERE user_id = %s;
DELETE FROM leaf_transactions WHERE user_id = %s;
DELETE FROM user_game_states WHERE user_id = %s;
DELETE FROM leaderboard_entries WHERE user_id = %s;
DELETE FROM rewards WHERE user_id = %s;
DELETE FROM otps WHERE user_id = %s;
DELETE FROM pets WHERE user_id = %s;
DELETE FROM users WHERE id = %s;
`, sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID))
		if err := runSQL(cfg, sql); err != nil {
			t.Fatalf("cleanup test user: %v", err)
		}
	})

	return userID
}

func runSQL(cfg testConfig, sql string) error {
	command := exec.Command(
		"docker",
		"compose",
		"exec",
		"-T",
		"postgres",
		"psql",
		"-q",
		"-v",
		"ON_ERROR_STOP=1",
		"-U",
		cfg.postgresUser,
		"-d",
		cfg.postgresDB,
	)
	command.Dir = cfg.repoRoot
	command.Stdin = strings.NewReader(sql)

	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func makeToken(t *testing.T, cfg testConfig, userID uuid.UUID) string {
	t.Helper()

	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "avito-hackathon-api",
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
	})

	signed, err := token.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return signed
}

func getTasks(t *testing.T, cfg testConfig, token, path string) tasksResponse {
	t.Helper()

	result := request(t, cfg, token, http.MethodGet, path, nil)
	if result.status != http.StatusOK {
		t.Fatalf("get tasks status = %d, want 200, body = %s", result.status, result.body)
	}

	var body tasksResponse
	if err := json.Unmarshal(result.body, &body); err != nil {
		t.Fatalf("decode tasks response: %v, body = %s", err, result.body)
	}

	return body
}

func getProgress(t *testing.T, cfg testConfig, token string) progressResponse {
	t.Helper()

	result := request(t, cfg, token, http.MethodGet, "/api/v1/tasks/progress", nil)
	if result.status != http.StatusOK {
		t.Fatalf("get progress status = %d, want 200, body = %s", result.status, result.body)
	}

	var body progressResponse
	if err := json.Unmarshal(result.body, &body); err != nil {
		t.Fatalf("decode progress response: %v, body = %s", err, result.body)
	}

	return body
}

func request(t *testing.T, cfg testConfig, token, method, path string, body any) apiResult {
	t.Helper()

	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}

	return rawRequest(t, cfg, token, method, path, string(data))
}

func rawRequest(t *testing.T, cfg testConfig, token, method, path, body string) apiResult {
	t.Helper()

	request, err := http.NewRequest(method, cfg.apiURL+path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("%s %s failed: %v", method, path, err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	result := apiResult{status: response.StatusCode, body: responseBody}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &result.json); err != nil {
			t.Fatalf("decode json body: %v, body = %s", err, responseBody)
		}
	}

	return result
}

func findTaskBySlot(t *testing.T, tasks []taskResponse, slot int) taskResponse {
	t.Helper()

	for _, task := range tasks {
		if task.Slot == slot {
			return task
		}
	}

	t.Fatalf("task in slot %d not found in %+v", slot, tasks)
	return taskResponse{}
}

func readEnvFile(paths ...string) (map[string]string, error) {
	values := map[string]string{}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.Trim(strings.TrimSpace(value), `"'`)

			if _, exists := values[key]; !exists {
				values[key] = value
			}
		}
	}

	return values, nil
}

func getEnv(dotenv map[string]string, key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if value := dotenv[key]; value != "" {
		return value
	}
	return fallback
}

func randomString(t *testing.T) string {
	t.Helper()

	data := make([]byte, 8)
	if _, err := rand.Read(data); err != nil {
		t.Fatalf("random string: %v", err)
	}
	return hex.EncodeToString(data)
}

func sqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func sqlUUID(value uuid.UUID) string {
	return sqlString(value.String()) + "::uuid"
}
