package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/internal/service"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

type Handler struct {
	study *service.StudyService
	auth  *service.AuthService
	admin *service.AdminService
}

type answerRequest struct {
	Rating string `json:"rating" binding:"required,oneof=again hard good easy"`
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
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=120"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type updateUserRequest struct {
	Name     string `json:"name" binding:"omitempty,min=2,max=80"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=8,max=120"`
}

func NewRouter(study *service.StudyService, auth *service.AuthService, adminService *service.AdminService, middlewares ...gin.HandlerFunc) *gin.Engine {
	handler := &Handler{study: study, auth: auth, admin: adminService}

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
	admin.PUT("/users/:id", handler.adminUpdateUser)
	admin.POST("/reset", handler.resetDatabase)
	admin.DELETE("/reset", handler.resetDatabase)

	api := router.Group("/api")
	api.Use(handler.authMiddleware())
	api.POST("/reset", requireAdmin(), handler.resetDatabase)
	api.DELETE("/reset", requireAdmin(), handler.resetDatabase)
	api.GET("/search", handler.searchAll)
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
	api.PUT("/cards/bulk", handler.updateCards)
	api.DELETE("/cards/bulk", handler.deleteCards)
	api.GET("/cards/difficult", handler.listDifficultCards)
	api.GET("/cards/:id", handler.getCard)
	api.PUT("/cards/:id", handler.updateCard)
	api.DELETE("/cards/:id", handler.deleteCard)
	api.GET("/reviews/due", handler.listDueCards)
	api.POST("/reviews/:id/answer", handler.answerCard)
	api.GET("/stats", handler.stats)

	studyRoutes := router.Group("/study")
	studyRoutes.Use(handler.authMiddleware())
	studyRoutes.GET("/today", handler.listDueCards)
	studyRoutes.POST("/cards/:id/answer", handler.answerCard)

	return router
}

func (handler *Handler) health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (handler *Handler) register(ctx *gin.Context) {
	var payload registerRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := handler.auth.Register(ctx.Request.Context(), service.RegisterInput{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, repository.ErrConflict) {
			status = http.StatusConflict
		}
		writeError(ctx, status, err)
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

func (handler *Handler) login(ctx *gin.Context) {
	var payload loginRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := handler.auth.Login(ctx.Request.Context(), service.LoginInput{
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		writeError(ctx, status, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) me(ctx *gin.Context) {
	user, err := handler.auth.UserByID(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, repository.ErrNotFound) {
			status = http.StatusNotFound
		}
		writeError(ctx, status, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (handler *Handler) updateMe(ctx *gin.Context) {
	var payload updateUserRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := handler.auth.UpdateSelf(ctx.Request.Context(), currentUserID(ctx), service.UpdateUserInput{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		writeAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (handler *Handler) adminUpdateUser(ctx *gin.Context) {
	var payload updateUserRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := handler.auth.AdminUpdateUser(ctx.Request.Context(), ctx.Param("id"), service.UpdateUserInput{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		writeAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (handler *Handler) resetDatabase(ctx *gin.Context) {
	result, err := handler.admin.ResetDatabase(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "database reset completed",
		"result":  result,
	})
}

func (handler *Handler) listDecks(ctx *gin.Context) {
	decks, err := handler.study.ListDecks(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, decks)
}

func (handler *Handler) searchDecks(ctx *gin.Context) {
	results, err := handler.study.SearchDecks(ctx.Request.Context(), ctx.Query("query"))
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (handler *Handler) searchAll(ctx *gin.Context) {
	results, err := handler.study.SearchAll(ctx.Request.Context(), ctx.Query("query"))
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (handler *Handler) getDeck(ctx *gin.Context) {
	deck, err := handler.study.DeckByID(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, deck)
}

func (handler *Handler) createDeck(ctx *gin.Context) {
	var payload deckRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	deck, err := handler.study.CreateDeck(ctx.Request.Context(), service.DeckInput{
		Name:        payload.Name,
		Description: payload.Description,
	})
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, deck)
}

func (handler *Handler) updateDeck(ctx *gin.Context) {
	var payload deckPatchRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	deck, err := handler.study.UpdateDeck(ctx.Request.Context(), ctx.Param("id"), deckPatchInput(payload))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, deck)
}

func (handler *Handler) updateDecks(ctx *gin.Context) {
	var payload bulkDeckUpdateRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	decks, err := handler.study.UpdateDecks(ctx.Request.Context(), payload.IDs, deckPatchInput(payload.Patch))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, decks)
}

func (handler *Handler) deleteDeck(ctx *gin.Context) {
	if err := handler.study.DeleteDeck(ctx.Request.Context(), ctx.Param("id")); err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (handler *Handler) deleteDecks(ctx *gin.Context) {
	var payload bulkDeleteRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	deleted, err := handler.study.DeleteDecks(ctx.Request.Context(), payload.IDs)
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": deleted})
}

func (handler *Handler) listCards(ctx *gin.Context) {
	cards, err := handler.study.ListCards(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, cards)
}

func (handler *Handler) searchCards(ctx *gin.Context) {
	results, err := handler.study.SearchCards(ctx.Request.Context(), ctx.Query("query"))
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (handler *Handler) getCard(ctx *gin.Context) {
	card, err := handler.study.CardByID(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, card)
}

func (handler *Handler) createCard(ctx *gin.Context) {
	var payload cardRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	card, err := handler.study.CreateCard(ctx.Request.Context(), cardInput(payload))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, card)
}

func (handler *Handler) updateCard(ctx *gin.Context) {
	var payload cardPatchRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	card, err := handler.study.UpdateCard(ctx.Request.Context(), ctx.Param("id"), cardPatchInput(payload))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, card)
}

func (handler *Handler) updateCards(ctx *gin.Context) {
	var payload bulkCardUpdateRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	cards, err := handler.study.UpdateCards(ctx.Request.Context(), payload.IDs, cardPatchInput(payload.Patch))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, cards)
}

func (handler *Handler) deleteCard(ctx *gin.Context) {
	if err := handler.study.DeleteCard(ctx.Request.Context(), ctx.Param("id")); err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (handler *Handler) deleteCards(ctx *gin.Context) {
	var payload bulkDeleteRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	deleted, err := handler.study.DeleteCards(ctx.Request.Context(), payload.IDs)
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": deleted})
}

func (handler *Handler) listDueCards(ctx *gin.Context) {
	cards, err := handler.study.DueCards(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, cards)
}

func (handler *Handler) listDifficultCards(ctx *gin.Context) {
	cards, err := handler.study.DifficultCards(ctx.Request.Context())
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

func (handler *Handler) authMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenValue := bearerToken(ctx.GetHeader("Authorization"))
		if tokenValue == "" {
			writeError(ctx, http.StatusUnauthorized, errors.New("missing bearer token"))
			ctx.Abort()
			return
		}

		claims, err := handler.auth.ParseToken(tokenValue)
		if err != nil {
			writeError(ctx, http.StatusUnauthorized, err)
			ctx.Abort()
			return
		}

		ctx.Set("userID", claims.UserID)
		ctx.Set("isAdmin", claims.IsAdmin)
		ctx.Next()
	}
}

func requireAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		isAdmin, _ := ctx.Get("isAdmin")
		if isAdmin != true {
			writeError(ctx, http.StatusForbidden, service.ErrForbidden)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func bearerToken(header string) string {
	prefix := "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func currentUserID(ctx *gin.Context) string {
	value, _ := ctx.Get("userID")
	userID, _ := value.(string)
	return userID
}

func writeError(ctx *gin.Context, status int, err error) {
	ctx.JSON(status, gin.H{"error": err.Error()})
}

func writeAuthError(ctx *gin.Context, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, repository.ErrNotFound) {
		status = http.StatusNotFound
	}
	if errors.Is(err, repository.ErrConflict) {
		status = http.StatusConflict
	}
	if errors.Is(err, service.ErrForbidden) {
		status = http.StatusForbidden
	}
	writeError(ctx, status, err)
}

func writeResourceError(ctx *gin.Context, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, repository.ErrNotFound) {
		status = http.StatusNotFound
	}
	if errors.Is(err, repository.ErrConflict) {
		status = http.StatusConflict
	}
	writeError(ctx, status, err)
}

func deckPatchInput(payload deckPatchRequest) service.DeckPatchInput {
	return service.DeckPatchInput{
		Name:        payload.Name,
		Description: payload.Description,
	}
}

func cardInput(payload cardRequest) service.CardInput {
	return service.CardInput{
		DeckID:             payload.DeckID,
		Kind:               core.CardKind(payload.Kind),
		Korean:             payload.Korean,
		Translation:        payload.Translation,
		Romanization:       payload.Romanization,
		ExampleKorean:      payload.ExampleKorean,
		ExampleTranslation: payload.ExampleTranslation,
		Tags:               payload.Tags,
	}
}

func cardPatchInput(payload cardPatchRequest) service.CardPatchInput {
	var kind *core.CardKind
	if payload.Kind != nil {
		value := core.CardKind(*payload.Kind)
		kind = &value
	}

	return service.CardPatchInput{
		DeckID:             payload.DeckID,
		Kind:               kind,
		Korean:             payload.Korean,
		Translation:        payload.Translation,
		Romanization:       payload.Romanization,
		ExampleKorean:      payload.ExampleKorean,
		ExampleTranslation: payload.ExampleTranslation,
		Tags:               payload.Tags,
	}
}
