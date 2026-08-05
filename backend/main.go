package main

import (
	"log"
	"net/http"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/config"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/database"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/email"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/handlers"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/rewards"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/tasks"
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
	taskDefinitions, err := tasks.LoadDefaultDefinitions()
	if err != nil {
		log.Fatalf("load task definitions: %v", err)
	}
	taskService := tasks.NewService(db, taskDefinitions)
	router := handlers.NewRouter(authService, rewardService, taskService)

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
