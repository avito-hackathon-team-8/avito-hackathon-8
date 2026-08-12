package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type Job interface {
	Name() string
	Run(context.Context, *gorm.DB, time.Time) error
}

type DailyJob interface {
	Job
	Daily()
}

type Reporter interface {
	Record(string, time.Time, error)
}

type JobObserver interface{ StartJob(string) func() }

type Runner struct {
	db           *gorm.DB
	reporter     Reporter
	pollInterval time.Duration
	jobs         []Job
	now          func() time.Time
}

func NewRunner(db *gorm.DB, reporter Reporter, pollInterval time.Duration, scheduled ...Job) *Runner {
	return &Runner{db: db, reporter: reporter, pollInterval: pollInterval, jobs: scheduled, now: time.Now}
}

func (runner *Runner) Run(ctx context.Context) {
	now := runner.now().UTC()

	runner.runJobs(ctx, now, true)
	runner.runJobs(ctx, now, false)

	pollTicker := time.NewTicker(runner.pollInterval)
	defer pollTicker.Stop()

	midnightTimer := time.NewTimer(time.Until(nextMidnight(runner.now().UTC())))
	defer midnightTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			runner.runJobs(ctx, runner.now().UTC(), false)
		case <-midnightTimer.C:
			runner.runJobs(ctx, runner.now().UTC(), true)

			midnightTimer.Reset(time.Until(nextMidnight(runner.now().UTC())))
		}
	}
}

func (runner *Runner) runJobs(ctx context.Context, now time.Time, daily bool) {
	for _, job := range runner.jobs {
		_, isDaily := job.(DailyJob)

		if isDaily != daily {
			continue
		}

		finish := func() {}

		if observer, ok := runner.reporter.(JobObserver); ok {
			finish = observer.StartJob(job.Name())
		}

		err := runner.runLocked(ctx, job, now.UTC())

		finish()

		runner.reporter.Record(job.Name(), now.UTC(), err)

		if err != nil {
			log.Printf("service=daily-tasks-service job=%s error_type=run_failed: %v", job.Name(), err)
		}
	}
}

func (runner *Runner) runLocked(ctx context.Context, job Job, now time.Time) error {
	sqlDB, err := runner.db.DB()

	if err != nil {
		return fmt.Errorf("get database connection: %w", err)
	}

	connection, err := sqlDB.Conn(ctx)

	if err != nil {
		return fmt.Errorf("reserve job connection: %w", err)
	}

	defer func() {
		if err := connection.Close(); err != nil {
			log.Printf("job %s connection close failed: %v", job.Name(), err)
		}
	}()

	var acquired bool

	if err := connection.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(hashtext($1))", "daily-tasks-service:"+job.Name()).Scan(&acquired); err != nil {
		return fmt.Errorf("acquire job lease: %w", err)
	}

	if !acquired {
		return nil
	}

	defer func() {
		var released bool

		if err := connection.QueryRowContext(context.Background(), "SELECT pg_advisory_unlock(hashtext($1))", "daily-tasks-service:"+job.Name()).Scan(&released); err != nil {
			log.Printf("job %s lease release failed: %v", job.Name(), err)
		}
	}()

	if _, daily := job.(DailyJob); daily {
		result := runner.db.WithContext(ctx).Exec(`INSERT INTO job_runs (job_name,run_day,ran_at) VALUES (?,?,?) ON CONFLICT DO NOTHING`, job.Name(), startOfDay(now), now.UTC())

		if result.Error != nil {
			return fmt.Errorf("record job run: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return nil
		}
	}

	if err := job.Run(ctx, runner.db, now.UTC()); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}

func nextMidnight(at time.Time) time.Time {
	at = at.UTC()

	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
}
