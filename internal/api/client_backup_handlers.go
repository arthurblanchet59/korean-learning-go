package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/internal/service"
)

func (handler *Handler) getClientBackup(ctx *gin.Context) {
	backup, err := handler.backup.Backup(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(ctx, http.StatusNotFound, err)
			return
		}
		writeInternalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, backup)
}

func (handler *Handler) saveClientBackup(ctx *gin.Context) {
	var payload clientBackupRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	backup, err := handler.backup.Save(ctx.Request.Context(), currentUserID(ctx), payload.Config, payload.State)
	if err != nil {
		if errors.Is(err, service.ErrInvalidClientBackup) {
			writeError(ctx, http.StatusBadRequest, err)
			return
		}
		writeInternalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, backup)
}
