package tasks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PuppeteerAssigner struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewPuppeteerAssigner(baseURL, token string) *PuppeteerAssigner {
	return &PuppeteerAssigner{
		baseURL: strings.TrimRight(baseURL, "/"), token: token,
		client: &http.Client{Timeout: 2 * time.Second},
	}
}

func (assigner *PuppeteerAssigner) EnsureDailyTasks(ctx context.Context, userID uuid.UUID) error {
	endpoint := fmt.Sprintf("%s/internal/v1/users/%s/daily-tasks/ensure", assigner.baseURL, userID)
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
		if err != nil {
			return err
		}
		request.Header.Set("X-Service-Token", assigner.token)
		response, err := assigner.client.Do(request)
		if err != nil {
			lastErr = err
			continue
		}
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		if response.StatusCode == http.StatusNoContent {
			return nil
		}
		lastErr = fmt.Errorf("puppeteer returned %s", response.Status)
		if response.StatusCode < http.StatusInternalServerError {
			return lastErr
		}
	}

	return lastErr
}
