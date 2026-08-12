package daily_report_test

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
	"sync"
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

type dailyReportResponse struct {
	LeavesEarnedToday int                 `json:"leavesEarnedToday"`
	Date              string              `json:"date"`
	Rewards           []dailyReportReward `json:"rewards"`
	LevelUp           *dailyReportLevelUp `json:"levelUp"`
	Tasks             []dailyReportTask   `json:"tasks"`
	VisitedToday      bool                `json:"visitedToday"`
	UpdatedAt         time.Time           `json:"updatedAt"`
}

type dailyReportReward struct {
	RewardID   string    `json:"rewardId"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	ExpiresAt  string    `json:"expiresAt"`
	ReceivedAt time.Time `json:"receivedAt"`
}

type dailyReportTask struct {
	TaskID       string    `json:"taskId"`
	Type         string    `json:"type"`
	Description  string    `json:"description"`
	RewardLeaves int       `json:"rewardLeaves"`
	CompletedAt  time.Time `json:"completedAt"`
}

type dailyReportLevelUp struct {
	FromLevel  int       `json:"fromLevel"`
	ToLevel    int       `json:"toLevel"`
	OccurredAt time.Time `json:"occurredAt"`
}

type dailyReportEvent struct {
	Event string              `json:"event"`
	Data  dailyReportResponse `json:"data"`
}

type weeklyLoginClaimResponse struct {
	Claim struct {
		RewardLeaves int `json:"rewardLeaves"`
	} `json:"claim"`
}

var (
	cfgOnce sync.Once
	cfg     testConfig
	cfgErr  error
)

func TestDailyReportAuthorizationAndEmptySnapshot(t *testing.T) {
	cfg := getConfig(t)

	unauthorized := request(t, cfg, "", http.MethodGet, "/api/v1/daily-report", nil)
	if unauthorized.status != http.StatusUnauthorized {
		t.Fatalf("unauthorized daily report status = %d, want 401, body = %s", unauthorized.status, unauthorized.body)
	}
	if unauthorized.json["code"] != "UNAUTHORIZED" {
		t.Fatalf("unauthorized daily report code = %v, want UNAUTHORIZED", unauthorized.json["code"])
	}

	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)
	dateBeforeRequest := time.Now().UTC().Format(time.DateOnly)
	report := getDailyReport(t, cfg, token)
	dateAfterRequest := time.Now().UTC().Format(time.DateOnly)

	if report.LeavesEarnedToday != 0 || report.LevelUp != nil || report.VisitedToday {
		t.Fatalf("empty daily report = %+v", report)
	}
	if report.Date != dateBeforeRequest && report.Date != dateAfterRequest {
		t.Fatalf("daily report date = %q, want current UTC date %q or %q", report.Date, dateBeforeRequest, dateAfterRequest)
	}
	if report.Rewards == nil || len(report.Rewards) != 0 {
		t.Fatalf("empty daily report rewards = %+v, want non-nil empty array", report.Rewards)
	}
	if report.Tasks == nil || len(report.Tasks) != 0 {
		t.Fatalf("empty daily report tasks = %+v, want non-nil empty array", report.Tasks)
	}
	wantUpdatedAt, err := time.Parse(time.DateOnly, report.Date)
	if err != nil {
		t.Fatalf("parse daily report date %q: %v", report.Date, err)
	}
	if !report.UpdatedAt.Equal(wantUpdatedAt) {
		t.Fatalf("empty daily report updatedAt = %s, want UTC day start %s", report.UpdatedAt, wantUpdatedAt)
	}
}

func TestDailyReportWebSocketSendsInitialAndUpdatedSnapshots(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)
	pet := request(t, cfg, token, http.MethodGet, "/api/v1/pet", nil)
	if pet.status != http.StatusOK {
		t.Fatalf("initial pet status = %d, want 200, body = %s", pet.status, pet.body)
	}

	connection := openDailyReportWebSocket(t, cfg, token)
	defer func() {
		_ = connection.Close()
	}()

	initial := readDailyReportEvent(t, connection)
	if initial.Event != "DAILY_REPORT_UPDATED" || initial.Data.LeavesEarnedToday != 0 ||
		initial.Data.VisitedToday || initial.Data.LevelUp != nil || initial.Data.Rewards == nil ||
		len(initial.Data.Rewards) != 0 || initial.Data.Tasks == nil || len(initial.Data.Tasks) != 0 {
		t.Fatalf("initial daily report WebSocket event = %+v, want empty snapshot", initial)
	}

	activity := request(t, cfg, token, http.MethodPost, "/api/v1/weekly-login/activity", nil)
	if activity.status != http.StatusNoContent {
		t.Fatalf("weekly login activity status = %d, want 204, body = %s", activity.status, activity.body)
	}
	activityUpdate := readDailyReportEvent(t, connection)
	if activityUpdate.Event != "DAILY_REPORT_UPDATED" || !activityUpdate.Data.VisitedToday ||
		activityUpdate.Data.LeavesEarnedToday != 0 || activityUpdate.Data.Rewards == nil || activityUpdate.Data.Tasks == nil {
		t.Fatalf("daily report activity update = %+v, want full visited snapshot", activityUpdate)
	}

	weeklyClaimResult := request(t, cfg, token, http.MethodPost, "/api/v1/weekly-login/claim", nil)
	if weeklyClaimResult.status != http.StatusOK {
		t.Fatalf("weekly login claim status = %d, want 200, body = %s", weeklyClaimResult.status, weeklyClaimResult.body)
	}
	var weeklyClaim weeklyLoginClaimResponse
	decode(t, weeklyClaimResult.body, &weeklyClaim)

	claimUpdate := readDailyReportEvent(t, connection)
	if claimUpdate.Event != "DAILY_REPORT_UPDATED" || !claimUpdate.Data.VisitedToday ||
		claimUpdate.Data.LeavesEarnedToday != weeklyClaim.Claim.RewardLeaves || claimUpdate.Data.Rewards == nil ||
		claimUpdate.Data.Tasks == nil || claimUpdate.Data.LevelUp != nil {
		t.Fatalf("daily report claim update = %+v, want full snapshot with %d leaves", claimUpdate, weeklyClaim.Claim.RewardLeaves)
	}
	if activityUpdate.Data.UpdatedAt.Before(initial.Data.UpdatedAt) || claimUpdate.Data.UpdatedAt.Before(activityUpdate.Data.UpdatedAt) {
		t.Fatalf("daily report WebSocket updatedAt is not monotonic: %s, %s, %s",
			initial.Data.UpdatedAt, activityUpdate.Data.UpdatedAt, claimUpdate.Data.UpdatedAt)
	}
}

func getConfig(t *testing.T) testConfig {
	t.Helper()
	if os.Getenv("RUN_API_SERVICE_E2E") != "1" {
		t.Skip("set RUN_API_SERVICE_E2E=1 to run api-service e2e tests")
	}

	cfgOnce.Do(func() {
		cfg, cfgErr = prepareConfig()
	})
	if cfgErr != nil {
		t.Skipf("daily report API tests need running api-service and Compose Postgres: %v", cfgErr)
	}

	return cfg
}

func prepareConfig() (testConfig, error) {
	repoRoot, err := filepath.Abs("../../..")
	if err != nil {
		return testConfig{}, err
	}

	dotenv := readEnv(filepath.Join(repoRoot, ".env"), filepath.Join(repoRoot, ".env.example"))
	cfg := testConfig{
		repoRoot:     repoRoot,
		apiURL:       envOr(dotenv, "DAILY_REPORT_API_BASE_URL", "http://127.0.0.1:8090"),
		postgresDB:   envOr(dotenv, "POSTGRES_DB", "hackathon"),
		postgresUser: envOr(dotenv, "POSTGRES_USER", "hackathon"),
	}
	cfg.apiURL = strings.TrimRight(cfg.apiURL, "/")

	if err := waitForAPIService(cfg.apiURL); err != nil {
		return testConfig{}, err
	}
	if err := runSQL(cfg, "SELECT 1;"); err != nil {
		return testConfig{}, err
	}
	jwtSecret, err := apiServiceJWTSecret(cfg)
	if err != nil {
		return testConfig{}, err
	}
	cfg.jwtSecret = jwtSecret

	return cfg, nil
}

func createUser(t *testing.T, cfg testConfig) uuid.UUID {
	t.Helper()

	userID := uuid.New()
	email := "daily-report-api-" + userID.String() + "@example.com"
	statement := fmt.Sprintf(
		"INSERT INTO users (id, email, verified, created_at, updated_at) VALUES (%s, %s, true, NOW(), NOW());",
		sqlUUID(userID),
		sqlString(email),
	)
	if err := runSQL(cfg, statement); err != nil {
		t.Fatalf("create daily report test user: %v", err)
	}

	t.Cleanup(func() {
		cleanup := fmt.Sprintf(`
DELETE FROM external_events WHERE user_id = %s;
DELETE FROM user_daily_tasks WHERE user_id = %s;
DELETE FROM weekly_login_claims WHERE user_id = %s;
DELETE FROM user_logins WHERE user_id = %s;
DELETE FROM leaf_transactions WHERE user_id = %s;
DELETE FROM leaderboard_entries WHERE user_id = %s;
DELETE FROM rewards WHERE user_id = %s;
DELETE FROM level_rewards WHERE user_id = %s;
DELETE FROM chest_openings WHERE user_id = %s;
DELETE FROM otps WHERE user_id = %s;
DELETE FROM pets WHERE user_id = %s;
DELETE FROM users WHERE id = %s;
`, sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID),
			sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID))
		if err := runSQL(cfg, cleanup); err != nil {
			t.Errorf("cleanup daily report test user: %v", err)
		}
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

func getDailyReport(t *testing.T, cfg testConfig, token string) dailyReportResponse {
	t.Helper()

	result := request(t, cfg, token, http.MethodGet, "/api/v1/daily-report", nil)
	if result.status != http.StatusOK {
		t.Fatalf("daily report status = %d, want 200, body = %s", result.status, result.body)
	}

	var report dailyReportResponse
	decode(t, result.body, &report)

	return report
}

func openDailyReportWebSocket(t *testing.T, cfg testConfig, token string) *websocket.Conn {
	t.Helper()

	connection, response, err := websocket.DefaultDialer.Dial(dailyReportWebSocketURL(t, cfg, token), nil)
	if err != nil {
		if response != nil {
			t.Fatalf("open daily report WebSocket: %v, status %d", err, response.StatusCode)
		}
		t.Fatalf("open daily report WebSocket: %v", err)
	}

	return connection
}

func readDailyReportEvent(t *testing.T, connection *websocket.Conn) dailyReportEvent {
	t.Helper()

	if err := connection.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set daily report WebSocket read deadline: %v", err)
	}
	var event dailyReportEvent
	if err := connection.ReadJSON(&event); err != nil {
		t.Fatalf("read daily report WebSocket event: %v", err)
	}

	return event
}

func dailyReportWebSocketURL(t *testing.T, cfg testConfig, token string) string {
	t.Helper()

	parsed, err := url.Parse(cfg.apiURL + "/api/v1/daily-report/ws")
	if err != nil {
		t.Fatalf("parse daily report WebSocket URL: %v", err)
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

	request, err := http.NewRequest(method, cfg.apiURL+path, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if len(payload) > 0 {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := http.DefaultClient.Do(request)
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
			t.Fatalf("decode response: %v, body = %s", err, responseBody)
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

func waitForAPIService(apiURL string) error {
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

	return fmt.Errorf("api-service is not ready: %w", lastErr)
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

func apiServiceJWTSecret(cfg testConfig) (string, error) {
	command := exec.Command("docker", "compose", "exec", "-T", "api-service", "printenv", "JWT_SECRET")
	command.Dir = cfg.repoRoot
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("read JWT_SECRET from running api-service: %w", err)
	}

	secret := strings.TrimSpace(string(output))
	if len(secret) < 32 {
		return "", errors.New("running api-service JWT_SECRET must be at least 32 characters")
	}

	return secret, nil
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
