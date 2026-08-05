package weekly_login

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHTTPActivityCheckerStatus(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", request.Method)
		}

		if request.URL.Path != "/external/v1/login-activity/query" {
			t.Errorf("path = %s, want /external/v1/login-activity/query", request.URL.Path)
		}

		if authorization := request.Header.Get("Authorization"); authorization != "Bearer service-token" {
			t.Errorf("Authorization = %q, want Bearer service-token", authorization)
		}

		var body struct {
			UserID uuid.UUID `json:"userId"`
			Dates  []string  `json:"dates"`
		}

		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}

		if body.UserID != userID {
			t.Errorf("userId = %s, want %s", body.UserID, userID)
		}

		if len(body.Dates) != 1 || body.Dates[0] != "2026-08-05" {
			t.Errorf("dates = %v, want [2026-08-05]", body.Dates)
		}

		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"days":[{"date":"2026-08-05","status":"ACTIVE"}]}`))
	}))
	defer server.Close()

	checker := NewHTTPActivityChecker(server.URL, "service-token", time.Second)
	status, err := checker.Status(
		context.Background(),
		userID,
		time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	if status != ActivityStatusActive {
		t.Fatalf("Status() = %q, want ACTIVE", status)
	}
}

func TestHTTPActivityCheckerRejectsUnexpectedStatusCode(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	checker := NewHTTPActivityChecker(server.URL, "", time.Second)
	status, err := checker.Status(context.Background(), uuid.New(), time.Now())

	if err == nil {
		t.Fatal("Status() error = nil, want an error")
	}

	if status != ActivityStatusUnknown {
		t.Fatalf("Status() = %q, want UNKNOWN", status)
	}
}
