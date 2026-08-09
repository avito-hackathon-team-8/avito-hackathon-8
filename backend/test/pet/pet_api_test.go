package pet_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

type petResponse struct {
	Name                  string `json:"name"`
	Level                 int    `json:"level"`
	Leaves                int64  `json:"leaves"`
	NextLevelTargetLeaves int64  `json:"nextLevelTargetLeaves"`
	ChestPrice            int64  `json:"chestPrice"`
	LevelUp               bool   `json:"levelUp"`
}

type petProgressEvent struct {
	Event string `json:"event"`
	Data  struct {
		Name                  string `json:"name"`
		Level                 int    `json:"level"`
		Leaves                int64  `json:"leaves"`
		NextLevelTargetLeaves int64  `json:"nextLevelTargetLeaves"`
		ChestPrice            int64  `json:"chestPrice"`
		LevelUp               bool   `json:"levelUp"`
	} `json:"data"`
}

type taskItem struct {
	TaskID       string `json:"taskId"`
	Slot         int    `json:"slot"`
	RewardLeaves int    `json:"rewardLeaves"`
	TargetCount  int    `json:"targetCount"`
}

type taskListResponse struct {
	Tasks []taskItem `json:"tasks"`
}

func TestPetLifecycleAndTaskRewardWebSocket(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)

	initial := request(t, cfg, token, http.MethodGet, "/api/v1/pet", nil)
	if initial.status != http.StatusOK {
		t.Fatalf("initial pet status = %d, body = %s", initial.status, initial.body)
	}
	if _, exists := initial.json["targetLeaves"]; exists {
		t.Fatalf("initial pet response contains removed targetLeaves field: %s", initial.body)
	}
	var pet petResponse
	decode(t, initial.body, &pet)
	if pet.Name != "" || pet.Level != 10 || pet.Leaves != 1000 || pet.NextLevelTargetLeaves != 0 || pet.ChestPrice != 200 || pet.LevelUp {
		t.Fatalf("initial pet = %+v, want empty level-ten pet with 1000 leaves", pet)
	}

	invalidName := request(t, cfg, token, http.MethodPatch, "/api/v1/pet", map[string]any{"name": "   "})
	if invalidName.status != http.StatusBadRequest {
		t.Fatalf("invalid name status = %d, want 400", invalidName.status)
	}

	updated := request(t, cfg, token, http.MethodPatch, "/api/v1/pet", map[string]any{"name": "  Листик  "})
	if updated.status != http.StatusOK {
		t.Fatalf("update name status = %d, body = %s", updated.status, updated.body)
	}
	if _, exists := updated.json["targetLeaves"]; exists {
		t.Fatalf("updated pet response contains removed targetLeaves field: %s", updated.body)
	}
	decode(t, updated.body, &pet)
	if pet.Name != "Листик" || pet.Level != 10 || pet.Leaves != 1000 || pet.NextLevelTargetLeaves != 0 || pet.ChestPrice != 200 || pet.LevelUp {
		t.Fatalf("updated pet = %+v, want renamed level-ten pet with 1000 leaves", pet)
	}

	connection := openPetWebSocket(t, cfg, token)
	defer func() {
		_ = connection.Close()
	}()

	initialEvent := readPetEvent(t, connection)
	if initialEvent.Event != "PET_PROGRESS_UPDATED" || initialEvent.Data.Name != "Листик" ||
		initialEvent.Data.Level != 10 || initialEvent.Data.Leaves != 1000 ||
		initialEvent.Data.NextLevelTargetLeaves != 0 || initialEvent.Data.ChestPrice != 200 || initialEvent.Data.LevelUp {
		t.Fatalf("initial WebSocket event = %+v, want current pet snapshot", initialEvent)
	}
	if err := connection.WriteJSON(map[string]string{"action": "GET_CHEST_PRICE"}); err != nil {
		t.Fatalf("request chest price over WebSocket: %v", err)
	}
	chestEvent := readPetEvent(t, connection)
	if chestEvent.Data.ChestPrice != 200 || chestEvent.Data.NextLevelTargetLeaves != 0 {
		t.Fatalf("chest price WebSocket event = %+v, want price 200 and target 0", chestEvent)
	}

	task := getFirstTask(t, cfg, token)
	seedCompletedTask(t, cfg, userID, task.TaskID, task.TargetCount)

	claim := request(t, cfg, token, http.MethodPost, "/api/v1/tasks/"+task.TaskID+"/claim", map[string]any{})
	if claim.status != http.StatusOK {
		t.Fatalf("claim status = %d, body = %s", claim.status, claim.body)
	}

	update := readPetEvent(t, connection)
	if update.Event != "PET_PROGRESS_UPDATED" || update.Data.Name != "Листик" ||
		update.Data.Level != 10 || update.Data.Leaves != 1000+int64(task.RewardLeaves) ||
		update.Data.NextLevelTargetLeaves != 0 || update.Data.ChestPrice != 200 || update.Data.LevelUp {
		t.Fatalf("reward WebSocket event = %+v, want %d leaves at level 10", update, 1000+int64(task.RewardLeaves))
	}
}

func TestPetAuthorizationAndNameSetupOrder(t *testing.T) {
	cfg := getConfig(t)
	unauthorized := request(t, cfg, "", http.MethodGet, "/api/v1/pet", nil)
	if unauthorized.status != http.StatusUnauthorized {
		t.Fatalf("unauthorized pet status = %d, want 401", unauthorized.status)
	}

	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)
	missingPet := request(t, cfg, token, http.MethodPatch, "/api/v1/pet", map[string]any{"name": "Листик"})
	if missingPet.status != http.StatusNotFound {
		t.Fatalf("name update before initial GET status = %d, want 404", missingPet.status)
	}

	created := request(t, cfg, token, http.MethodGet, "/api/v1/pet", nil)
	if created.status != http.StatusOK {
		t.Fatalf("initial pet status = %d, body = %s", created.status, created.body)
	}
}

func TestPetWebSocketReportsLevelUp(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)
	if result := request(t, cfg, token, http.MethodGet, "/api/v1/pet", nil); result.status != http.StatusOK {
		t.Fatalf("initial pet status = %d, body = %s", result.status, result.body)
	}
	seedPetProgress(t, cfg, userID, 1, 70)

	connection := openPetWebSocket(t, cfg, token)
	defer func() {
		_ = connection.Close()
	}()
	initial := readPetEvent(t, connection)
	if initial.Data.Level != 1 || initial.Data.Leaves != 70 || initial.Data.NextLevelTargetLeaves != 100 ||
		initial.Data.ChestPrice != 200 {
		t.Fatalf("initial level-up snapshot = %+v, want level 1 with 70 leaves", initial)
	}

	task := getFirstTask(t, cfg, token)
	seedCompletedTask(t, cfg, userID, task.TaskID, task.TargetCount)
	claim := request(t, cfg, token, http.MethodPost, "/api/v1/tasks/"+task.TaskID+"/claim", map[string]any{})
	if claim.status != http.StatusOK {
		t.Fatalf("claim status = %d, body = %s", claim.status, claim.body)
	}

	update := readPetEvent(t, connection)
	wantLeaves := int64(task.RewardLeaves - 30)
	if update.Data.Level != 2 || update.Data.Leaves != wantLeaves || update.Data.NextLevelTargetLeaves != 130 ||
		update.Data.ChestPrice != 200 || !update.Data.LevelUp {
		t.Fatalf("level-up WebSocket event = %+v, want level 2, leaves %d, target 130, and level up", update, wantLeaves)
	}
}

func TestPetWebSocketRejectsUnauthorizedClient(t *testing.T) {
	cfg := getConfig(t)
	wsURL := websocketURL(t, cfg, "")
	connection, response, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if connection != nil {
		_ = connection.Close()
	}
	if err == nil {
		t.Fatal("WebSocket connection without token succeeded")
	}
	if response == nil || response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized WebSocket response = %+v, want 401", response)
	}
}

func getFirstTask(t *testing.T, cfg testConfig, token string) taskItem {
	t.Helper()
	result := request(t, cfg, token, http.MethodGet, "/api/v1/tasks", nil)
	if result.status != http.StatusOK {
		t.Fatalf("tasks status = %d, body = %s", result.status, result.body)
	}

	var response taskListResponse
	decode(t, result.body, &response)
	for _, task := range response.Tasks {
		if task.Slot == 1 {
			return task
		}
	}
	t.Fatal("slot 1 task was not returned")
	return taskItem{}
}

func seedCompletedTask(t *testing.T, cfg testConfig, userID uuid.UUID, taskID string, targetCount int) {
	t.Helper()
	statement := fmt.Sprintf(
		"UPDATE user_daily_tasks SET current_count = %d, completed_at = NOW(), status = 'COMPLETED', updated_at = NOW() WHERE id = %s AND user_id = %s;",
		targetCount,
		sqlString(taskID),
		sqlUUID(userID),
	)
	if err := runSQL(cfg, statement); err != nil {
		t.Fatalf("seed completed task: %v", err)
	}
}

func seedPetProgress(t *testing.T, cfg testConfig, userID uuid.UUID, level int, leaves int64) {
	t.Helper()
	statement := fmt.Sprintf("UPDATE pets SET level = %d, leaves = %d WHERE user_id = %s;", level, leaves, sqlUUID(userID))
	if err := runSQL(cfg, statement); err != nil {
		t.Fatalf("seed pet progress: %v", err)
	}
}

func openPetWebSocket(t *testing.T, cfg testConfig, token string) *websocket.Conn {
	t.Helper()
	connection, response, err := websocket.DefaultDialer.Dial(websocketURL(t, cfg, token), nil)
	if err != nil {
		if response != nil {
			t.Fatalf("open pet WebSocket: %v, status %d", err, response.StatusCode)
		}
		t.Fatalf("open pet WebSocket: %v", err)
	}
	return connection
}

func readPetEvent(t *testing.T, connection *websocket.Conn) petProgressEvent {
	t.Helper()
	if err := connection.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set WebSocket read deadline: %v", err)
	}
	var event petProgressEvent
	if err := connection.ReadJSON(&event); err != nil {
		t.Fatalf("read pet WebSocket event: %v", err)
	}
	return event
}

func getConfig(t *testing.T) testConfig {
	t.Helper()
	if os.Getenv("RUN_BACKEND_E2E") != "1" {
		t.Skip("set RUN_BACKEND_E2E=1 to run backend e2e tests")
	}

	repoRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}

	dotenv := readEnv(filepath.Join(repoRoot, ".env"), filepath.Join(repoRoot, ".env.example"))
	cfg := testConfig{
		repoRoot:     repoRoot,
		apiURL:       envOr(dotenv, "PET_API_BASE_URL", "http://127.0.0.1:8090"),
		postgresDB:   envOr(dotenv, "POSTGRES_DB", "hackathon"),
		postgresUser: envOr(dotenv, "POSTGRES_USER", "hackathon"),
		jwtSecret:    envOr(dotenv, "JWT_SECRET", ""),
	}
	if len(cfg.jwtSecret) < 32 {
		t.Skip("pet e2e tests require JWT_SECRET with at least 32 characters")
	}
	if err := waitForBackend(cfg.apiURL); err != nil {
		t.Skipf("pet e2e tests need a running backend: %v", err)
	}
	if err := runSQL(cfg, "SELECT 1;"); err != nil {
		t.Skipf("pet e2e tests need Compose Postgres: %v", err)
	}
	return cfg
}

func createUser(t *testing.T, cfg testConfig) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	email := "pet-e2e-" + userID.String() + "@example.com"
	sql := fmt.Sprintf("INSERT INTO users (id, email, verified, created_at, updated_at) VALUES (%s, %s, true, NOW(), NOW());", sqlUUID(userID), sqlString(email))
	if err := runSQL(cfg, sql); err != nil {
		t.Fatalf("create e2e user: %v", err)
	}
	t.Cleanup(func() {
		statement := fmt.Sprintf(`
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
		_ = runSQL(cfg, statement)
	})
	return userID
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

func request(t *testing.T, cfg testConfig, token, method, path string, body any) apiResult {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
	}
	req, err := http.NewRequest(method, cfg.apiURL+path, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if len(payload) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	result := apiResult{status: response.StatusCode, body: responseBody}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &result.json); err != nil {
			t.Fatalf("decode response: %v", err)
		}
	}
	return result
}

func decode(t *testing.T, body []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("decode JSON: %v, body = %s", err, body)
	}
}

func websocketURL(t *testing.T, cfg testConfig, token string) string {
	t.Helper()

	parsed, err := url.Parse(cfg.apiURL + "/api/v1/pet/ws")

	if err != nil {
		t.Fatalf("parse API URL: %v", err)
	}
	switch parsed.Scheme {
	case "http":
		parsed.Scheme = "ws"
	case "https":
		parsed.Scheme = "wss"
	}

	if token != "" {
		query := parsed.Query()
		query.Set("token", token)
		parsed.RawQuery = query.Encode()
	}
	return parsed.String()
}

func waitForBackend(apiURL string) error {
	deadline := time.Now().Add(10 * time.Second)
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
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("backend is not ready: %w", lastErr)
}

func runSQL(cfg testConfig, statement string) error {
	command := exec.Command("docker", "compose", "exec", "-T", "postgres", "psql", "-q", "-v", "ON_ERROR_STOP=1", "-U", cfg.postgresUser, "-d", cfg.postgresDB)
	command.Dir = cfg.repoRoot
	command.Stdin = strings.NewReader(statement)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func readEnv(paths ...string) map[string]string {
	values := make(map[string]string)
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if ok {
				key = strings.TrimSpace(key)
				if _, exists := values[key]; !exists {
					values[key] = strings.Trim(strings.TrimSpace(value), `"'`)
				}
			}
		}
	}
	return values
}

func envOr(values map[string]string, key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if value := values[key]; value != "" {
		return value
	}
	return fallback
}

func sqlUUID(value uuid.UUID) string {
	return "'" + value.String() + "'"
}

func sqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
