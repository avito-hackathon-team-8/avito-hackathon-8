package shop_test

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

const (
	fashionableBowlID = "fashionable-bowl"
	cyberBowlID       = "cyber-bowl"
	helperBowlID      = "helper-bowl"
	traderBedID       = "trader-bed"
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

type shopResponse struct {
	Items []shopItem `json:"items"`
}

type shopItem struct {
	ID            string `json:"id"`
	Category      string `json:"category"`
	Status        string `json:"status"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	ImageURL      string `json:"imageUrl"`
	RequiredLevel int    `json:"requiredLevel"`
	PriceLeaves   int64  `json:"priceLeaves"`
	DurationDays  int    `json:"durationDays"`
}

type rewardsResponse struct {
	Groups []rewardGroup `json:"groups"`
}

type rewardGroup struct {
	Category string       `json:"category"`
	Items    []rewardItem `json:"items"`
}

type rewardItem struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`
	Active    bool      `json:"active"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expiresAt"`
	ItemType  *string   `json:"itemType"`
}

var (
	cfgOnce sync.Once
	cfg     testConfig
	cfgErr  error
)

func TestShopRequiresAuthentication(t *testing.T) {
	cfg := getConfig(t)

	list := request(t, cfg, "", http.MethodGet, "/api/v1/shop", nil)
	assertError(t, list, http.StatusUnauthorized, "UNAUTHORIZED")

	purchase := request(t, cfg, "", http.MethodPost, "/api/v1/shop/"+fashionableBowlID+"/purchase", nil)
	assertError(t, purchase, http.StatusUnauthorized, "UNAUTHORIZED")
}

func TestShopCatalogReturnsPersonalStatuses(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg, 7, 1000)
	token := makeToken(t, cfg, userID)

	catalog := getShop(t, cfg, token)
	if len(catalog.Items) != 6 {
		t.Fatalf("items = %d, want 6", len(catalog.Items))
	}

	fashionable := findShopItem(t, catalog.Items, fashionableBowlID)
	if fashionable.ImageURL != "/api/v1/shop-images/bowl-fashionable.webp" {
		t.Fatalf("fashionable bowl imageUrl = %q", fashionable.ImageURL)
	}
	if fashionable.Category != "BOWL" || fashionable.Status != "AVAILABLE" ||
		fashionable.RequiredLevel != 5 || fashionable.PriceLeaves != 100 || fashionable.DurationDays != 3 {
		t.Fatalf("fashionable bowl = %+v", fashionable)
	}
	if fashionable.Title == "" || fashionable.Description == "" {
		t.Fatalf("fashionable bowl lacks display fields: %+v", fashionable)
	}

	if status := findShopItem(t, catalog.Items, cyberBowlID).Status; status != "AVAILABLE" {
		t.Fatalf("cyber bowl status = %q, want AVAILABLE", status)
	}
	if status := findShopItem(t, catalog.Items, helperBowlID).Status; status != "LOCKED" {
		t.Fatalf("helper bowl status = %q, want LOCKED", status)
	}
	if status := findShopItem(t, catalog.Items, "pro-bed").Status; status != "LOCKED" {
		t.Fatalf("pro bed status = %q, want LOCKED", status)
	}
	if category := findShopItem(t, catalog.Items, "pro-bed").Category; category != "BED" {
		t.Fatalf("pro bed category = %q, want BED", category)
	}
}

func TestShopPurchaseExtendsAndReplacesActiveItem(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg, 7, 1000)
	token := makeToken(t, cfg, userID)

	assertPurchase(t, request(t, cfg, token, http.MethodPost, "/api/v1/shop/"+fashionableBowlID+"/purchase", nil))
	assertPetLeaves(t, cfg, userID, 900)

	catalog := getShop(t, cfg, token)
	if status := findShopItem(t, catalog.Items, fashionableBowlID).Status; status != "ACTIVE" {
		t.Fatalf("fashionable bowl status = %q, want ACTIVE", status)
	}
	if status := findShopItem(t, catalog.Items, cyberBowlID).Status; status != "AVAILABLE" {
		t.Fatalf("cyber bowl status = %q, want AVAILABLE", status)
	}

	shopReward := findShopReward(t, getRewards(t, cfg, token), "BOWL", "FASHIONABLE_BOWL")
	if !shopReward.Active || shopReward.Status != "ACTIVE" {
		t.Fatalf("shop reward after purchase = %+v", shopReward)
	}
	firstExpiration := shopReward.ExpiresAt
	assertDurationNear(t, firstExpiration.Sub(time.Now().UTC()), 72*time.Hour)

	assertPurchase(t, request(t, cfg, token, http.MethodPost,
		"/api/v1/shop/"+fashionableBowlID+"/purchase", map[string]any{"confirmReplacement": false}))
	assertPetLeaves(t, cfg, userID, 800)
	extendedReward := findShopReward(t, getRewards(t, cfg, token), "BOWL", "FASHIONABLE_BOWL")
	assertDurationNear(t, extendedReward.ExpiresAt.Sub(firstExpiration), 72*time.Hour)

	replacementRequired := request(t, cfg, token, http.MethodPost, "/api/v1/shop/"+cyberBowlID+"/purchase", nil)
	assertError(t, replacementRequired, http.StatusConflict, "SHOP_REPLACEMENT_CONFIRMATION_REQUIRED")
	assertPetLeaves(t, cfg, userID, 800)

	assertPurchase(t, request(t, cfg, token, http.MethodPost,
		"/api/v1/shop/"+cyberBowlID+"/purchase", map[string]any{"confirmReplacement": true}))
	assertPetLeaves(t, cfg, userID, 650)

	catalog = getShop(t, cfg, token)
	if status := findShopItem(t, catalog.Items, fashionableBowlID).Status; status != "AVAILABLE" {
		t.Fatalf("replaced fashionable bowl status = %q, want AVAILABLE", status)
	}
	if status := findShopItem(t, catalog.Items, cyberBowlID).Status; status != "ACTIVE" {
		t.Fatalf("cyber bowl status = %q, want ACTIVE", status)
	}

	rewards := getRewards(t, cfg, token)
	if _, exists := lookupShopReward(rewards, "BOWL", "FASHIONABLE_BOWL"); exists {
		t.Fatalf("replaced fashionable bowl is still returned in %+v", rewards)
	}
	newReward := findShopReward(t, rewards, "BOWL", "CYBER_BOWL")
	if !newReward.Active || newReward.Status != "ACTIVE" {
		t.Fatalf("replacement reward = %+v, want active", newReward)
	}
}

func TestShopAllowsBowlAndBedAtTheSameTime(t *testing.T) {
	cfg := getConfig(t)
	userID := createUser(t, cfg, 7, 1000)
	token := makeToken(t, cfg, userID)

	assertPurchase(t, request(t, cfg, token, http.MethodPost, "/api/v1/shop/"+fashionableBowlID+"/purchase", nil))
	assertPurchase(t, request(t, cfg, token, http.MethodPost, "/api/v1/shop/"+traderBedID+"/purchase", nil))
	assertPetLeaves(t, cfg, userID, 600)

	catalog := getShop(t, cfg, token)
	if findShopItem(t, catalog.Items, fashionableBowlID).Status != "ACTIVE" ||
		findShopItem(t, catalog.Items, traderBedID).Status != "ACTIVE" {
		t.Fatalf("catalog after independent purchases = %+v", catalog.Items)
	}

	rewards := getRewards(t, cfg, token)
	bedReward := findShopReward(t, rewards, "BED", "TRADER_BED")
	if !bedReward.Active || bedReward.Status != "ACTIVE" {
		t.Fatalf("bed reward after purchase = %+v, want active", bedReward)
	}
}

func TestShopPurchaseErrorsDoNotSpendLeaves(t *testing.T) {
	cfg := getConfig(t)

	lockedUserID := createUser(t, cfg, 4, 1000)
	lockedToken := makeToken(t, cfg, lockedUserID)
	locked := request(t, cfg, lockedToken, http.MethodPost, "/api/v1/shop/"+fashionableBowlID+"/purchase", nil)
	assertError(t, locked, http.StatusConflict, "SHOP_LEVEL_REQUIRED")
	assertPetLeaves(t, cfg, lockedUserID, 1000)

	poorUserID := createUser(t, cfg, 5, 99)
	poorToken := makeToken(t, cfg, poorUserID)
	insufficient := request(t, cfg, poorToken, http.MethodPost, "/api/v1/shop/"+fashionableBowlID+"/purchase", nil)
	assertError(t, insufficient, http.StatusConflict, "INSUFFICIENT_LEAVES")
	assertPetLeaves(t, cfg, poorUserID, 99)

	missing := request(t, cfg, poorToken, http.MethodPost, "/api/v1/shop/missing/purchase", nil)
	assertError(t, missing, http.StatusNotFound, "SHOP_ITEM_NOT_FOUND")
	assertPetLeaves(t, cfg, poorUserID, 99)

	invalid := rawRequest(t, cfg, poorToken, http.MethodPost,
		"/api/v1/shop/"+fashionableBowlID+"/purchase", `{"unknown":true}`)
	assertError(t, invalid, http.StatusBadRequest, "INVALID_REQUEST")
	assertPetLeaves(t, cfg, poorUserID, 99)
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
		t.Skipf("shop API tests need running api-service and compose postgres: %v", cfgErr)
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
		apiURL:       envOr(dotenv, "SHOP_API_BASE_URL", "http://127.0.0.1:8090"),
		postgresDB:   envOr(dotenv, "POSTGRES_DB", "hackathon"),
		postgresUser: envOr(dotenv, "POSTGRES_USER", "hackathon"),
		jwtSecret:    envOr(dotenv, "JWT_SECRET", ""),
	}
	cfg.apiURL = strings.TrimRight(cfg.apiURL, "/")

	if err := waitForAPIService(cfg.apiURL); err != nil {
		return testConfig{}, err
	}
	if err := runSQL(cfg, "SELECT 1;"); err != nil {
		return testConfig{}, err
	}
	secret, err := apiServiceJWTSecret(cfg)
	if err != nil {
		return testConfig{}, err
	}
	cfg.jwtSecret = secret

	probe := requestWithoutTest(cfg, http.MethodGet, "/api/v1/shop")
	if probe.status != http.StatusUnauthorized {
		return testConfig{}, fmt.Errorf("running api-service does not expose current shop API: GET /api/v1/shop returned %d", probe.status)
	}

	return cfg, nil
}

func createUser(t *testing.T, cfg testConfig, level int, leaves int64) uuid.UUID {
	t.Helper()

	userID := uuid.New()
	petID := uuid.New()
	email := "shop-api-" + userID.String() + "@example.com"
	statement := fmt.Sprintf(`
INSERT INTO users (id, email, verified, created_at, updated_at)
VALUES (%s, %s, true, NOW(), NOW());
INSERT INTO pets (id, user_id, name, level, leaves, created_at, updated_at)
VALUES (%s, %s, '', %d, %d, NOW(), NOW());
`, sqlUUID(userID), sqlString(email), sqlUUID(petID), sqlUUID(userID), level, leaves)
	if err := runSQL(cfg, statement); err != nil {
		t.Fatalf("create test user: %v", err)
	}

	t.Cleanup(func() {
		cleanup := fmt.Sprintf(`
DELETE FROM leaf_transactions WHERE user_id = %s;
DELETE FROM rewards WHERE user_id = %s;
DELETE FROM leaderboard_entries WHERE user_id = %s;
DELETE FROM pets WHERE user_id = %s;
DELETE FROM users WHERE id = %s;
`, sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID), sqlUUID(userID))
		if err := runSQL(cfg, cleanup); err != nil {
			t.Errorf("cleanup test user: %v", err)
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

func getShop(t *testing.T, cfg testConfig, token string) shopResponse {
	t.Helper()
	result := request(t, cfg, token, http.MethodGet, "/api/v1/shop", nil)
	if result.status != http.StatusOK {
		t.Fatalf("get shop status = %d, want 200, body = %s", result.status, result.body)
	}
	var body shopResponse
	decode(t, result.body, &body)
	return body
}

func getRewards(t *testing.T, cfg testConfig, token string) rewardsResponse {
	t.Helper()
	result := request(t, cfg, token, http.MethodGet, "/api/app/rewards", nil)
	if result.status != http.StatusOK {
		t.Fatalf("get rewards status = %d, want 200, body = %s", result.status, result.body)
	}
	var body rewardsResponse
	decode(t, result.body, &body)
	return body
}

func assertPurchase(t *testing.T, result apiResult) {
	t.Helper()
	if result.status != http.StatusNoContent {
		t.Fatalf("purchase status = %d, want 204, body = %s", result.status, result.body)
	}
	if len(result.body) != 0 {
		t.Fatalf("purchase body = %q, want empty body", result.body)
	}
}

func findShopItem(t *testing.T, items []shopItem, itemID string) shopItem {
	t.Helper()
	for _, item := range items {
		if item.ID == itemID {
			return item
		}
	}
	t.Fatalf("shop item %q not found in %+v", itemID, items)
	return shopItem{}
}

func findShopReward(t *testing.T, rewards rewardsResponse, category, itemType string) rewardItem {
	t.Helper()
	item, exists := lookupShopReward(rewards, category, itemType)
	if exists {
		return item
	}
	t.Fatalf("shop reward category=%q itemType=%q not found in %+v", category, itemType, rewards)
	return rewardItem{}
}

func lookupShopReward(rewards rewardsResponse, category, itemType string) (rewardItem, bool) {
	for _, group := range rewards.Groups {
		if group.Category != category {
			continue
		}
		for _, item := range group.Items {
			if item.Source == "SHOP" && item.ItemType != nil && *item.ItemType == itemType {
				return item, true
			}
		}
	}
	return rewardItem{}, false
}

func assertDurationNear(t *testing.T, got, want time.Duration) {
	t.Helper()
	if got < want-time.Second || got > want+time.Second {
		t.Fatalf("duration = %s, want approximately %s", got, want)
	}
}

func assertPetLeaves(t *testing.T, cfg testConfig, userID uuid.UUID, want int64) {
	t.Helper()
	statement := fmt.Sprintf("SELECT leaves FROM pets WHERE user_id = %s;", sqlUUID(userID))
	output, err := querySQL(cfg, statement)
	if err != nil {
		t.Fatalf("read pet leaves: %v", err)
	}
	if output != fmt.Sprintf("%d", want) {
		t.Fatalf("pet leaves = %s, want %d", output, want)
	}
}

func assertError(t *testing.T, result apiResult, status int, code string) {
	t.Helper()
	if result.status != status || result.json["code"] != code {
		t.Fatalf("response = %d %s, want %d with code %s", result.status, result.body, status, code)
	}
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
	return doRequest(t, cfg, token, method, path, data)
}

func rawRequest(t *testing.T, cfg testConfig, token, method, path, body string) apiResult {
	t.Helper()
	return doRequest(t, cfg, token, method, path, []byte(body))
}

func doRequest(t *testing.T, cfg testConfig, token, method, path string, body []byte) apiResult {
	t.Helper()
	request, err := http.NewRequest(method, cfg.apiURL+path, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if len(body) > 0 {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("%s %s failed: %v", method, path, err)
	}
	defer func() { _ = response.Body.Close() }()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	result := apiResult{status: response.StatusCode, body: responseBody}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &result.json); err != nil {
			t.Fatalf("decode response JSON: %v, body = %s", err, responseBody)
		}
	}
	return result
}

func requestWithoutTest(cfg testConfig, method, path string) apiResult {
	request, err := http.NewRequest(method, cfg.apiURL+path, nil)
	if err != nil {
		return apiResult{}
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return apiResult{}
	}
	defer func() { _ = response.Body.Close() }()
	body, _ := io.ReadAll(response.Body)
	return apiResult{status: response.StatusCode, body: body}
}

func decode(t *testing.T, body []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("decode JSON: %v, body = %s", err, body)
	}
}

func waitForAPIService(apiURL string) error {
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
	return fmt.Errorf("api-service is not ready, run `make up` first: %w", lastErr)
}

func runSQL(cfg testConfig, statement string) error {
	_, err := executeSQL(cfg, statement, "-q")
	return err
}

func querySQL(cfg testConfig, statement string) (string, error) {
	output, err := executeSQL(cfg, statement, "-qAt")
	return strings.TrimSpace(string(output)), err
}

func executeSQL(cfg testConfig, statement, outputFlags string) ([]byte, error) {
	command := exec.Command("docker", "compose", "exec", "-T", "postgres", "psql", outputFlags, "-v", "ON_ERROR_STOP=1", "-U", cfg.postgresUser, "-d", cfg.postgresDB)
	command.Dir = cfg.repoRoot
	command.Stdin = strings.NewReader(statement)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return output, nil
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
	return sqlString(value.String()) + "::uuid"
}

func sqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
