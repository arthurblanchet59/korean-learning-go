package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/service"
)

type Handler struct {
	study *service.StudyService
}

func NewHandler(study *service.StudyService) *Handler {
	return &Handler{study: study}
}

func (handler *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.health)
	mux.HandleFunc("GET /api/decks", handler.listDecks)
	mux.HandleFunc("GET /api/cards", handler.listCards)
	mux.HandleFunc("GET /api/reviews/due", handler.listDueCards)
	mux.HandleFunc("POST /api/reviews/", handler.answerCard)
	mux.HandleFunc("GET /api/stats", handler.stats)

	return withCORS(mux)
}

func (handler *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (handler *Handler) listDecks(w http.ResponseWriter, r *http.Request) {
	decks, err := handler.study.ListDecks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, decks)
}

func (handler *Handler) listCards(w http.ResponseWriter, r *http.Request) {
	cards, err := handler.study.ListCards(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, cards)
}

func (handler *Handler) listDueCards(w http.ResponseWriter, r *http.Request) {
	cards, err := handler.study.DueCards(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, cards)
}

func (handler *Handler) stats(w http.ResponseWriter, r *http.Request) {
	stats, err := handler.study.Stats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (handler *Handler) answerCard(w http.ResponseWriter, r *http.Request) {
	cardID, err := cardIDFromPath(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var payload answerRequest
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	review, err := handler.study.AnswerCard(r.Context(), cardID, payload.Rating)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, repository.ErrNotFound) {
			status = http.StatusNotFound
		}

		writeError(w, status, err)
		return
	}

	writeJSON(w, http.StatusOK, review)
}

func cardIDFromPath(path string) (string, error) {
	const prefix = "/api/reviews/"
	const suffix = "/answer"

	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return "", errors.New("invalid review answer path")
	}

	cardID := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	cardID = strings.Trim(cardID, "/")
	if cardID == "" {
		return "", errors.New("missing card id")
	}

	return cardID, nil
}
