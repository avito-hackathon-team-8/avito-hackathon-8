package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/daily-tasks-service/internal/config"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/daily-tasks-service/internal/database"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/daily-tasks-service/internal/health"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/daily-tasks-service/internal/jobs"
	appmetrics "github.com/avito-hackathon-team-8/avito-hackathon-8/daily-tasks-service/internal/metrics"
)

type reporter struct {
	status  *health.Status
	metrics *appmetrics.Metrics
}

func (reporter reporter) Record(name string, at time.Time, err error) {
	reporter.status.Record(name, at, err)
	reporter.metrics.Record(name, at, err)
}
func (reporter reporter) StartJob(name string) func() { return reporter.metrics.StartJob(name) }

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(cfg.DatabaseURL)

	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	sqlDB, err := database.Configure(context.Background(), db, cfg.DatabasePool)

	if err != nil {
		log.Fatalf("get database connection: %v", err)
	}

	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("close database connection: %v", err)
		}
	}()

	taskDefinitions, err := jobs.LoadTaskDefinitions(cfg.TaskDefinitionsConfig)

	if err != nil {
		log.Printf("load task definitions: %v", err)
		return
	}

	leaderboardRewards, err := jobs.LoadLeaderboardRewards(cfg.LeaderboardRewardsConfig)

	if err != nil {
		log.Printf("load leaderboard rewards: %v", err)

		return
	}

	metrics := appmetrics.New(sqlDB)

	if err := metrics.InstrumentGORM(db); err != nil {
		log.Printf("instrument database metrics: %v", err)

		return
	}
	taskAssigner := jobs.NewTaskAssigner(db, taskDefinitions, nil, metrics)

	if err := taskAssigner.Seed(context.Background()); err != nil {
		log.Printf("seed task definitions: %v", err)

		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	status := health.NewStatus()
	jobReporter := reporter{status: status, metrics: metrics}

	runner := jobs.NewRunner(db, jobReporter, cfg.PollInterval,
		jobs.CalculateLeaderboard{Rewards: leaderboardRewards, Observer: metrics},
		jobs.DistributeDailyTasks{Assigner: taskAssigner},
	)

	root := http.NewServeMux()

	root.Handle("GET /metrics", metrics.Handler())
	root.Handle("/", health.NewHandler(status, db, cfg.InternalServiceToken, taskAssigner))

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           metrics.Middleware(root),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	runnerDone := make(chan struct{})

	go func() {
		defer close(runnerDone)
		runner.Run(ctx)
	}()

	go func() {
		log.Printf("daily-tasks-service health endpoint listening on %s", cfg.HTTPAddress)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("health server failed: %v", err)

			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shut down health server: %v", err)
	}

	select {
	case <-runnerDone:
	case <-time.After(5 * time.Second):
		log.Printf("daily-tasks-service runner did not stop before shutdown deadline")
	}
}
