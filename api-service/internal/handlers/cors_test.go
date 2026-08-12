package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSAllowsIdempotencyKey(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/pet/care", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}

	if allowed := response.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(allowed, "Idempotency-Key") {
		t.Fatalf("allowed headers = %q", allowed)
	}
}
