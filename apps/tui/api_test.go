package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/packages/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func keyEnter() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyEnter}
}

func TestSplitFields(t *testing.T) {
	fields := splitFields("deck-id | 안녕하세요 | bonjour")
	if len(fields) != 3 || fields[1] != "안녕하세요" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestVisibleBoundsKeepsCursorVisible(t *testing.T) {
	start, end := visibleBounds(100, 52, 10)
	if start > 52 || end <= 52 || end-start != 10 {
		t.Fatalf("cursor not visible in [%d:%d]", start, end)
	}
}

func TestStudyViewHidesAnswerUntilReveal(t *testing.T) {
	card := core.Card{Korean: "집", Translation: "maison", Romanization: "jip"}
	m := model{data: DashboardData{Due: []core.Card{card}}, studyDirection: "korean-to-french"}

	hidden := m.studyView(100, 24)
	if strings.Contains(hidden, "maison") || strings.Contains(hidden, "jip") {
		t.Fatalf("answer leaked before reveal: %q", hidden)
	}

	m.revealed = true
	revealed := m.studyView(100, 24)
	if !strings.Contains(revealed, "maison") || !strings.Contains(revealed, "jip") {
		t.Fatalf("revealed answer is incomplete: %q", revealed)
	}
}

func TestHomeViewShowsMainApplicationOptions(t *testing.T) {
	m := model{data: DashboardData{User: User{Name: "Arthur"}, Stats: core.StudyStats{DueCards: 4}}}
	view := m.homeView(110, 28)
	for _, expected := range []string{"Réviser maintenant", "Gérer ma bibliothèque", "Continuer les leçons", "Thème, API et backup"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("home screen is missing %q: %q", expected, view)
		}
	}
}

func TestSettingsViewExposesThemesLocalFilesAndBackupActions(t *testing.T) {
	m := model{config: defaultConfig(), activeUserID: "user-1", tab: tabSettings}
	view := m.settingsView(120, 30)
	for _, expected := range []string{"Émeraude", "Océan", "Ambre", "Rose", "config.json", "state.json", "u envoyer", "o restaurer"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("settings screen is missing %q: %q", expected, view)
		}
	}
}

func TestLessonsViewUsesCompletionStateWithoutScore(t *testing.T) {
	lesson := Lesson{
		Lesson:   core.Lesson{ID: "lesson", Title: "Le présent", Description: "Une description", Level: "A1", Content: strings.Repeat("Contenu détaillé de la leçon. ", 30)},
		Progress: core.LessonProgress{Completed: true, Score: 42, UpdatedAt: time.Now()},
	}
	m := model{data: DashboardData{Lessons: []Lesson{lesson}}}

	view := m.lessonsView(100, 20)
	if strings.Contains(view, "Score") || strings.Contains(view, "42") {
		t.Fatalf("lesson score should not be displayed: %q", view)
	}
	if !strings.Contains(view, "Terminée") || !strings.Contains(view, "PgUp/PgDn") {
		t.Fatalf("completion or scrolling hint missing: %q", view)
	}
}

func TestScrollableTextRespectsViewport(t *testing.T) {
	view := scrollableText(strings.Repeat("mot ", 80), 30, 8, 0)
	if lines := strings.Count(view, "\n") + 1; lines > 8 {
		t.Fatalf("viewport contains %d lines, expected at most 8", lines)
	}
}

func TestLessonCompletionDoesNotSendScore(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/lessons/lesson-1/progress" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write([]byte(`{"completed":true}`))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL)
	client.Token = "test-token"
	if _, err := client.Execute("lesson-complete lesson-1"); err != nil {
		t.Fatal(err)
	}
	if payload["completed"] != true {
		t.Fatalf("completion flag missing: %#v", payload)
	}
	if _, exists := payload["score"]; exists {
		t.Fatalf("score should not be sent by the TUI: %#v", payload)
	}
}

func TestAdminTabIsVisibleOnlyForAdmins(t *testing.T) {
	regular := model{data: DashboardData{User: User{IsAdmin: false}}}
	if len(regular.visibleTabs()) != len(tabs) {
		t.Fatal("regular users must not see the admin tab")
	}
	admin := model{data: DashboardData{User: User{IsAdmin: true}}}
	visible := admin.visibleTabs()
	if len(visible) != len(tabs)+1 || visible[len(visible)-1] != "ADMIN" {
		t.Fatalf("admin tab missing: %#v", visible)
	}
}

func TestAdminViewListsManageableUsers(t *testing.T) {
	m := model{client: &APIClient{BaseURL: "http://localhost:8080"}, data: DashboardData{
		User:  User{IsAdmin: true},
		Users: []User{{ID: "user-1", Name: "Arthur", Email: "arthur@example.test"}},
	}}
	view := m.adminView(110, 24)
	if !strings.Contains(view, "Arthur") || !strings.Contains(view, "arthur@example.test") || !strings.Contains(view, "e modifier") {
		t.Fatalf("admin user is not manageable from the view: %q", view)
	}
}

func TestAdminEditorUsesGuidedFields(t *testing.T) {
	m := model{
		adminEditing: true,
		adminField:   1,
		adminName:    "Arthur",
		adminEmail:   "arthur@example.test",
	}
	view := m.adminView(110, 28)
	for _, expected := range []string{"MODIFIER UN UTILISATEUR", "NOM", "EMAIL", "NOUVEAU MOT DE PASSE", "échap annuler"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("admin editor is missing %q: %q", expected, view)
		}
	}
}

func TestAdminUpdateUserSendsOptionalPassword(t *testing.T) {
	var payload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut || request.URL.Path != "/admin/users/user-1" {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL)
	client.Token = "admin-token"
	if err := client.AdminUpdateUser("user-1", "Arthur B.", "arthur.b@example.test", "new-password"); err != nil {
		t.Fatal(err)
	}
	if payload["name"] != "Arthur B." || payload["email"] != "arthur.b@example.test" || payload["password"] != "new-password" {
		t.Fatalf("unexpected admin update payload: %#v", payload)
	}
}

func TestLogoutAlwaysReturnsToLoginMode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	m := model{
		client:      &APIClient{Token: "token"},
		loggedIn:    true,
		registering: true,
		loginName:   "Old name",
		loginEmail:  "old@example.test",
		loginPass:   "old-password",
		loginField:  2,
	}

	loggedOut := m.logout()
	if loggedOut.loggedIn || loggedOut.registering || loggedOut.loginField != 0 {
		t.Fatalf("logout did not return to login mode: %#v", loggedOut)
	}
	if loggedOut.loginName != "" || loggedOut.loginEmail != "" || loggedOut.loginPass != "" {
		t.Fatalf("logout kept authentication fields: %#v", loggedOut)
	}
}

func TestUpdateProfileSendsEditableFields(t *testing.T) {
	var payload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut || request.URL.Path != "/user/me" {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing bearer token: %q", request.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write([]byte(`{"id":"user-1","name":"Arthur B.","email":"arthur.b@example.test"}`))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL)
	client.Token = "test-token"
	if err := client.UpdateProfile("Arthur B.", "arthur.b@example.test", "new-password"); err != nil {
		t.Fatal(err)
	}
	if payload["name"] != "Arthur B." || payload["email"] != "arthur.b@example.test" || payload["password"] != "new-password" {
		t.Fatalf("unexpected profile payload: %#v", payload)
	}
}

func TestProfileViewExposesEditAndLogoutActions(t *testing.T) {
	m := model{data: DashboardData{User: User{ID: "user-1", Name: "Arthur", Email: "arthur@example.test"}}}
	view := m.profileView(100, 24)
	if !strings.Contains(view, "e modifier mes informations") || !strings.Contains(view, "D se déconnecter") {
		t.Fatalf("profile actions are missing: %q", view)
	}
}

func TestJournalViewExposesGuidedActions(t *testing.T) {
	m := model{data: DashboardData{Journal: []core.JournalEntry{{
		ID: "journal-1", Title: "Ma journée", OriginalText: "저는 공부해요.", CorrectedText: "저는 공부해요.",
	}}}}
	view := m.journalView(110, 28)
	for _, expected := range []string{"n nouvelle entrée", "e modifier l'entrée", "d supprimer"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("journal action %q is missing: %q", expected, view)
		}
	}
	if strings.Contains(view, "journal-update") {
		t.Fatalf("technical command is still exposed: %q", view)
	}
}

func TestJournalEditKeyPrefillsGuidedEditor(t *testing.T) {
	m := model{
		tab: tabJournal,
		data: DashboardData{Journal: []core.JournalEntry{{
			ID: "journal-1", Title: "Ma journée", OriginalText: "저는 공부해요.",
		}}},
	}
	updated, _ := m.updateNavigation(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	editor := updated.(model)
	if !editor.journalEditing || editor.journalID != "journal-1" || editor.journalTitle != "Ma journée" || editor.journalText != "저는 공부해요." {
		t.Fatalf("journal editor was not prefilled: %#v", editor)
	}
	view := editor.journalView(110, 28)
	if !strings.Contains(view, "MODIFIER L'ENTRÉE DU JOURNAL") || !strings.Contains(view, "TEXTE EN CORÉEN") {
		t.Fatalf("guided journal editor is incomplete: %q", view)
	}
}

func TestSaveJournalEntryUpdatesSelectedEntry(t *testing.T) {
	var payload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut || request.URL.Path != "/api/journal/journal-1" {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL)
	client.Token = "test-token"
	if err := client.SaveJournalEntry("journal-1", " Ma journée ", " 저는 공부해요. "); err != nil {
		t.Fatal(err)
	}
	if payload["title"] != "Ma journée" || payload["text"] != "저는 공부해요." {
		t.Fatalf("unexpected journal payload: %#v", payload)
	}
}

func TestRegisterTrimsUserInput(t *testing.T) {
	var payload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		response.WriteHeader(http.StatusBadRequest)
		_, _ = response.Write([]byte(`{"error":"test response"}`))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL)
	_, _ = client.Register("  Test user  ", "  te\u200bst@gmail.com  ", "password-123")
	if payload["name"] != "Test user" || payload["email"] != "test@gmail.com" {
		t.Fatalf("registration input was not normalized: %#v", payload)
	}
}

func TestRegistrationFormSubmitsDisplayedValues(t *testing.T) {
	var payload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		response.WriteHeader(http.StatusBadRequest)
		_, _ = response.Write([]byte(`{"error":"test response"}`))
	}))
	defer server.Close()

	m := model{
		client:      NewAPIClient(server.URL),
		registering: true,
		loginField:  2,
		loginName:   "testabt",
		loginEmail:  "test@gmail.com",
		loginPass:   "testtest",
	}
	_, command := m.updateLogin(tea.KeyMsg{Type: tea.KeyEnter})
	if command == nil {
		t.Fatal("registration form did not submit")
	}
	message, ok := command().(loginMsg)
	if !ok || message.err == nil {
		t.Fatalf("unexpected command result: %#v", message)
	}
	if payload["name"] != "testabt" || payload["email"] != "test@gmail.com" || payload["password"] != "testtest" {
		t.Fatalf("form values changed before submission: %#v", payload)
	}
}

func TestLoginUpMovesToPreviousField(t *testing.T) {
	m := model{registering: true, loginField: 0}

	updated, _ := m.updateLogin(tea.KeyMsg{Type: tea.KeyUp})
	if field := updated.(model).loginField; field != 2 {
		t.Fatalf("up should select the previous field, got %d", field)
	}
}

func TestLibraryViewExposesEveryAvailableAction(t *testing.T) {
	m := model{
		libraryCards: true,
		data: DashboardData{Cards: []core.Card{{
			ID: "card-1", Korean: "집", Translation: "maison",
		}}},
	}

	view := m.libraryView(110, 28)
	for _, expected := range []string{"n nouveau", "e modifier", "d supprimer", "c options avancées"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("library action %q is missing: %q", expected, view)
		}
	}
}

func TestLibraryAdvancedShortcutOpensCommandInput(t *testing.T) {
	m := model{tab: tabLibrary, input: "ancienne commande"}

	updated, _ := m.updateNavigation(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	commandModel := updated.(model)
	if commandModel.inputMode != "command" || commandModel.input != "" {
		t.Fatalf("advanced shortcut did not open a clean command input: %#v", commandModel)
	}
}

func TestLibraryEditPrefillsSelectedCard(t *testing.T) {
	m := model{
		tab:          tabLibrary,
		libraryCards: true,
		data: DashboardData{Cards: []core.Card{{
			ID: "card-1", Korean: "집", Translation: "maison", Romanization: "jip",
		}}},
	}

	updated, _ := m.updateNavigation(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	commandModel := updated.(model)
	expected := "card-update card-1 | 집 | maison | jip"
	if commandModel.inputMode != "command" || commandModel.input != expected {
		t.Fatalf("edit shortcut prefilled %q, expected %q", commandModel.input, expected)
	}
}

func TestHelpBlocksBackgroundActions(t *testing.T) {
	m := model{tab: tabJournal, showHelp: true}

	updated, command := m.updateNavigation(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	helpModel := updated.(model)
	if command != nil || helpModel.journalEditing || !helpModel.showHelp {
		t.Fatalf("help allowed a background action: %#v", helpModel)
	}

	updated, _ = helpModel.updateNavigation(tea.KeyMsg{Type: tea.KeyEsc})
	if updated.(model).showHelp {
		t.Fatal("escape did not close the help")
	}
}

func TestHelpCanScrollInSmallTerminal(t *testing.T) {
	m := model{showHelp: true, height: 24}

	firstPage := m.helpView(80, 16)
	updated, _ := m.updateNavigation(tea.KeyMsg{Type: tea.KeyPgDown})
	scrolled := updated.(model)
	secondPage := scrolled.helpView(80, 16)

	if scrolled.detailScroll == 0 || firstPage == secondPage {
		t.Fatalf("help did not scroll: offset=%d", scrolled.detailScroll)
	}
}

func TestFooterUsesCompactHintsWhenStatusNeedsSpace(t *testing.T) {
	m := model{status: "Une opération vient de se terminer correctement"}

	view := m.footerView(80)
	if lipgloss.Width(view) > 80 {
		t.Fatalf("footer exceeds its viewport: width=%d", lipgloss.Width(view))
	}
}

func TestBackupClientUsesProtectedEndpoint(t *testing.T) {
	var method, authorization string
	var payload map[string]json.RawMessage
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		method = request.Method
		authorization = request.Header.Get("Authorization")
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"config":{"version":1,"apiUrl":"https://example.test","theme":"ocean"},"state":{"version":1,"activeView":"home","studyDirection":"korean-to-french","libraryCards":true},"updatedAt":"2026-07-15T08:00:00Z"}`))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL)
	client.Token = "backup-token"
	backup, err := client.UploadBackup(
		AppConfig{Version: 1, APIURL: "https://example.test", Theme: "ocean"},
		AppState{Version: 1, ActiveView: "home", StudyDirection: "korean-to-french", LibraryCards: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPut || authorization != "Bearer backup-token" {
		t.Fatalf("unexpected request: %s %s", method, authorization)
	}
	if len(payload["config"]) == 0 || len(payload["state"]) == 0 || backup.Config.Theme != "ocean" {
		t.Fatalf("unexpected backup exchange: payload=%#v backup=%#v", payload, backup)
	}
}
