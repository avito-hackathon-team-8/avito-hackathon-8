package main

import (
	"log"
	"net/http"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/chest"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/config"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/database"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/email"
	activityevents "github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/events"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/handlers"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/leaves"
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

	mailer, err := email.NewSender(cfg.Email)

	if err != nil {
		log.Fatalf("configure email: %v", err)
	}

	authService := auth.NewService(db, mailer, cfg.Auth)
	rewardService := rewards.NewService(db)
	levelClaimsService := pet.NewLevelClaimsService(db, rewardService)
	taskAssigner := tasks.NewPuppeteerAssigner(cfg.PuppeteerInternalURL, cfg.InternalServiceToken)
	taskService := tasks.NewService(db, taskAssigner)
	leafService := leaves.NewService(db)
	petService := pet.NewService(db)
	petService.SetLevelClaimsService(levelClaimsService)
	chestService := chest.NewService(db, petService, rewardService)
	activityService := weekly_login.NewLoginService(db)
	weeklyLoginService := weekly_login.NewService(db, activityService, leafService)
	eventService := activityevents.NewService(db, taskService)
	router := handlers.NewRouter(db, authService, rewardService, taskService, leafService, petService, weeklyLoginService, eventService, cfg.InternalServiceToken, activityService, chestService)

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
