package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/chest"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/config"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/daily_report"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/database"
	activityevents "github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/events"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/handlers"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/reward_catalog"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/rewards"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/tasks"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/weekly_login"
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

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get database connection: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("close database connection: %v", err)
		}
	}()

	authService := auth.NewService(db, cfg.Auth)
	dailyReportService := daily_report.NewService(db)
	rewardService := rewards.NewService(db, dailyReportService)
	rewardCatalog, err := reward_catalog.Load(cfg.LevelRewardsConfig)
	if err != nil {
		log.Printf("load reward catalog: %v", err)
		return
	}
	petService := pet.NewService(db, dailyReportService)
	levelClaimsService := pet.NewLevelClaimsService(db, dailyReportService, rewardService, rewardCatalog.LevelRewards())
	petService.SetLevelClaimsService(levelClaimsService)
	taskAssigner := tasks.NewPuppeteerAssigner(cfg.PuppeteerInternalURL, cfg.InternalServiceToken)
	taskService := tasks.NewService(db, dailyReportService, taskAssigner)
	chestService := chest.NewService(db, dailyReportService, petService, rewardService, rewardCatalog.ChestRewards())
	weeklyLoginService := weekly_login.NewService(db, dailyReportService, petService)
	eventService := activityevents.NewService(db, dailyReportService, taskService)
	router := handlers.NewRouter(db, authService, rewardService,
		taskService, petService, levelClaimsService,
		weeklyLoginService, eventService, dailyReportService,
		cfg.InternalServiceToken, chestService)

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	log.Printf("backend listening on %s", cfg.HTTPAddress)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("serve backend: %v", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shut down backend: %v", err)
		}
	}
}
