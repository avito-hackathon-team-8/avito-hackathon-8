package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMetricsEndpointAndStableRouteLabels(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()

	if err != nil {
		t.Fatal(err)
	}

	metrics := New("test", sqlDB)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{userId}", func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) })

	metrics.Middleware(mux).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/users/unique-user-id", nil))

	response := httptest.NewRecorder()

	metrics.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := response.Body.String()

	if response.Code != http.StatusOK || !strings.Contains(body, `route="GET /users/{userId}"`) {
		t.Fatalf("metrics response status=%d body=%s", response.Code, body)
	}

	if strings.Contains(body, "unique-user-id") {
		t.Fatal("metrics contain a high-cardinality user id")
	}
}
