package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func useTemporaryConfigDirectory(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	t.Setenv("APPDATA", directory)
	t.Setenv("XDG_CONFIG_HOME", directory)
	t.Setenv("HOME", directory)
	return filepath.Join(directory, "korean-learning-go")
}

func TestDefaultConfigUsesBuildAPIURL(t *testing.T) {
	previous := defaultAPIURL
	defaultAPIURL = "https://api.example.test"
	t.Cleanup(func() { defaultAPIURL = previous })

	if config := defaultConfig(); config.APIURL != defaultAPIURL {
		t.Fatalf("default config ignored build API URL: %#v", config)
	}
}

func TestConfigAndStateAreStoredAsJSON(t *testing.T) {
	directory := useTemporaryConfigDirectory(t)
	config := AppConfig{APIURL: "https://api.example.test/", Theme: "ocean"}
	state := AppState{ActiveView: "lessons", StudyDirection: "french-to-korean", LibraryCards: true}

	if err := saveConfig(config); err != nil {
		t.Fatal(err)
	}
	if err := saveState(state); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"config.json", "state.json"} {
		content, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if !json.Valid(content) {
			t.Fatalf("%s is not valid JSON: %s", name, content)
		}
	}

	loadedConfig, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	loadedState, err := loadState()
	if err != nil {
		t.Fatal(err)
	}
	if loadedConfig.APIURL != "https://api.example.test" || loadedConfig.Theme != "ocean" {
		t.Fatalf("unexpected config: %#v", loadedConfig)
	}
	if loadedState.ActiveView != "lessons" || loadedState.StudyDirection != "french-to-korean" || !loadedState.LibraryCards {
		t.Fatalf("unexpected state: %#v", loadedState)
	}
}

func TestSettingsApplyAndPersistTheme(t *testing.T) {
	useTemporaryConfigDirectory(t)
	m := model{config: defaultConfig(), activeUserID: "user-1", tab: tabSettings, cursor: themeIndex("rose"), studyDirection: "korean-to-french", libraryCards: true}

	updated, _ := m.updateNavigation(keyEnter())
	result := updated.(model)
	if result.config.Theme != "rose" || themeName(result.config.Theme) != "Rose" {
		t.Fatalf("theme was not applied: %#v", result.config)
	}
	stored, err := loadUserConfig("user-1", defaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if stored.Theme != "rose" {
		t.Fatalf("theme was not persisted: %#v", stored)
	}
	applyTheme("emerald")
}

func TestActivatingAccountRestoresItsOwnSettings(t *testing.T) {
	useTemporaryConfigDirectory(t)
	if err := saveUserConfig("user-1", AppConfig{APIURL: "https://api.example.test", Theme: "rose"}); err != nil {
		t.Fatal(err)
	}
	if err := saveUserState("user-1", AppState{ActiveView: "lessons", StudyDirection: "french-to-korean"}); err != nil {
		t.Fatal(err)
	}
	if err := saveUserConfig("user-2", AppConfig{APIURL: "https://api.example.test", Theme: "ocean"}); err != nil {
		t.Fatal(err)
	}
	if err := saveUserState("user-2", AppState{ActiveView: "journal", StudyDirection: "korean-to-french"}); err != nil {
		t.Fatal(err)
	}

	m := model{client: &APIClient{BaseURL: "https://api.example.test"}, config: defaultConfig()}
	m.activateUserProfile("user-1")
	if m.activeUserID != "user-1" || m.config.Theme != "rose" || m.tab != tabLessons {
		t.Fatalf("first account was not restored: %#v", m)
	}
	m.activateUserProfile("user-2")
	if m.activeUserID != "user-2" || m.config.Theme != "ocean" || m.tab != tabJournal {
		t.Fatalf("second account was not restored: %#v", m)
	}
	applyTheme("emerald")
}

func TestDownloadedBackupRestoresLocalFilesAndModel(t *testing.T) {
	useTemporaryConfigDirectory(t)
	m := model{
		client:         &APIClient{BaseURL: "http://old.example.test"},
		config:         defaultConfig(),
		activeUserID:   "user-1",
		loggedIn:       true,
		libraryCards:   true,
		studyDirection: "korean-to-french",
	}

	updated, _ := m.Update(backupMsg{action: "download", backup: RemoteBackup{
		Config: AppConfig{APIURL: "https://new.example.test", Theme: "amber"},
		State:  AppState{ActiveView: "journal", StudyDirection: "french-to-korean", LibraryCards: false},
	}})
	result := updated.(model)
	if result.config.Theme != "amber" || result.client.BaseURL != "https://new.example.test" || result.tab != tabJournal || result.studyDirection != "french-to-korean" || result.libraryCards {
		t.Fatalf("backup was not restored: %#v", result)
	}
	if _, err := os.Stat(userConfigFilePath("user-1")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(userStateFilePath("user-1")); err != nil {
		t.Fatal(err)
	}
	applyTheme("emerald")
}

func TestLocalSettingsAndStateAreIsolatedByUser(t *testing.T) {
	useTemporaryConfigDirectory(t)
	firstConfig := AppConfig{APIURL: "https://api.example.test", Theme: "rose"}
	secondConfig := AppConfig{APIURL: "https://api.example.test", Theme: "ocean"}
	firstState := AppState{ActiveView: "lessons", StudyDirection: "french-to-korean", LibraryCards: false}
	secondState := AppState{ActiveView: "journal", StudyDirection: "korean-to-french", LibraryCards: true}

	if err := saveUserConfig("user-1", firstConfig); err != nil {
		t.Fatal(err)
	}
	if err := saveUserState("user-1", firstState); err != nil {
		t.Fatal(err)
	}
	if err := saveUserConfig("user-2", secondConfig); err != nil {
		t.Fatal(err)
	}
	if err := saveUserState("user-2", secondState); err != nil {
		t.Fatal(err)
	}

	loadedFirstConfig, err := loadUserConfig("user-1", defaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	loadedFirstState, err := loadUserState("user-1", defaultState())
	if err != nil {
		t.Fatal(err)
	}
	loadedSecondConfig, err := loadUserConfig("user-2", defaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	loadedSecondState, err := loadUserState("user-2", defaultState())
	if err != nil {
		t.Fatal(err)
	}

	if loadedFirstConfig.Theme != "rose" || loadedFirstState.ActiveView != "lessons" {
		t.Fatalf("unexpected first user settings: %#v %#v", loadedFirstConfig, loadedFirstState)
	}
	if loadedSecondConfig.Theme != "ocean" || loadedSecondState.ActiveView != "journal" {
		t.Fatalf("unexpected second user settings: %#v %#v", loadedSecondConfig, loadedSecondState)
	}
	if userConfigFilePath("user-1") == userConfigFilePath("user-2") {
		t.Fatal("different users share the same config path")
	}
}
