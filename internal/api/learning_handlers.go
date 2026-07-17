package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/arthurblanchet59/korean-learning-go/internal/service"
)

func (handler *Handler) listLessons(ctx *gin.Context) {
	lessons, err := handler.study.ListLessons(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		writeInternalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, lessons)
}

func (handler *Handler) getLesson(ctx *gin.Context) {
	lesson, err := handler.study.LessonByID(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id"))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, lesson)
}

func (handler *Handler) updateLessonProgress(ctx *gin.Context) {
	var payload lessonProgressRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	progress, err := handler.study.UpdateLessonProgress(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id"), payload.Completed, payload.Score)
	if err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, progress)
}

func (handler *Handler) listJournalEntries(ctx *gin.Context) {
	entries, err := handler.study.ListJournalEntries(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		writeInternalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, entries)
}

func (handler *Handler) createJournalEntry(ctx *gin.Context) {
	var payload journalRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	entry, err := handler.study.CreateJournalEntry(ctx.Request.Context(), currentUserID(ctx), service.JournalInput{Title: payload.Title, Text: payload.Text})
	if err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, entry)
}

func (handler *Handler) getJournalEntry(ctx *gin.Context) {
	entry, err := handler.study.JournalEntryByID(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id"))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, entry)
}

func (handler *Handler) correctJournalText(ctx *gin.Context) {
	var payload journalRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	corrected, corrections := service.CorrectKorean(payload.Text)
	ctx.JSON(http.StatusOK, gin.H{"correctedText": corrected, "corrections": corrections})
}

func (handler *Handler) updateJournalEntry(ctx *gin.Context) {
	var payload journalRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	entry, err := handler.study.UpdateJournalEntry(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id"), service.JournalInput{Title: payload.Title, Text: payload.Text})
	if err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, entry)
}

func (handler *Handler) deleteJournalEntry(ctx *gin.Context) {
	if err := handler.study.DeleteJournalEntry(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id")); err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
