package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/internal/service"
)

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
		user, err := handler.auth.UserByID(ctx.Request.Context(), claims.UserID)
		if err != nil {
			writeError(ctx, http.StatusUnauthorized, service.ErrInvalidCredentials)
			ctx.Abort()
			return
		}

		ctx.Set("userID", user.ID)
		ctx.Set("isAdmin", user.IsAdmin)
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
	if errors.Is(err, service.ErrCorrectionUnavailable) || errors.Is(err, service.ErrEmbeddingUnavailable) {
		status = http.StatusBadGateway
	}
	writeError(ctx, status, err)
}
