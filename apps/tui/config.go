package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const localDataVersion = 1

type AppConfig struct {
	Version int    `json:"version"`
	APIURL  string `json:"apiUrl"`
	Theme   string `json:"theme"`
}

type AppState struct {
	Version        int       `json:"version"`
	ActiveView     string    `json:"activeView"`
	StudyDirection string    `json:"studyDirection"`
	LibraryCards   bool      `json:"libraryCards"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type RemoteBackup struct {
	Config    AppConfig `json:"config"`
	State     AppState  `json:"state"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func defaultConfig() AppConfig {
	return AppConfig{Version: localDataVersion, APIURL: defaultAPIURL, Theme: "emerald"}
}

func defaultState() AppState {
	return AppState{
		Version:        localDataVersion,
		ActiveView:     "home",
		StudyDirection: "korean-to-french",
		LibraryCards:   true,
		UpdatedAt:      time.Now().UTC(),
	}
}

func loadConfig() (AppConfig, error) {
	config := defaultConfig()
	err := readJSONFile(configFilePath(), &config)
	if os.IsNotExist(err) {
		return config, nil
	}
	if err != nil {
		return defaultConfig(), fmt.Errorf("lire config.json: %w", err)
	}
	return normalizeConfig(config)
}

func saveConfig(config AppConfig) error {
	config, err := normalizeConfig(config)
	if err != nil {
		return err
	}
	return writeJSONFile(configFilePath(), config)
}

func loadState() (AppState, error) {
	state := defaultState()
	err := readJSONFile(stateFilePath(), &state)
	if os.IsNotExist(err) {
		return state, nil
	}
	if err != nil {
		return defaultState(), fmt.Errorf("lire state.json: %w", err)
	}
	return normalizeState(state), nil
}

func saveState(state AppState) error {
	state = normalizeState(state)
	state.UpdatedAt = time.Now().UTC()
	return writeJSONFile(stateFilePath(), state)
}

func loadUserConfig(userID string, fallback AppConfig) (AppConfig, error) {
	if strings.TrimSpace(userID) == "" {
		return AppConfig{}, fmt.Errorf("identifiant utilisateur manquant")
	}
	config := fallback
	err := readJSONFile(userConfigFilePath(userID), &config)
	if os.IsNotExist(err) {
		return normalizeConfig(config)
	}
	if err != nil {
		return AppConfig{}, fmt.Errorf("lire le config.json du compte: %w", err)
	}
	return normalizeConfig(config)
}

func saveUserConfig(userID string, config AppConfig) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("identifiant utilisateur manquant")
	}
	config, err := normalizeConfig(config)
	if err != nil {
		return err
	}
	return writeJSONFile(userConfigFilePath(userID), config)
}

func loadUserState(userID string, fallback AppState) (AppState, error) {
	if strings.TrimSpace(userID) == "" {
		return AppState{}, fmt.Errorf("identifiant utilisateur manquant")
	}
	state := fallback
	err := readJSONFile(userStateFilePath(userID), &state)
	if os.IsNotExist(err) {
		return normalizeState(state), nil
	}
	if err != nil {
		return AppState{}, fmt.Errorf("lire le state.json du compte: %w", err)
	}
	return normalizeState(state), nil
}

func saveUserState(userID string, state AppState) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("identifiant utilisateur manquant")
	}
	state = normalizeState(state)
	state.UpdatedAt = time.Now().UTC()
	return writeJSONFile(userStateFilePath(userID), state)
}

func normalizeConfig(config AppConfig) (AppConfig, error) {
	apiURL, err := normalizeAPIURL(config.APIURL)
	if err != nil {
		return AppConfig{}, err
	}
	if !isThemeSupported(config.Theme) {
		config.Theme = defaultConfig().Theme
	}
	config.Version = localDataVersion
	config.APIURL = apiURL
	return config, nil
}

func normalizeState(state AppState) AppState {
	if !isViewSupported(state.ActiveView) {
		state.ActiveView = defaultState().ActiveView
	}
	if state.StudyDirection != "korean-to-french" && state.StudyDirection != "french-to-korean" {
		state.StudyDirection = defaultState().StudyDirection
	}
	state.Version = localDataVersion
	return state
}

func normalizeAPIURL(value string) (string, error) {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", fmt.Errorf("l'URL de l'API doit commencer par http:// ou https://")
	}
	return value, nil
}

func configDirectory() string {
	directory, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(directory) == "" {
		home, _ := os.UserHomeDir()
		directory = filepath.Join(home, ".config")
	}
	return filepath.Join(directory, "korean-learning-go")
}

func configFilePath() string {
	return filepath.Join(configDirectory(), "config.json")
}

func stateFilePath() string {
	return filepath.Join(configDirectory(), "state.json")
}

func userConfigFilePath(userID string) string {
	return filepath.Join(userConfigDirectory(userID), "config.json")
}

func userStateFilePath(userID string) string {
	return filepath.Join(userConfigDirectory(userID), "state.json")
}

func userConfigDirectory(userID string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(userID)))
	return filepath.Join(configDirectory(), "users", fmt.Sprintf("%x", digest[:12]))
}

func hasUserProfiles() bool {
	entries, err := os.ReadDir(filepath.Join(configDirectory(), "users"))
	return err == nil && len(entries) > 0
}

func readJSONFile(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, target); err != nil {
		return err
	}
	return nil
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o600)
}
