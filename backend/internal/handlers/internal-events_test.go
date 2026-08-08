package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInternalEventsRequiresServiceToken(t *testing.T) {
	handler := &internalEventHandler{token: strings.Repeat("a", 32)}
	request := httptest.NewRequest(http.MethodPost, "/api/internal/v1/users/not-a-uuid/events", strings.NewReader(`{"events":[]}`))
	response := httptest.NewRecorder()
	handler.record(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestInternalEventsAcceptsValidServiceTokenBeforeValidation(t *testing.T) {
	token := strings.Repeat("a", 32)
	handler := &internalEventHandler{token: token}
	request := httptest.NewRequest(http.MethodPost, "/api/internal/v1/users/not-a-uuid/events", strings.NewReader(`{"events":[]}`))
	request.Header.Set("X-Service-Token", token)
	request.SetPathValue("userId", "not-a-uuid")
	response := httptest.NewRecorder()
	handler.record(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}
