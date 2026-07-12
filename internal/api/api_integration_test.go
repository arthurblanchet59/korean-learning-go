package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	sqliterepo "github.com/arthurblanchet59/korean-learning-go/internal/repository/sqlite"
	"github.com/arthurblanchet59/korean-learning-go/internal/service"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

type authResponse struct {
	Token string `json:"token"`
}

func testRouter(t *testing.T) (*gin.Engine, *sqliterepo.Store) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	store, err := sqliterepo.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.Migrate(); err != nil {
		t.Fatal(err)
	}
	auth := service.NewAuthService(store, "integration-test-secret-with-enough-entropy")
	if err := auth.EnsureAdmin(context.Background(), "Admin", "admin@example.test", "admin-password"); err != nil {
		t.Fatal(err)
	}
	study := service.NewStudyService(store, store, store, store, store, core.NewScheduler())
	return NewRouter(study, auth, service.NewAdminService(store)), store
}

func performJSON(t *testing.T, router http.Handler, method string, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var payload *bytes.Reader
	if body == nil {
		payload = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		payload = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, path, payload)
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func registerLearner(t *testing.T, router http.Handler, email string) string {
	t.Helper()
	response := performJSON(t, router, http.MethodPost, "/user/register", map[string]any{
		"name": "Learner", "email": email, "password": "password-123",
	}, "")
	if response.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", response.Code, response.Body.String())
	}
	var result authResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result.Token
}

func TestAuthenticatedDeckCRUD(t *testing.T) {
	router, _ := testRouter(t)
	token := registerLearner(t, router, "learner@example.test")

	created := performJSON(t, router, http.MethodPost, "/api/decks", map[string]string{
		"name": "Voyage", "description": "Vocabulaire utile",
	}, token)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var deck core.Deck
	if err := json.Unmarshal(created.Body.Bytes(), &deck); err != nil {
		t.Fatal(err)
	}

	updated := performJSON(t, router, http.MethodPut, "/api/decks/"+deck.ID, map[string]string{"name": "Corée"}, token)
	if updated.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", updated.Code, updated.Body.String())
	}
	searched := performJSON(t, router, http.MethodGet, "/api/decks/search?query=Cor%C3%A9e", nil, token)
	if searched.Code != http.StatusOK || !bytes.Contains(searched.Body.Bytes(), []byte(deck.ID)) {
		t.Fatalf("search status=%d body=%s", searched.Code, searched.Body.String())
	}
	deleted := performJSON(t, router, http.MethodDelete, "/api/decks/"+deck.ID, nil, token)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", deleted.Code, deleted.Body.String())
	}
}

func TestResetInvalidatesDeletedUserToken(t *testing.T) {
	router, _ := testRouter(t)
	learnerToken := registerLearner(t, router, "reset@example.test")
	login := performJSON(t, router, http.MethodPost, "/user/login", map[string]string{
		"email": "admin@example.test", "password": "admin-password",
	}, "")
	var admin authResponse
	if err := json.Unmarshal(login.Body.Bytes(), &admin); err != nil {
		t.Fatal(err)
	}
	reset := performJSON(t, router, http.MethodPost, "/admin/reset", map[string]any{}, admin.Token)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status=%d body=%s", reset.Code, reset.Body.String())
	}
	afterReset := performJSON(t, router, http.MethodGet, "/api/decks", nil, learnerToken)
	if afterReset.Code != http.StatusUnauthorized {
		t.Fatalf("deleted user token should be rejected, got %d", afterReset.Code)
	}
}

func TestProtectedRouteRejectsMissingToken(t *testing.T) {
	router, _ := testRouter(t)
	response := performJSON(t, router, http.MethodGet, "/api/cards", nil, "")
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

