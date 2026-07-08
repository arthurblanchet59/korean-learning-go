package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

type appState struct {
	mu        sync.RWMutex
	decks     []core.Deck
	cards     []core.Card
	scheduler core.Scheduler
}

type answerRequest struct {
	Rating string `json:"rating"`
}

func main() {
	now := time.Now().UTC()
	state := &appState{
		decks:     []core.Deck{core.SeedDeck(now)},
		cards:     core.SeedCards(now),
		scheduler: core.NewScheduler(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /api/decks", state.handleDecks)
	mux.HandleFunc("GET /api/cards", state.handleCards)
	mux.HandleFunc("GET /api/reviews/due", state.handleDueCards)
	mux.HandleFunc("POST /api/reviews/", state.handleReviewAnswer)
	mux.HandleFunc("GET /api/stats", state.handleStats)

	handler := withCORS(mux)
	addr := ":8080"

	log.Printf("korean-learning backend listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (state *appState) handleDecks(w http.ResponseWriter, r *http.Request) {
	state.mu.RLock()
	defer state.mu.RUnlock()

	writeJSON(w, http.StatusOK, state.decks)
}

func (state *appState) handleCards(w http.ResponseWriter, r *http.Request) {
	state.mu.RLock()
	defer state.mu.RUnlock()

	writeJSON(w, http.StatusOK, state.cards)
}

func (state *appState) handleDueCards(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()

	state.mu.RLock()
	defer state.mu.RUnlock()

	due := make([]core.Card, 0)
	for _, card := range state.cards {
		if card.Due(now) {
			due = append(due, card)
		}
	}

	writeJSON(w, http.StatusOK, due)
}

func (state *appState) handleStats(w http.ResponseWriter, r *http.Request) {
	state.mu.RLock()
	defer state.mu.RUnlock()

	writeJSON(w, http.StatusOK, core.BuildStats(state.cards, time.Now().UTC()))
}

func (state *appState) handleReviewAnswer(w http.ResponseWriter, r *http.Request) {
	cardID, err := cardIDFromPath(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var payload answerRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	rating, err := core.ParseRating(payload.Rating)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	for index := range state.cards {
		if state.cards[index].ID != cardID {
			continue
		}

		previous := state.cards[index].ReviewState
		next := state.scheduler.Schedule(previous, rating, time.Now().UTC())
		state.cards[index].ReviewState = next

		writeJSON(w, http.StatusOK, core.Review{
			ID:         "review-" + cardID,
			CardID:     cardID,
			Rating:     rating,
			ReviewedAt: next.LastReviewAt,
			Previous:   previous,
			Next:       next,
		})
		return
	}

	writeError(w, http.StatusNotFound, errors.New("card not found"))
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
