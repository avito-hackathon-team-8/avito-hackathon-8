package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/puppeteer/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

		err := runner.runLocked(ctx, job, now.UTC())

		runner.reporter.Record(job.Name(), now.UTC(), err)

		if err != nil {
			log.Printf("job %s failed: %v", job.Name(), err)
		}
	}
}

func (runner *Runner) runLocked(ctx context.Context, job Job, now time.Time) error {
	return runner.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var acquired bool

		if err := tx.Raw("SELECT pg_try_advisory_xact_lock(hashtext(?))", "puppeteer:"+job.Name()).Scan(&acquired).Error; err != nil {
			return fmt.Errorf("acquire job lease: %w", err)
		}

		if !acquired {
			return nil
		}

		if _, daily := job.(DailyJob); daily {
			run := models.JobRun{JobName: job.Name(), RunDay: startOfDay(now), RanAt: now.UTC()}

			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&run)

			if result.Error != nil {
				return fmt.Errorf("record job run: %w", result.Error)
			}

			if result.RowsAffected == 0 {
				return nil
			}
		}

		if err := job.Run(ctx, tx, now.UTC()); err != nil {
			return fmt.Errorf("run: %w", err)
		}

		return nil
	})
}

func nextMidnight(at time.Time) time.Time {
	at = at.UTC()

	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
}
