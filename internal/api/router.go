package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/arthurblanchet59/korean-learning-go/internal/service"
)

type Handler struct {
	study  *service.StudyService
	auth   *service.AuthService
	admin  *service.AdminService
	backup *service.ClientBackupService
}

type clientBackupRequest struct {
	Config json.RawMessage `json:"config" binding:"required"`
	State  json.RawMessage `json:"state" binding:"required"`
}

type answerRequest struct {
	Rating string `json:"rating" binding:"required,oneof=again hard good easy"`
}

type answerCheckRequest struct {
	Answer    string `json:"answer" binding:"required,max=500"`
	Direction string `json:"direction" binding:"omitempty,oneof=korean-to-french french-to-korean"`
}

type importCardsRequest struct {
	DeckID string `json:"deckId" binding:"required"`
	CSV    string `json:"csv" binding:"required"`
}

type lessonProgressRequest struct {
	Completed bool `json:"completed"`
	Score     int  `json:"score" binding:"min=0,max=100"`
}

type journalRequest struct {
	Title string `json:"title" binding:"max=120"`
	Text  string `json:"text" binding:"required,min=1,max=10000"`
}

type deckRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=120"`
	Description string `json:"description" binding:"max=500"`
}

type deckPatchRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=120"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

type cardRequest struct {
	DeckID             string   `json:"deckId" binding:"required"`
	Kind               string   `json:"kind" binding:"required,oneof=vocabulary phrase hangul"`
	Korean             string   `json:"korean" binding:"required,min=1,max=200"`
	Translation        string   `json:"translation" binding:"required,min=1,max=200"`
	Romanization       string   `json:"romanization" binding:"max=200"`
	ExampleKorean      string   `json:"exampleKorean" binding:"max=500"`
	ExampleTranslation string   `json:"exampleTranslation" binding:"max=500"`
	Tags               []string `json:"tags"`
}

type cardPatchRequest struct {
	DeckID             *string   `json:"deckId"`
	Kind               *string   `json:"kind" binding:"omitempty,oneof=vocabulary phrase hangul"`
	Korean             *string   `json:"korean" binding:"omitempty,min=1,max=200"`
	Translation        *string   `json:"translation" binding:"omitempty,min=1,max=200"`
	Romanization       *string   `json:"romanization" binding:"omitempty,max=200"`
	ExampleKorean      *string   `json:"exampleKorean" binding:"omitempty,max=500"`
	ExampleTranslation *string   `json:"exampleTranslation" binding:"omitempty,max=500"`
	Tags               *[]string `json:"tags"`
}

type bulkDeckUpdateRequest struct {
	IDs   []string         `json:"ids" binding:"required,min=1"`
	Patch deckPatchRequest `json:"patch" binding:"required"`
}

type bulkCardUpdateRequest struct {
	IDs   []string         `json:"ids" binding:"required,min=1"`
	Patch cardPatchRequest `json:"patch" binding:"required"`
}

type bulkDeleteRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}

type registerRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=80"`
	Email    string `json:"email" binding:"required,max=254"`
	Password string `json:"password" binding:"required,min=8,max=120"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,max=254"`
	Password string `json:"password" binding:"required"`
}

type updateUserRequest struct {
	Name     string `json:"name" binding:"omitempty,min=2,max=80"`
	Email    string `json:"email" binding:"omitempty,max=254"`
	Password string `json:"password" binding:"omitempty,min=8,max=120"`
}

func NewRouter(study *service.StudyService, auth *service.AuthService, adminService *service.AdminService, backupService *service.ClientBackupService, middlewares ...gin.HandlerFunc) *gin.Engine {
	handler := &Handler{study: study, auth: auth, admin: adminService, backup: backupService}

	router := gin.New()
	if len(middlewares) > 0 {
		router.Use(middlewares...)
	} else {
		router.Use(gin.Logger(), gin.Recovery())
	}
	router.Use(cors())

	router.GET("/health", handler.health)
	registerSwaggerRoutes(router)
	router.POST("/user/register", handler.register)
	router.POST("/user/login", handler.login)
	router.GET("/search", handler.authMiddleware(), handler.searchAll)

	user := router.Group("/user")
	user.Use(handler.authMiddleware())
	user.GET("/me", handler.me)
	user.PUT("/me", handler.updateMe)

	admin := router.Group("/admin")
	admin.Use(handler.authMiddleware(), requireAdmin())
	admin.GET("/users", handler.adminListUsers)
	admin.PUT("/users/:id", handler.adminUpdateUser)
	admin.POST("/reset", handler.resetDatabase)
	admin.DELETE("/reset", handler.resetDatabase)
	admin.POST("/rag/reindex", handler.reindexKnowledge)

	api := router.Group("/api")
	api.Use(handler.authMiddleware())
	api.POST("/reset", requireAdmin(), handler.resetDatabase)
	api.DELETE("/reset", requireAdmin(), handler.resetDatabase)
	api.GET("/search", handler.searchAll)
	api.GET("/client-backup", handler.getClientBackup)
	api.PUT("/client-backup", handler.saveClientBackup)
	api.GET("/decks", handler.listDecks)
	api.GET("/decks/search", handler.searchDecks)
	api.POST("/decks", handler.createDeck)
	api.PUT("/decks/bulk", handler.updateDecks)
	api.DELETE("/decks/bulk", handler.deleteDecks)
	api.GET("/decks/:id", handler.getDeck)
	api.PUT("/decks/:id", handler.updateDeck)
	api.DELETE("/decks/:id", handler.deleteDeck)
	api.GET("/cards", handler.listCards)
	api.GET("/cards/search", handler.searchCards)
	api.POST("/cards", handler.createCard)
	api.GET("/cards/export", handler.exportCards)
	api.POST("/cards/import", handler.importCards)
	api.PUT("/cards/bulk", handler.updateCards)
	api.DELETE("/cards/bulk", handler.deleteCards)
	api.GET("/cards/difficult", handler.listDifficultCards)
	api.GET("/cards/:id", handler.getCard)
	api.PUT("/cards/:id", handler.updateCard)
	api.DELETE("/cards/:id", handler.deleteCard)
	api.GET("/reviews/due", handler.listDueCards)
	api.POST("/reviews/:id/answer", handler.answerCard)
	api.GET("/stats", handler.stats)
	api.GET("/lessons", handler.listLessons)
	api.GET("/lessons/:id", handler.getLesson)
	api.PUT("/lessons/:id/progress", handler.updateLessonProgress)
	api.GET("/journal", handler.listJournalEntries)
	api.GET("/rag/status", handler.knowledgeIndexStatus)
	api.POST("/journal", handler.createJournalEntry)
	api.POST("/journal/correct", handler.correctJournalText)
	api.GET("/journal/:id", handler.getJournalEntry)
	api.PUT("/journal/:id", handler.updateJournalEntry)
	api.DELETE("/journal/:id", handler.deleteJournalEntry)

	studyRoutes := router.Group("/study")
	studyRoutes.Use(handler.authMiddleware())
	studyRoutes.GET("/today", handler.listDueCards)
	studyRoutes.POST("/cards/:id/answer", handler.answerCard)
	studyRoutes.POST("/cards/:id/check", handler.checkAnswer)

	return router
}

func (handler *Handler) health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
