package jobs

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/puppeteer/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type blockingDailyJob struct {
	name    string
	started chan struct{}
	release chan struct{}
	runs    atomic.Int32
}

func (job *blockingDailyJob) Name() string { return job.name }
func (*blockingDailyJob) Daily()           {}

func (job *blockingDailyJob) Run(ctx context.Context, _ *gorm.DB, _ time.Time) error {
	job.runs.Add(1)
	close(job.started)

	select {
	case <-job.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestRunnerRunLockedUsesAdvisoryLock(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := db.AutoMigrate(&models.JobRun{}); err != nil {
		t.Fatalf("migrate job runs: %v", err)
	}

	job := &blockingDailyJob{
		name:    "test-advisory-lock-" + time.Now().UTC().Format("20060102150405.000000000"),
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	now := time.Now().UTC().Truncate(24 * time.Hour)
	runner := NewRunner(db, nil, time.Second, job)

	firstDone := make(chan error, 1)
	go func() {
		firstDone <- runner.runLocked(context.Background(), job, now)
	}()

	select {
	case <-job.started:
	case <-time.After(5 * time.Second):
		t.Fatal("first job did not start")
	}

	secondDone := make(chan error, 1)
	go func() {
		secondDone <- runner.runLocked(context.Background(), job, now)
	}()

	select {
	case err := <-secondDone:
		if err != nil {
			t.Fatalf("second run error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("second run did not return while lock was held")
	}

	close(job.release)

	if err := <-firstDone; err != nil {
		t.Fatalf("first run error = %v", err)
	}

	if got := job.runs.Load(); got != 1 {
		t.Fatalf("job runs = %d, want 1", got)
	}

	if err := db.Where("job_name = ?", job.name).Delete(&models.JobRun{}).Error; err != nil {
		t.Fatalf("cleanup job run: %v", err)
	}
}
