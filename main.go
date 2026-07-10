package main

import (
	"context"
	"log"

	"github.com/arthurblanchet59/korean-learning-go/internal/api"
	"github.com/arthurblanchet59/korean-learning-go/internal/config"
	sqliterepo "github.com/arthurblanchet59/korean-learning-go/internal/repository/sqlite"
	"github.com/arthurblanchet59/korean-learning-go/internal/service"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

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

	studyService := service.NewStudyService(store, store, store, core.NewScheduler())
	authService := service.NewAuthService(store, cfg.JWTSecret)
	if err := authService.EnsureAdmin(
		ctx,
		cfg.AdminName,
		cfg.AdminEmail,
		cfg.AdminPassword,
	); err != nil {
		log.Fatalf("seed admin user: %v", err)
	}

	router := api.NewRouter(studyService, authService)

	log.Printf("korean-learning API listening on http://localhost%s", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
