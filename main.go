package main

import (
	"context"
	"log"

	"github.com/arthurblanchet59/korean-learning-go/internal/api"
	"github.com/arthurblanchet59/korean-learning-go/internal/config"
	"github.com/arthurblanchet59/korean-learning-go/internal/logging"
	sqliterepo "github.com/arthurblanchet59/korean-learning-go/internal/repository/sqlite"
	"github.com/arthurblanchet59/korean-learning-go/internal/service"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	appLogger, err := logging.New(cfg.LogDir)
	if err != nil {
		log.Fatalf("open application logs: %v", err)
	}
	defer appLogger.Close()

	store, err := sqliterepo.Open(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("open sqlite store: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		log.Fatalf("run sqlite migrations: %v", err)
	}

	if cfg.SeedDatabase {
		if err := store.SeedIfEmpty(); err != nil {
			log.Fatalf("seed sqlite database: %v", err)
		}
	}

	studyService := service.NewStudyService(store, store, store, store, store, core.NewScheduler())
	authService := service.NewAuthService(store, cfg.JWTSecret)
	adminService := service.NewAdminService(store)
	if err := authService.EnsureAdmin(
		ctx,
		cfg.AdminName,
		cfg.AdminEmail,
		cfg.AdminPassword,
	); err != nil {
		log.Fatalf("seed admin user: %v", err)
	}

	router := api.NewRouter(
		studyService,
		authService,
		adminService,
		appLogger.AccessMiddleware(),
		appLogger.RecoveryMiddleware(),
	)
	if cfg.WebRoot != "" {
		if err := api.ServeWebApp(router, cfg.WebRoot); err != nil {
			log.Fatalf("configure web application: %v", err)
		}
	}

	log.Printf("korean-learning API listening on http://localhost%s", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
