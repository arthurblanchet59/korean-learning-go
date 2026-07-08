package main

import (
	"context"
	"log"
	"net/http"

	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/config"
	httpapi "github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/http"
	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/repository/memory"
	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/repository/postgres"
	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/service"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	store, closeStore := buildStore(ctx, cfg)
	defer closeStore()

	studyService := service.NewStudyService(store, store, store, core.NewScheduler())
	handler := httpapi.NewHandler(studyService)

	log.Printf("korean-learning backend listening on http://localhost%s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler.Routes()); err != nil {
		log.Fatal(err)
	}
}

type studyStore interface {
	repository.DeckRepository
	repository.CardRepository
	repository.ReviewRepository
}

func buildStore(ctx context.Context, cfg config.Config) (studyStore, func()) {
	if cfg.DatabaseURL == "" {
		log.Println("DATABASE_URL is empty; using in-memory repository")
		store := memory.NewSeededStore()
		return store, func() {}
	}

	store, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open postgres repository: %v", err)
	}

	if cfg.AutoMigrate {
		if err := store.Migrate(ctx); err != nil {
			_ = store.Close()
			log.Fatalf("run postgres migrations: %v", err)
		}
	}

	if cfg.SeedDatabase {
		if err := store.SeedIfEmpty(ctx); err != nil {
			_ = store.Close()
			log.Fatalf("seed postgres database: %v", err)
		}
	}

	return store, func() {
		if err := store.Close(); err != nil {
			log.Printf("close postgres repository: %v", err)
		}
	}
}
