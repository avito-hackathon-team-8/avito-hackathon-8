package main

import (
	"log"
	"net/http"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/chest"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/config"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/daily_report"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/database"
	activityevents "github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/events"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/handlers"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/pet"
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

	authService := auth.NewService(db, cfg.Auth)
	dailyReportService := daily_report.NewService(db)
	rewardService := rewards.NewService(db, dailyReportService)
	petService := pet.NewService(db, dailyReportService)
	levelClaimsService := pet.NewLevelClaimsService(db, dailyReportService, rewardService)
	petService.SetLevelClaimsService(levelClaimsService)
	taskAssigner := tasks.NewPuppeteerAssigner(cfg.PuppeteerInternalURL, cfg.InternalServiceToken)
	taskService := tasks.NewService(db, dailyReportService, taskAssigner)
	chestService := chest.NewService(db, dailyReportService, petService, rewardService)
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
	}

	log.Printf("backend listening on %s", cfg.HTTPAddress)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
