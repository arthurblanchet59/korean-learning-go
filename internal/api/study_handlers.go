package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/arthurblanchet59/korean-learning-go/internal/service"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func (handler *Handler) listDecks(ctx *gin.Context) {
	decks, err := handler.study.ListDecks(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		writeInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, decks)
}

func (handler *Handler) searchDecks(ctx *gin.Context) {
	results, err := handler.study.SearchDecks(ctx.Request.Context(), currentUserID(ctx), ctx.Query("query"))
	if err != nil {
		writeInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (handler *Handler) searchAll(ctx *gin.Context) {
	results, err := handler.study.SearchAll(ctx.Request.Context(), currentUserID(ctx), ctx.Query("query"))
	if err != nil {
		writeInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (handler *Handler) getDeck(ctx *gin.Context) {
	deck, err := handler.study.DeckByID(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id"))
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

	deck, err := handler.study.CreateDeck(ctx.Request.Context(), currentUserID(ctx), service.DeckInput{
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

	deck, err := handler.study.UpdateDeck(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id"), deckPatchInput(payload))
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

	decks, err := handler.study.UpdateDecks(ctx.Request.Context(), currentUserID(ctx), payload.IDs, deckPatchInput(payload.Patch))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, decks)
}

func (handler *Handler) deleteDeck(ctx *gin.Context) {
	if err := handler.study.DeleteDeck(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id")); err != nil {
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

	deleted, err := handler.study.DeleteDecks(ctx.Request.Context(), currentUserID(ctx), payload.IDs)
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": deleted})
}

func (handler *Handler) listCards(ctx *gin.Context) {
	cards, err := handler.study.ListCards(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		writeInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, cards)
}

func (handler *Handler) searchCards(ctx *gin.Context) {
	results, err := handler.study.SearchCards(ctx.Request.Context(), currentUserID(ctx), ctx.Query("query"))
	if err != nil {
		writeInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (handler *Handler) getCard(ctx *gin.Context) {
	card, err := handler.study.CardByID(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id"))
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

	card, err := handler.study.CreateCard(ctx.Request.Context(), currentUserID(ctx), cardInput(payload))
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

	card, err := handler.study.UpdateCard(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id"), cardPatchInput(payload))
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

	cards, err := handler.study.UpdateCards(ctx.Request.Context(), currentUserID(ctx), payload.IDs, cardPatchInput(payload.Patch))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, cards)
}

func (handler *Handler) deleteCard(ctx *gin.Context) {
	if err := handler.study.DeleteCard(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id")); err != nil {
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

	deleted, err := handler.study.DeleteCards(ctx.Request.Context(), currentUserID(ctx), payload.IDs)
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": deleted})
}

func (handler *Handler) listDueCards(ctx *gin.Context) {
	cards, err := handler.study.DueCards(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		writeInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, cards)
}

func (handler *Handler) listDifficultCards(ctx *gin.Context) {
	cards, err := handler.study.DifficultCards(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		writeInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, cards)
}

func (handler *Handler) stats(ctx *gin.Context) {
	stats, err := handler.study.Stats(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		writeInternalError(ctx, err)
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

	review, err := handler.study.AnswerCard(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id"), payload.Rating)
	if err != nil {
		writeResourceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, review)
}

func (handler *Handler) checkAnswer(ctx *gin.Context) {
	var payload answerCheckRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := handler.study.CheckAnswer(ctx.Request.Context(), currentUserID(ctx), ctx.Param("id"), payload.Answer, payload.Direction)
	if err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (handler *Handler) exportCards(ctx *gin.Context) {
	content, err := handler.study.ExportCardsCSV(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		writeInternalError(ctx, err)
		return
	}
	ctx.Header("Content-Disposition", `attachment; filename="korean-cards.csv"`)
	ctx.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(content))
}

func (handler *Handler) importCards(ctx *gin.Context) {
	var payload importCardsRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	cards, err := handler.study.ImportCardsCSV(ctx.Request.Context(), currentUserID(ctx), payload.DeckID, payload.CSV)
	if err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"imported": len(cards), "cards": cards})
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
