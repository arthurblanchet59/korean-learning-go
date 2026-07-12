package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

type APIClient struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

type User struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`
}

type AuthResult struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type Lesson struct {
	core.Lesson
	Progress core.LessonProgress `json:"progress"`
}

type DashboardData struct {
	User      User
	Stats     core.StudyStats
	Due       []core.Card
	Cards     []core.Card
	Difficult []core.Card
	Decks     []core.Deck
	Lessons   []Lesson
	Journal   []core.JournalEntry
}

type SearchResult struct {
	Decks []core.Deck `json:"decks"`
	Cards []core.Card `json:"cards"`
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{BaseURL: strings.TrimRight(baseURL, "/"), Token: loadToken(), HTTP: &http.Client{Timeout: 12 * time.Second}}
}

func (client *APIClient) Login(email string, password string) (AuthResult, error) {
	var result AuthResult
	err := client.do(http.MethodPost, "/user/login", map[string]string{"email": email, "password": password}, &result)
	if err != nil {
		return AuthResult{}, err
	}
	client.Token = result.Token
	if err := saveToken(result.Token); err != nil {
		return AuthResult{}, err
	}
	return result, nil
}

func (client *APIClient) Register(name string, email string, password string) (AuthResult, error) {
	var result AuthResult
	err := client.do(http.MethodPost, "/user/register", map[string]string{"name": name, "email": email, "password": password}, &result)
	if err != nil {
		return AuthResult{}, err
	}
	client.Token = result.Token
	if err := saveToken(result.Token); err != nil {
		return AuthResult{}, err
	}
	return result, nil
}

func (client *APIClient) LoadDashboard() (DashboardData, error) {
	var data DashboardData
	requests := []struct {
		path string
		out  any
	}{
		{"/user/me", &data.User}, {"/api/stats", &data.Stats}, {"/api/reviews/due", &data.Due},
		{"/api/cards", &data.Cards}, {"/api/cards/difficult", &data.Difficult}, {"/api/decks", &data.Decks},
		{"/api/lessons", &data.Lessons}, {"/api/journal", &data.Journal},
	}
	for _, request := range requests {
		if err := client.do(http.MethodGet, request.path, nil, request.out); err != nil {
			return DashboardData{}, err
		}
	}
	return data, nil
}

func (client *APIClient) Answer(cardID string, rating string) error {
	return client.do(http.MethodPost, "/study/cards/"+url.PathEscape(cardID)+"/answer", map[string]string{"rating": rating}, nil)
}

func (client *APIClient) Check(cardID string, answer string, direction string) (core.AnswerCheck, error) {
	var result core.AnswerCheck
	err := client.do(http.MethodPost, "/study/cards/"+url.PathEscape(cardID)+"/check", map[string]string{"answer": answer, "direction": direction}, &result)
	return result, err
}

func (client *APIClient) Search(query string) (SearchResult, error) {
	var result SearchResult
	err := client.do(http.MethodGet, "/search?query="+url.QueryEscape(query), nil, &result)
	return result, err
}

func (client *APIClient) Execute(command string) (string, error) {
	parts := strings.SplitN(strings.TrimSpace(command), " ", 2)
	name := strings.ToLower(parts[0])
	argument := ""
	if len(parts) == 2 {
		argument = strings.TrimSpace(parts[1])
	}
	fields := splitFields(argument)

	switch name {
	case "deck-add":
		if len(fields) < 1 {
			return "", fmt.Errorf("deck-add NOM | DESCRIPTION")
		}
		return "Deck cree", client.do(http.MethodPost, "/api/decks", map[string]string{"name": fields[0], "description": field(fields, 1)}, nil)
	case "deck-update":
		if len(fields) < 2 {
			return "", fmt.Errorf("deck-update ID | NOM | DESCRIPTION")
		}
		return "Deck modifie", client.do(http.MethodPut, "/api/decks/"+url.PathEscape(fields[0]), map[string]string{"name": fields[1], "description": field(fields, 2)}, nil)
	case "deck-delete":
		if len(fields) != 1 {
			return "", fmt.Errorf("deck-delete ID")
		}
		return "Deck supprime", client.do(http.MethodDelete, "/api/decks/"+url.PathEscape(fields[0]), nil, nil)
	case "decks-delete":
		ids := commaIDs(argument)
		return "Decks supprimes", client.do(http.MethodDelete, "/api/decks/bulk", map[string]any{"ids": ids}, nil)
	case "decks-description":
		if len(fields) != 2 {
			return "", fmt.Errorf("decks-description ID1,ID2 | DESCRIPTION")
		}
		return "Decks modifies", client.do(http.MethodPut, "/api/decks/bulk", map[string]any{"ids": commaIDs(fields[0]), "patch": map[string]string{"description": fields[1]}}, nil)
	case "card-add":
		if len(fields) < 3 {
			return "", fmt.Errorf("card-add DECK_ID | COREEN | TRADUCTION | ROMANISATION")
		}
		payload := map[string]any{"deckId": fields[0], "kind": "vocabulary", "korean": fields[1], "translation": fields[2], "romanization": field(fields, 3), "tags": []string{}}
		return "Carte creee", client.do(http.MethodPost, "/api/cards", payload, nil)
	case "card-update":
		if len(fields) < 3 {
			return "", fmt.Errorf("card-update ID | COREEN | TRADUCTION | ROMANISATION")
		}
		payload := map[string]any{"korean": fields[1], "translation": fields[2], "romanization": field(fields, 3)}
		return "Carte modifiee", client.do(http.MethodPut, "/api/cards/"+url.PathEscape(fields[0]), payload, nil)
	case "card-delete":
		if len(fields) != 1 {
			return "", fmt.Errorf("card-delete ID")
		}
		return "Carte supprimee", client.do(http.MethodDelete, "/api/cards/"+url.PathEscape(fields[0]), nil, nil)
	case "cards-delete":
		ids := commaIDs(argument)
		return fmt.Sprintf("%d carte(s) demandee(s)", len(ids)), client.do(http.MethodDelete, "/api/cards/bulk", map[string]any{"ids": ids}, nil)
	case "cards-move":
		if len(fields) != 2 {
			return "", fmt.Errorf("cards-move ID1,ID2 | DECK_ID")
		}
		return "Cartes deplacees", client.do(http.MethodPut, "/api/cards/bulk", map[string]any{"ids": commaIDs(fields[0]), "patch": map[string]string{"deckId": fields[1]}}, nil)
	case "journal-add":
		if len(fields) < 2 {
			return "", fmt.Errorf("journal-add TITRE | TEXTE")
		}
		return "Entree corrigee et enregistree", client.do(http.MethodPost, "/api/journal", map[string]string{"title": fields[0], "text": fields[1]}, nil)
	case "journal-update":
		if len(fields) < 3 {
			return "", fmt.Errorf("journal-update ID | TITRE | TEXTE")
		}
		return "Entree modifiee", client.do(http.MethodPut, "/api/journal/"+url.PathEscape(fields[0]), map[string]string{"title": fields[1], "text": fields[2]}, nil)
	case "journal-delete":
		if len(fields) != 1 {
			return "", fmt.Errorf("journal-delete ID")
		}
		return "Entree supprimee", client.do(http.MethodDelete, "/api/journal/"+url.PathEscape(fields[0]), nil, nil)
	case "lesson-complete":
		if len(fields) < 1 {
			return "", fmt.Errorf("lesson-complete ID | SCORE")
		}
		score := 100
		if len(fields) > 1 {
			_, _ = fmt.Sscanf(fields[1], "%d", &score)
		}
		return "Lecon terminee", client.do(http.MethodPut, "/api/lessons/"+url.PathEscape(fields[0])+"/progress", map[string]any{"completed": true, "score": score}, nil)
	case "profile":
		if len(fields) < 2 {
			return "", fmt.Errorf("profile NOM | EMAIL")
		}
		return "Profil modifie", client.do(http.MethodPut, "/user/me", map[string]string{"name": fields[0], "email": fields[1]}, nil)
	case "admin-user":
		if len(fields) < 2 {
			return "", fmt.Errorf("admin-user ID | NOUVEAU_NOM")
		}
		return "Utilisateur modifie", client.do(http.MethodPut, "/admin/users/"+url.PathEscape(fields[0]), map[string]string{"name": fields[1]}, nil)
	case "reset":
		if argument != "CONFIRM" {
			return "", fmt.Errorf("utilise reset CONFIRM")
		}
		return "Base reinitialisee", client.do(http.MethodPost, "/admin/reset", map[string]any{}, nil)
	case "export":
		if argument == "" {
			argument = "korean-cards.csv"
		}
		content, err := client.getText("/api/cards/export")
		if err != nil {
			return "", err
		}
		return "Export ecrit dans " + argument, os.WriteFile(argument, []byte(content), 0o644)
	case "import":
		if len(fields) != 2 {
			return "", fmt.Errorf("import DECK_ID | CHEMIN_CSV")
		}
		content, err := os.ReadFile(fields[1])
		if err != nil {
			return "", err
		}
		return "CSV importe", client.do(http.MethodPost, "/api/cards/import", map[string]string{"deckId": fields[0], "csv": string(content)}, nil)
	default:
		return "", fmt.Errorf("commande inconnue: %s", name)
	}
}

func (client *APIClient) do(method string, path string, payload any, output any) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, client.BaseURL+path, body)
	if err != nil {
		return err
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if client.Token != "" {
		request.Header.Set("Authorization", "Bearer "+client.Token)
	}

	response, err := client.HTTP.Do(request)
	if err != nil {
		return fmt.Errorf("API inaccessible: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var apiError struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(response.Body).Decode(&apiError)
		if apiError.Error == "" {
			apiError.Error = response.Status
		}
		return fmt.Errorf("%s", apiError.Error)
	}
	if output == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(response.Body).Decode(output)
}

func (client *APIClient) getText(path string) (string, error) {
	request, _ := http.NewRequest(http.MethodGet, client.BaseURL+path, nil)
	request.Header.Set("Authorization", "Bearer "+client.Token)
	response, err := client.HTTP.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("export: %s", response.Status)
	}
	content, err := io.ReadAll(response.Body)
	return string(content), err
}

func splitFields(value string) []string {
	parts := strings.Split(value, "|")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	if len(result) == 1 && result[0] == "" {
		return nil
	}
	return result
}

func field(fields []string, index int) string {
	if index >= len(fields) {
		return ""
	}
	return fields[index]
}

func commaIDs(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if id := strings.TrimSpace(part); id != "" {
			result = append(result, id)
		}
	}
	return result
}

func tokenPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".korean-learning-go", "token")
}

func loadToken() string {
	if token := strings.TrimSpace(os.Getenv("KOREAN_TOKEN")); token != "" {
		return token
	}
	content, _ := os.ReadFile(tokenPath())
	return strings.TrimSpace(string(content))
}

func saveToken(token string) error {
	path := tokenPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(token), 0o600)
}
