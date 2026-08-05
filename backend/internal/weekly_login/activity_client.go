package weekly_login

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type HTTPActivityChecker struct {
	endpoint string
	token    string
	client   *http.Client
}

func NewHTTPActivityChecker(baseURL, token string, timeout time.Duration) *HTTPActivityChecker {
	return &HTTPActivityChecker{
		endpoint: strings.TrimRight(baseURL, "/") + "/external/v1/login-activity/query",
		token:    token,
		client:   &http.Client{Timeout: timeout},
	}
}

func (checker *HTTPActivityChecker) Status(
	ctx context.Context,
	userID uuid.UUID,
	date time.Time,
) (ActivityStatus, error) {
	dateString := utcDate(date).Format(time.DateOnly)
	body, err := json.Marshal(struct {
		UserID uuid.UUID `json:"userId"`
		Dates  []string  `json:"dates"`
	}{
		UserID: userID,
		Dates:  []string{dateString},
	})

	if err != nil {
		return ActivityStatusUnknown, fmt.Errorf("encode activity request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, checker.endpoint, bytes.NewReader(body))

	if err != nil {
		return ActivityStatusUnknown, fmt.Errorf("create activity request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	if checker.token != "" {
		request.Header.Set("Authorization", "Bearer "+checker.token)
	}

	response, err := checker.client.Do(request)

	if err != nil {
		return ActivityStatusUnknown, fmt.Errorf("request activity: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return ActivityStatusUnknown, fmt.Errorf("request activity: unexpected status %d", response.StatusCode)
	}

	var result struct {
		Days []struct {
			Date   string         `json:"date"`
			Status ActivityStatus `json:"status"`
		} `json:"days"`
	}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return ActivityStatusUnknown, fmt.Errorf("decode activity response: %w", err)
	}

	for _, day := range result.Days {
		if day.Date == dateString {
			switch day.Status {
			case ActivityStatusActive,
				ActivityStatusInactive,
				ActivityStatusUnknown,
				ActivityStatusFuture,
				ActivityStatusBeforeRegistration:
				return day.Status, nil
			default:
				return ActivityStatusUnknown, fmt.Errorf("activity response contains invalid status %q", day.Status)
			}
		}
	}

	return ActivityStatusUnknown, errors.New("activity response does not contain requested date")
}
