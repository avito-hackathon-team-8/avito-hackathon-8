package chest_test

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

type petResponse struct {
	Level  int   `json:"level"`
	Leaves int64 `json:"leaves"`
}

type rewardResponse struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Status   string `json:"status"`
	Active   bool   `json:"active"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

func TestOpenChestIssuesRewardAndSpendsLeaves(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg)
	token := makeToken(t, cfg, userID)

	unauthorized := request(t, cfg, "", http.MethodPost, "/api/v1/pet/chests/open", nil)
	if unauthorized.status != http.StatusUnauthorized {
		t.Fatalf("unauthorized chest opening status = %d, want 401", unauthorized.status)
	}

	initialPet := request(t, cfg, token, http.MethodGet, "/api/v1/pet", nil)
	if initialPet.status != http.StatusOK {
		t.Fatalf("get initial pet status = %d, body = %s", initialPet.status, initialPet.body)
	}
	var currentPet petResponse
	decode(t, initialPet.body, &currentPet)
	if currentPet.Level != 10 || currentPet.Leaves != 1000 {
		t.Fatalf("initial pet = %+v, want level 10 with 1000 leaves", currentPet)
	}

	for openingsLeft := 5; openingsLeft > 0; openingsLeft-- {
		opened := request(t, cfg, token, http.MethodPost, "/api/v1/pet/chests/open", nil)
		if opened.status != http.StatusOK {
			t.Fatalf("open chest with %d openings left status = %d, body = %s", openingsLeft, opened.status, opened.body)
		}
		if _, exists := opened.json["leaves"]; exists {
			t.Fatal("open chest response contains leaves, want only reward fields")
		}

		var reward rewardResponse
		decode(t, opened.body, &reward)
		if reward.ID == "" || reward.Source != "CHEST" || !reward.Active || reward.Status != "ACTIVE" || reward.Title == "" || reward.Category == "" {
			t.Fatalf("open chest reward = %+v, want active chest reward", reward)
		}

		rewards := request(t, cfg, token, http.MethodGet, "/api/app/rewards", nil)
		if rewards.status != http.StatusOK {
			t.Fatalf("list rewards status = %d, body = %s", rewards.status, rewards.body)
		}
		if !containsReward(rewards.json, reward.ID) {
			t.Fatalf("reward %s is not returned by GET /api/app/rewards", reward.ID)
		}
	}

	insufficient := request(t, cfg, token, http.MethodPost, "/api/v1/pet/chests/open", nil)
	if insufficient.status != http.StatusConflict {
		t.Fatalf("second chest opening status = %d, body = %s", insufficient.status, insufficient.body)
	}
	if insufficient.json["code"] != "INSUFFICIENT_LEAVES" {
		t.Fatalf("opening after balance is exhausted error = %v, want INSUFFICIENT_LEAVES", insufficient.json)
	}

	finalPet := request(t, cfg, token, http.MethodGet, "/api/v1/pet", nil)
	if finalPet.status != http.StatusOK {
		t.Fatalf("get final pet status = %d, body = %s", finalPet.status, finalPet.body)
	}
	decode(t, finalPet.body, &currentPet)
	if currentPet.Level != 10 || currentPet.Leaves != 0 {
		t.Fatalf("pet after five openings = %+v, want level 10 with 0 leaves", currentPet)
	}
}

func containsReward(response map[string]any, rewardID string) bool {
	groups, ok := response["groups"].([]any)
	if !ok {
		return false
	}

	for _, group := range groups {
		groupData, ok := group.(map[string]any)
		if !ok {
			continue
		}
		items, ok := groupData["items"].([]any)
		if !ok {
			continue
		}
		for _, item := range items {
			itemData, ok := item.(map[string]any)
			if ok && itemData["id"] == rewardID && itemData["source"] == "CHEST" {
				return true
			}
		}
	}

	return false
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
		apiURL:       envOr(dotenv, "CHEST_API_BASE_URL", "http://127.0.0.1:8090"),
		postgresDB:   envOr(dotenv, "POSTGRES_DB", "hackathon"),
		postgresUser: envOr(dotenv, "POSTGRES_USER", "hackathon"),
		jwtSecret:    envOr(dotenv, "JWT_SECRET", ""),
	}
	if err := waitForBackend(cfg.apiURL); err != nil {
		t.Skipf("chest e2e tests need a running backend: %v", err)
	}
	if err := runSQL(cfg, "SELECT 1;"); err != nil {
		t.Skipf("chest e2e tests need Compose Postgres: %v", err)
	}
	secret, err := backendJWTSecret(cfg)
	if err != nil {
		t.Skipf("chest e2e tests need a backend JWT secret: %v", err)
	}
	cfg.jwtSecret = secret

	return cfg
}

func createUser(t *testing.T, cfg testConfig) uuid.UUID {
	t.Helper()

	userID := uuid.New()
	email := "chest-e2e-" + userID.String() + "@example.com"
	statement := fmt.Sprintf(
		"INSERT INTO users (id, email, verified, created_at, updated_at) VALUES (%s, %s, true, NOW(), NOW());",
		sqlUUID(userID),
		sqlString(email),
	)
	if err := runSQL(cfg, statement); err != nil {
		t.Fatalf("create e2e user: %v", err)
	}
	t.Cleanup(func() {
		_ = runSQL(cfg, fmt.Sprintf("DELETE FROM users WHERE id = %s;", sqlUUID(userID)))
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

func backendJWTSecret(cfg testConfig) (string, error) {
	command := exec.Command("docker", "compose", "exec", "-T", "backend", "printenv", "JWT_SECRET")
	command.Dir = cfg.repoRoot
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("read JWT_SECRET from running backend: %w", err)
	}

	secret := strings.TrimSpace(string(output))
	if len(secret) < 32 {
		return "", errors.New("running backend JWT_SECRET must be at least 32 characters")
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
	return "'" + value.String() + "'"
}

func sqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
