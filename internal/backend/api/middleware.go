package api

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/arthurblanchet59/korean-learning-go/internal/backend/repository"
	"github.com/arthurblanchet59/korean-learning-go/internal/backend/service"
)

func cors(allowedOrigin string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", allowedOrigin)
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

// writeInternalError logs the real error server-side and returns a generic 500
// so internal details (SQL errors, etc.) never leak to the client.
func writeInternalError(ctx *gin.Context, err error) {
	log.Printf("%s %s: %v", ctx.Request.Method, ctx.Request.URL.Path, err)
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func isValidationError(err error) bool {
	var validation service.ValidationError
	return errors.As(err, &validation)
}

func writeAuthError(ctx *gin.Context, err error) {
	switch {
	case isValidationError(err):
		writeError(ctx, http.StatusBadRequest, err)
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(ctx, http.StatusUnauthorized, err)
	case errors.Is(err, repository.ErrNotFound):
		writeError(ctx, http.StatusNotFound, err)
	case errors.Is(err, repository.ErrConflict):
		writeError(ctx, http.StatusConflict, err)
	case errors.Is(err, service.ErrForbidden):
		writeError(ctx, http.StatusForbidden, err)
	default:
		writeInternalError(ctx, err)
	}
}

func writeResourceError(ctx *gin.Context, err error) {
	switch {
	case isValidationError(err):
		writeError(ctx, http.StatusBadRequest, err)
	case errors.Is(err, repository.ErrNotFound):
		writeError(ctx, http.StatusNotFound, err)
	case errors.Is(err, repository.ErrConflict):
		writeError(ctx, http.StatusConflict, err)
	default:
		writeInternalError(ctx, err)
	}
}
