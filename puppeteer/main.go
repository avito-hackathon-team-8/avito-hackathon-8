package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/puppeteer/internal/config"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/puppeteer/internal/database"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/puppeteer/internal/health"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/puppeteer/internal/jobs"
)

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(cfg.DatabaseURL)

	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	taskDefinitions, err := jobs.LoadTaskDefinitions(cfg.TaskDefinitionsConfig)
	if err != nil {
		log.Fatalf("load task definitions: %v", err)
	}
	leaderboardRewards, err := jobs.LoadLeaderboardRewards(cfg.LeaderboardRewardsConfig)
	if err != nil {
		log.Fatalf("load leaderboard rewards: %v", err)
	}
	taskAssigner := jobs.NewTaskAssigner(db, taskDefinitions, nil)
	if err := taskAssigner.Seed(context.Background()); err != nil {
		log.Fatalf("seed task definitions: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	status := health.NewStatus()

	runner := jobs.NewRunner(db, status, cfg.PollInterval,
		jobs.CalculateLeaderboard{Rewards: leaderboardRewards},
		jobs.DistributeDailyTasks{Assigner: taskAssigner},
	)

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           health.NewHandler(status, db, cfg.InternalServiceToken, taskAssigner),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go runner.Run(ctx)

	go func() {
		log.Printf("puppeteer health endpoint listening on %s", cfg.HTTPAddress)

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
}
