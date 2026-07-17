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
	study := service.NewStudyService(store, store, store, store, store, core.NewScheduler(), nil)
	return NewRouter(study, auth, service.NewAdminService(store), service.NewClientBackupService(store)), store
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
	adminAfterReset := performJSON(t, router, http.MethodGet, "/user/me", nil, admin.Token)
	if adminAfterReset.Code != http.StatusOK {
		t.Fatalf("admin must survive reset, status=%d body=%s", adminAfterReset.Code, adminAfterReset.Body.String())
	}
}

func TestProtectedRouteRejectsMissingToken(t *testing.T) {
	router, _ := testRouter(t)
	response := performJSON(t, router, http.MethodGet, "/api/cards", nil, "")
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestClientBackupIsValidatedAndScopedByUser(t *testing.T) {
	router, _ := testRouter(t)
	ownerToken := registerLearner(t, router, "backup-owner@example.test")
	otherToken := registerLearner(t, router, "backup-other@example.test")

	missing := performJSON(t, router, http.MethodGet, "/api/client-backup", nil, ownerToken)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing backup status=%d body=%s", missing.Code, missing.Body.String())
	}

	invalid := performJSON(t, router, http.MethodPut, "/api/client-backup", map[string]any{
		"config": []string{"not", "an", "object"},
		"state":  map[string]any{"activeView": "home"},
	}, ownerToken)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid backup status=%d body=%s", invalid.Code, invalid.Body.String())
	}

	saved := performJSON(t, router, http.MethodPut, "/api/client-backup", map[string]any{
		"config": map[string]any{"apiUrl": "https://example.test", "theme": "ocean"},
		"state":  map[string]any{"activeView": "lessons", "studyDirection": "french-to-korean"},
	}, ownerToken)
	if saved.Code != http.StatusOK {
		t.Fatalf("save backup status=%d body=%s", saved.Code, saved.Body.String())
	}

	loaded := performJSON(t, router, http.MethodGet, "/api/client-backup", nil, ownerToken)
	if loaded.Code != http.StatusOK {
		t.Fatalf("load backup status=%d body=%s", loaded.Code, loaded.Body.String())
	}
	var backup map[string]json.RawMessage
	if err := json.Unmarshal(loaded.Body.Bytes(), &backup); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(backup["config"], []byte(`"theme":"ocean"`)) || !bytes.Contains(backup["state"], []byte(`"activeView":"lessons"`)) {
		t.Fatalf("unexpected backup: %s", loaded.Body.String())
	}

	other := performJSON(t, router, http.MethodGet, "/api/client-backup", nil, otherToken)
	if other.Code != http.StatusNotFound {
		t.Fatalf("another user accessed backup: status=%d body=%s", other.Code, other.Body.String())
	}
}

func TestAdminCanListAndUpdateOnlyNonAdminUsers(t *testing.T) {
	router, _ := testRouter(t)
	learnerToken := registerLearner(t, router, "managed@example.test")
	forbidden := performJSON(t, router, http.MethodGet, "/admin/users", nil, learnerToken)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("non-admin list status=%d body=%s", forbidden.Code, forbidden.Body.String())
	}

	login := performJSON(t, router, http.MethodPost, "/user/login", map[string]string{
		"email": "admin@example.test", "password": "admin-password",
	}, "")
	var admin authResponse
	if err := json.Unmarshal(login.Body.Bytes(), &admin); err != nil {
		t.Fatal(err)
	}
	listed := performJSON(t, router, http.MethodGet, "/admin/users", nil, admin.Token)
	if listed.Code != http.StatusOK {
		t.Fatalf("admin list status=%d body=%s", listed.Code, listed.Body.String())
	}
	var users []map[string]any
	if err := json.Unmarshal(listed.Body.Bytes(), &users); err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0]["email"] != "managed@example.test" || users[0]["passwordHash"] != nil {
		t.Fatalf("unexpected public users: %#v", users)
	}
	userID, _ := users[0]["id"].(string)
	updated := performJSON(t, router, http.MethodPut, "/admin/users/"+userID, map[string]string{
		"name": "Student", "email": "student@example.test",
	}, admin.Token)
	if updated.Code != http.StatusOK || !bytes.Contains(updated.Body.Bytes(), []byte("student@example.test")) {
		t.Fatalf("admin update status=%d body=%s", updated.Code, updated.Body.String())
	}
	adminUpdate := performJSON(t, router, http.MethodPut, "/admin/users/admin", map[string]string{"name": "Changed"}, admin.Token)
	if adminUpdate.Code != http.StatusForbidden {
		t.Fatalf("admin account must be protected, got %d", adminUpdate.Code)
	}
}

func TestUserCanUpdateOwnProfileAndPassword(t *testing.T) {
	router, _ := testRouter(t)
	token := registerLearner(t, router, "profile@example.test")

	updated := performJSON(t, router, http.MethodPut, "/user/me", map[string]string{
		"name": "Updated learner", "email": "updated@example.test", "password": "new-password-123",
	}, token)
	if updated.Code != http.StatusOK {
		t.Fatalf("profile update status=%d body=%s", updated.Code, updated.Body.String())
	}
	if bytes.Contains(updated.Body.Bytes(), []byte("password")) || !bytes.Contains(updated.Body.Bytes(), []byte("updated@example.test")) {
		t.Fatalf("unsafe or incomplete profile response: %s", updated.Body.String())
	}

	login := performJSON(t, router, http.MethodPost, "/user/login", map[string]string{
		"email": "updated@example.test", "password": "new-password-123",
	}, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login with updated credentials status=%d body=%s", login.Code, login.Body.String())
	}
}

func TestRegistrationNormalizesEmailBeforeValidation(t *testing.T) {
	router, _ := testRouter(t)
	registered := performJSON(t, router, http.MethodPost, "/user/register", map[string]string{
		"name": "Test user", "email": "  test@gmail.com  ", "password": "password-123",
	}, "")
	if registered.Code != http.StatusCreated || !bytes.Contains(registered.Body.Bytes(), []byte(`"email":"test@gmail.com"`)) {
		t.Fatalf("normalized registration status=%d body=%s", registered.Code, registered.Body.String())
	}

	invalid := performJSON(t, router, http.MethodPost, "/user/register", map[string]string{
		"name": "Test user", "email": "not-an-email", "password": "password-123",
	}, "")
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid email should return 400, got %d body=%s", invalid.Code, invalid.Body.String())
	}
}
