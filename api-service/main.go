package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/auth"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/chest"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/config"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/daily_report"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/database"
	activityevents "github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/events"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/handlers"
	appmetrics "github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/metrics"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/pet"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/petstate"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/reward_catalog"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/rewards"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/shop"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/tasks"
	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/weekly_login"
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

	sqlDB, err := database.Configure(context.Background(), db, cfg.DatabasePool)

	if err != nil {
		log.Fatalf("get database connection: %v", err)
	}

	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("close database connection: %v", err)
		}
	}()

	metrics := appmetrics.New("api-service", sqlDB)

	if err := metrics.InstrumentGORM(db); err != nil {
		log.Printf("service=api-service error_type=metrics message=%q", err)

		return
	}

	authService := auth.NewService(db, cfg.Auth)
	dailyReportService := daily_report.NewService(db)
	rewardService := rewards.NewService(db, dailyReportService)
	shopCatalog, err := shop.LoadCatalog(cfg.ShopItemsConfig)

	if err != nil {
		log.Printf("load shop catalog: %v", err)

		return
	}

	rewardCatalog, err := reward_catalog.Load(cfg.LevelRewardsConfig)

	if err != nil {
		log.Printf("load reward catalog: %v", err)

		return
	}

	petService := pet.NewService(db, dailyReportService)
	shopService := shop.NewService(db, dailyReportService, petService, rewardService, shopCatalog)
	petStateService, err := petstate.NewService(cfg.PetStateInternalURL, cfg.InternalServiceToken, cfg.KafkaBrokers, cfg.PetStateKafkaTopic, cfg.APIServiceInstanceID, metrics)

	if err != nil {
		log.Printf("configure pet state service: %v", err)

		return
	}

	defer petStateService.Close()

	levelClaimsService := pet.NewLevelClaimsService(db, dailyReportService, rewardService, rewardCatalog.LevelRewards())
	petService.SetLevelClaimsService(levelClaimsService)
	taskAssigner := tasks.NewDailyTasksAssigner(cfg.DailyTasksInternalURL, cfg.InternalServiceToken, metrics)
	taskService := tasks.NewService(db, dailyReportService, taskAssigner)
	chestService := chest.NewService(db, dailyReportService, petService, rewardService, rewardCatalog.ChestRewards())
	weeklyLoginService := weekly_login.NewService(db, dailyReportService, petService, petStateService)
	weeklyLoginService.SetMetrics(metrics)
	eventService := activityevents.NewService(db, dailyReportService, taskService)
	router := handlers.NewRouter(db, authService, rewardService,
		taskService, petService, levelClaimsService,
		weeklyLoginService, eventService, dailyReportService,
		cfg.InternalServiceToken, chestService, shopService, petStateService, metrics,
		cfg.ShopImagesDir,
		handlers.HTTPReadinessChecker{URL: cfg.DailyTasksInternalURL, Client: &http.Client{Timeout: 750 * time.Millisecond}})

	root := http.NewServeMux()

	root.Handle("GET /metrics", metrics.Handler())
	root.Handle("/", metrics.Middleware(router))

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           root,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	log.Printf("api-service listening on %s", cfg.HTTPAddress)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go petStateService.Run(ctx)

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("serve api-service: %v", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shut down api-service: %v", err)
		}
	}
}
