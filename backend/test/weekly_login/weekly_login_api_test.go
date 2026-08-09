package weekly_login_test

import (
	"bytes"
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

type weeklyLoginDay struct {
	Weekday      int    `json:"weekday"`
	Date         string `json:"date"`
	Status       string `json:"status"`
	RewardLeaves int    `json:"rewardLeaves"`
	ClaimID      string `json:"claimId"`
}

type weeklyLoginResponse struct {
	ClaimedDaysCount int              `json:"claimedDaysCount"`
	Claims           []weeklyLoginDay `json:"claims"`
}

type weeklyLoginClaim struct {
	ID           string `json:"id"`
	Weekday      int    `json:"weekday"`
	Date         string `json:"date"`
	Status       string `json:"status"`
	RewardLeaves int    `json:"rewardLeaves"`
}

type weeklyLoginClaimResponse struct {
	Claim weeklyLoginClaim `json:"claim"`
}

type petResponse struct {
	Level  int   `json:"level"`
	Leaves int64 `json:"leaves"`
}

var (
	cfgOnce sync.Once
	cfg     testConfig
	cfgErr  error
)

func TestWeeklyLoginActivityGetAndClaim(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)

	available := recordActivityAndGetAvailableDay(t, cfg, token, 1)
	claim := claimAvailableDay(t, cfg, token, available)

	week := getWeeklyLogin(t, cfg, token)
	if week.ClaimedDaysCount != 1 {
		t.Fatalf("claimedDaysCount after claim = %d, want 1", week.ClaimedDaysCount)
	}

	claimed := findDayByDate(t, week.Claims, available.Date)
	if claimed.Status != "CLAIMED" || claimed.RewardLeaves != available.RewardLeaves || claimed.ClaimID != claim.ID {
		t.Fatalf("day after claim = %+v, want CLAIMED with reward %d and claimId %s", claimed, available.RewardLeaves, claim.ID)
	}

	petResult := request(t, cfg, token, http.MethodGet, "/api/v1/pet", nil)
	if petResult.status != http.StatusOK {
		t.Fatalf("pet status after weekly claim = %d, want 200, body = %s", petResult.status, petResult.body)
	}
	var pet petResponse
	decode(t, petResult.body, &pet)
	if pet.Level != 1 || pet.Leaves != int64(available.RewardLeaves) {
		t.Fatalf("pet after weekly claim = %+v, want level 1 with %d leaves", pet, available.RewardLeaves)
	}
}

func TestWeeklyLoginActivityIsIdempotentAndClaimCannotBeRepeated(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)

	available := recordActivityAndGetAvailableDay(t, cfg, token, 2)
	claimAvailableDay(t, cfg, token, available)

	repeatedClaim := request(t, cfg, token, http.MethodPost, "/api/v1/weekly-login/claim", nil)
	if repeatedClaim.status != http.StatusConflict {
		t.Fatalf("repeated claim status = %d, want 409, body = %s", repeatedClaim.status, repeatedClaim.body)
	}
	if repeatedClaim.json["code"] != "WEEKLY_LOGIN_REWARD_ALREADY_CLAIMED" {
		t.Fatalf("repeated claim code = %v, want WEEKLY_LOGIN_REWARD_ALREADY_CLAIMED", repeatedClaim.json["code"])
	}
}

func recordActivityAndGetAvailableDay(t *testing.T, cfg testConfig, token string, activityCalls int) weeklyLoginDay {
	t.Helper()

	dateBeforeRequests := time.Now().UTC().Format(time.DateOnly)
	for range activityCalls {
		activity := request(t, cfg, token, http.MethodPost, "/api/v1/weekly-login/activity", nil)
		if activity.status != http.StatusNoContent {
			t.Fatalf("activity status = %d, want 204, body = %s", activity.status, activity.body)
		}
		if len(activity.body) != 0 {
			t.Fatalf("activity response body = %q, want empty body", activity.body)
		}
	}

	week := getWeeklyLogin(t, cfg, token)
	if week.ClaimedDaysCount != 0 {
		t.Fatalf("claimedDaysCount before claim = %d, want 0", week.ClaimedDaysCount)
	}
	if len(week.Claims) != 7 {
		t.Fatalf("weekly login returned %d days, want 7", len(week.Claims))
	}

	availableDays := make([]weeklyLoginDay, 0, 1)
	for _, day := range week.Claims {
		if day.Status == "AVAILABLE" {
			availableDays = append(availableDays, day)
		}
	}
	if len(availableDays) != 1 {
		t.Fatalf("available days before claim = %+v, want exactly one", availableDays)
	}

	available := availableDays[0]
	dateAfterRequests := time.Now().UTC().Format(time.DateOnly)
	if available.Date != dateBeforeRequests && available.Date != dateAfterRequests {
		t.Fatalf("available date = %s, want current UTC date %s or %s", available.Date, dateBeforeRequests, dateAfterRequests)
	}
	if available.Weekday < 1 || available.Weekday > 7 {
		t.Fatalf("available weekday = %d, want value from 1 to 7", available.Weekday)
	}
	if available.RewardLeaves != 10 || available.ClaimID != "" {
		t.Fatalf("available day = %+v, want first 10-leaf reward without claimId", available)
	}

	return available
}

func claimAvailableDay(t *testing.T, cfg testConfig, token string, available weeklyLoginDay) weeklyLoginClaim {
	t.Helper()

	result := request(t, cfg, token, http.MethodPost, "/api/v1/weekly-login/claim", nil)
	if result.status != http.StatusOK {
		t.Fatalf("claim status = %d, want 200, body = %s", result.status, result.body)
	}

	var response weeklyLoginClaimResponse
	decode(t, result.body, &response)
	claim := response.Claim
	if _, err := uuid.Parse(claim.ID); err != nil {
		t.Fatalf("claim id = %q, want UUID: %v", claim.ID, err)
	}
	if claim.Date != available.Date || claim.Weekday != available.Weekday ||
		claim.Status != "CLAIMED" || claim.RewardLeaves != available.RewardLeaves {
		t.Fatalf("claim = %+v, want claimed available day %+v", claim, available)
	}

	return claim
}

func getWeeklyLogin(t *testing.T, cfg testConfig, token string) weeklyLoginResponse {
	t.Helper()

	result := request(t, cfg, token, http.MethodGet, "/api/v1/weekly-login", nil)
	if result.status != http.StatusOK {
		t.Fatalf("weekly login status = %d, want 200, body = %s", result.status, result.body)
	}

	var response weeklyLoginResponse
	decode(t, result.body, &response)

	return response
}

func findDayByDate(t *testing.T, days []weeklyLoginDay, date string) weeklyLoginDay {
	t.Helper()

	for _, day := range days {
		if day.Date == date {
			return day
		}
	}

	t.Fatalf("weekly login day for %s not found in %+v", date, days)
	return weeklyLoginDay{}
}

func getConfig(t *testing.T) testConfig {
	t.Helper()

	cfgOnce.Do(func() {
		cfg, cfgErr = prepareConfig()
	})
	if cfgErr != nil {
		t.Skipf("weekly login API tests need running backend and Compose Postgres: %v", cfgErr)
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
		apiURL:       envOr(dotenv, "WEEKLY_LOGIN_API_BASE_URL", "http://127.0.0.1:8090"),
		postgresDB:   envOr(dotenv, "POSTGRES_DB", "hackathon"),
		postgresUser: envOr(dotenv, "POSTGRES_USER", "hackathon"),
		jwtSecret:    envOr(dotenv, "JWT_SECRET", ""),
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

func createUser(t *testing.T, cfg testConfig) uuid.UUID {
	t.Helper()

	userID := uuid.New()
	email := "weekly-login-api-" + userID.String() + "@example.com"
	statement := fmt.Sprintf(
		"INSERT INTO users (id, email, verified, created_at, updated_at) VALUES (%s, %s, true, NOW(), NOW());",
		sqlUUID(userID),
		sqlString(email),
	)
	if err := runSQL(cfg, statement); err != nil {
		t.Fatalf("create weekly login test user: %v", err)
	}

	t.Cleanup(func() {
		cleanup := fmt.Sprintf(`
DELETE FROM weekly_login_claims WHERE user_id = %s;
DELETE FROM user_logins WHERE user_id = %s;
DELETE FROM leaf_transactions WHERE user_id = %s;
DELETE FROM user_game_states WHERE user_id = %s;
DELETE FROM pets WHERE user_id = %s;
DELETE FROM users WHERE id = %s;
`, sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID))
		if err := runSQL(cfg, cleanup); err != nil {
			t.Errorf("cleanup weekly login test user: %v", err)
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

func runSQL(cfg testConfig, statement string) error {
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
				values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
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
	return "'" + value.String() + "'::uuid"
}

func sqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
