package main

import (
	"log"
	"net/http"
	"time"

	httpapi "github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/http"
	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/repository/memory"
	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/service"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func main() {
	now := time.Now().UTC()
	store := memory.NewStore(
		[]core.Deck{core.SeedDeck(now)},
		core.SeedCards(now),
	)

	studyService := service.NewStudyService(store, store, core.NewScheduler())
	handler := httpapi.NewHandler(studyService)

	addr := ":8080"
	log.Printf("korean-learning backend listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, handler.Routes()); err != nil {
		log.Fatal(err)
	}
}
