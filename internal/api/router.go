package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/internal/service"
)

type Handler struct {
	study *service.StudyService
}

type answerRequest struct {
	Rating string `json:"rating" binding:"required,oneof=again hard good easy"`
}

func NewRouter(study *service.StudyService) *gin.Engine {
	handler := &Handler{study: study}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), cors())

	router.GET("/health", handler.health)

	api := router.Group("/api")
	api.GET("/decks", handler.listDecks)
	api.GET("/cards", handler.listCards)
	api.GET("/reviews/due", handler.listDueCards)
	api.POST("/reviews/:id/answer", handler.answerCard)
	api.GET("/stats", handler.stats)

	return router
}

func (handler *Handler) health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (handler *Handler) listDecks(ctx *gin.Context) {
	decks, err := handler.study.ListDecks(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, decks)
}

func (handler *Handler) listCards(ctx *gin.Context) {
	cards, err := handler.study.ListCards(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, cards)
}

func (handler *Handler) listDueCards(ctx *gin.Context) {
	cards, err := handler.study.DueCards(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, cards)
}

func (handler *Handler) stats(ctx *gin.Context) {
	stats, err := handler.study.Stats(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

func (handler *Handler) answerCard(ctx *gin.Context) {
	var payload answerRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	review, err := handler.study.AnswerCard(ctx.Request.Context(), ctx.Param("id"), payload.Rating)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, repository.ErrNotFound) {
			status = http.StatusNotFound
		}

		writeError(ctx, status, err)
		return
	}

	ctx.JSON(http.StatusOK, review)
}

func cors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}

func writeError(ctx *gin.Context, status int, err error) {
	ctx.JSON(status, gin.H{"error": err.Error()})
}
