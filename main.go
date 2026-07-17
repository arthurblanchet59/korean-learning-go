package main

import (
	"context"
	"log"

	"github.com/arthurblanchet59/korean-learning-go/internal/api"
	"github.com/arthurblanchet59/korean-learning-go/internal/config"
	"github.com/arthurblanchet59/korean-learning-go/internal/foundry"
	"github.com/arthurblanchet59/korean-learning-go/internal/logging"
	sqliterepo "github.com/arthurblanchet59/korean-learning-go/internal/repository/sqlite"
	"github.com/arthurblanchet59/korean-learning-go/internal/service"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
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

	var corrector service.KoreanCorrector = service.LocalKoreanCorrector{}
	if cfg.AzureAIEnabled() {
		generator, err := foundry.NewCorrector(cfg.AzureAIEndpoint, cfg.AzureAIAPIKey, cfg.AzureAIModel)
		if err != nil {
			log.Fatalf("configure Microsoft Foundry correction: %v", err)
		}
		corrector = generator
		if cfg.RAGEnabled() {
			embedder, err := foundry.NewEmbedder(
				cfg.EmbeddingEndpoint(),
				cfg.EmbeddingAPIKey(),
				cfg.AzureAIEmbeddingModel,
				cfg.AzureAIEmbeddingDimensions,
			)
			if err != nil {
				log.Fatalf("configure Microsoft Foundry embeddings: %v", err)
			}
			corrector = service.NewRAGCorrector(store, embedder, generator)
			log.Printf("pedagogical RAG enabled with embedding deployment %q", cfg.AzureAIEmbeddingModel)
		}
		log.Printf("journal correction enabled with Microsoft Foundry deployment %q", cfg.AzureAIModel)
	} else {
		log.Print("journal correction uses local rules; configure AZURE_AI_ENDPOINT, AZURE_AI_API_KEY and AZURE_AI_MODEL to enable Microsoft Foundry")
	}
	studyService := service.NewStudyService(store, store, store, store, store, core.NewScheduler(), corrector)
	authService := service.NewAuthService(store, cfg.JWTSecret)
	adminService := service.NewAdminService(store)
	backupService := service.NewClientBackupService(store)
	if err := authService.EnsureAdmin(
		ctx,
		cfg.AdminName,
		cfg.AdminEmail,
		cfg.AdminPassword,
	); err != nil {
		log.Fatalf("seed admin user: %v", err)
	}
	if cfg.SeedDatabase {
		if err := store.SeedAllUsers(ctx); err != nil {
			log.Fatalf("seed study curriculum: %v", err)
		}
	}
	if cfg.RAGEnabled() {
		go func() {
			status, err := studyService.EnsureKnowledgeIndex(context.Background())
			if err != nil {
				log.Printf("pedagogical index initialization failed: %v", err)
				return
			}
			log.Printf("pedagogical index ready with %d chunks", status.ChunkCount)
		}()
	}

	router := api.NewRouter(
		studyService,
		authService,
		adminService,
		backupService,
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
