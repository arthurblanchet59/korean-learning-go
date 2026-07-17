package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/arthurblanchet59/korean-learning-go/internal/backend/service"
)

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
		writeAuthError(ctx, err)
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
		writeAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) me(ctx *gin.Context) {
	user, err := handler.auth.UserByID(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		writeAuthError(ctx, err)
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

func (handler *Handler) adminListUsers(ctx *gin.Context) {
	users, err := handler.auth.ListUsers(ctx.Request.Context())
	if err != nil {
		writeInternalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, users)
}

func (handler *Handler) resetDatabase(ctx *gin.Context) {
	result, err := handler.admin.ResetDatabase(ctx.Request.Context())
	if err != nil {
		writeInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "database reset completed",
		"result":  result,
	})
}
