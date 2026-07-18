package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/packages/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	green, brightGreen, red, muted, panel, border, text lipgloss.Color
	baseStyle, activeStyle, mutedStyle, redStyle        lipgloss.Style
	selectedStyle, panelStyle, errorStyle               lipgloss.Style
)

type themeDefinition struct {
	ID          string
	Name        string
	Description string
	Accent      string
	Bright      string
	Danger      string
	Muted       string
	Panel       string
	Border      string
	Text        string
	Selection   string
}

var themes = []themeDefinition{
	{ID: "emerald", Name: "Émeraude", Description: "Vert profond et accents lumineux", Accent: "#4EA67A", Bright: "#78D7A6", Danger: "#E05A47", Muted: "#7E8B86", Panel: "#17211E", Border: "#395048", Text: "#F1F5F3", Selection: "#08110E"},
	{ID: "ocean", Name: "Océan", Description: "Bleu calme avec contraste cyan", Accent: "#4E91B8", Bright: "#82C7EA", Danger: "#F07167", Muted: "#82939E", Panel: "#121D27", Border: "#35536A", Text: "#F2F7FA", Selection: "#07131B"},
	{ID: "amber", Name: "Ambre", Description: "Or chaud sur fond anthracite", Accent: "#D1A447", Bright: "#F0CD73", Danger: "#E66A4E", Muted: "#948B78", Panel: "#211D16", Border: "#5B4E32", Text: "#F8F3E7", Selection: "#171006"},
	{ID: "rose", Name: "Rose", Description: "Framboise douce et gris fumé", Accent: "#B85D7A", Bright: "#E88EAB", Danger: "#F06B5D", Muted: "#92858B", Panel: "#21191D", Border: "#60404B", Text: "#FAF2F5", Selection: "#190B10"},
}

const (
	tabHome = iota
	tabStudy
	tabLibrary
	tabLessons
	tabJournal
	tabStats
	tabSearch
	tabSettings
	tabProfile
	tabAdmin
)

var tabs = []string{"ACCUEIL", "RÉVISION", "BIBLIOTHÈQUE", "LEÇONS", "JOURNAL", "PROGRESSION", "RECHERCHE", "PARAMÈTRES", "PROFIL"}
var tabIDs = []string{"home", "study", "library", "lessons", "journal", "stats", "search", "settings", "profile"}

type model struct {
	client         *APIClient
	config         AppConfig
	activeUserID   string
	data           DashboardData
	search         SearchResult
	width          int
	height         int
	tab            int
	cursor         int
	detailScroll   int
	loading        bool
	loggedIn       bool
	revealed       bool
	showHelp       bool
	libraryCards   bool
	inputMode      string
	input          string
	loginEmail     string
	loginPass      string
	loginName      string
	loginField     int
	registering    bool
	profileEditing bool
	profileField   int
	profileName    string
	profileEmail   string
	profilePass    string
	adminEditing   bool
	adminField     int
	adminUserID    string
	adminName      string
	adminEmail     string
	adminPass      string
	journalEditing bool
	journalField   int
	journalID      string
	journalTitle   string
	journalText    string
	studyDirection string
	lastBackupAt   time.Time
	status         string
	err            error
}

type dashboardMsg struct {
	data DashboardData
	err  error
}
type loginMsg struct {
	result AuthResult
	err    error
}
type mutationMsg struct {
	status string
	err    error
}
type adminUserMsg struct {
	err error
}
type journalSaveMsg struct {
	created bool
	err     error
}
type searchMsg struct {
	result SearchResult
	err    error
}
type checkMsg struct {
	result core.AnswerCheck
	err    error
}
type backupMsg struct {
	action string
	backup RemoteBackup
	err    error
}

func init() {
	applyTheme("emerald")
}

var (
	version       = "dev"
	defaultAPIURL = localDevelopmentAPIURL
)

func main() {
	apiURL := flag.String("api", "", "URL du backend (prioritaire sur config.json)")
	showVersion := flag.Bool("version", false, "Affiche la version et quitte")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	configValue, configErr := loadConfig()
	stateValue, stateErr := loadState()
	configValue = migrateReleaseAPIURL(configValue)
	if envURL := strings.TrimSpace(os.Getenv("KOREAN_API_URL")); envURL != "" {
		configValue.APIURL = envURL
	}
	if strings.TrimSpace(*apiURL) != "" {
		configValue.APIURL = *apiURL
	}
	if normalized, err := normalizeConfig(configValue); err == nil {
		configValue = normalized
	} else {
		configErr = err
		configValue = defaultConfig()
	}
	_ = saveConfig(configValue)
	_ = saveState(stateValue)
	applyTheme(configValue.Theme)

	client := NewAPIClient(configValue.APIURL)
	application := model{
		client:         client,
		config:         configValue,
		loggedIn:       client.Token != "",
		loading:        client.Token != "",
		libraryCards:   stateValue.LibraryCards,
		loginEmail:     "admin@korean.local",
		studyDirection: stateValue.StudyDirection,
		tab:            tabIndexForID(stateValue.ActiveView),
	}
	application.resetCursorForTab()
	if configErr != nil {
		application.status = configErr.Error()
	} else if stateErr != nil {
		application.status = stateErr.Error()
	}
	if _, err := tea.NewProgram(application, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	if m.loggedIn {
		return loadDashboard(m.client)
	}
	return nil
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case dashboardMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.data = msg.data
			if m.activeUserID != msg.data.User.ID {
				m.activateUserProfile(msg.data.User.ID)
			}
			m.status = "Données synchronisées"
			if m.tab == tabAdmin && !m.data.User.IsAdmin {
				m.tab = tabHome
			}
			m.cursor = clamp(m.cursor, 0, max(0, m.itemCount()-1))
			m.persistState()
		} else if strings.Contains(strings.ToLower(msg.err.Error()), "token") || strings.Contains(strings.ToLower(msg.err.Error()), "unauthorized") {
			m.persistState()
			m.loggedIn = false
			m.activeUserID = ""
			m.client.Token = ""
			_ = os.Remove(tokenPath())
		}
		return m, nil
	case loginMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.loggedIn = true
			m.activateUserProfile(msg.result.User.ID)
			m.status = "Bienvenue " + msg.result.User.Name
			m.loading = true
			return m, loadDashboard(m.client)
		}
		return m, nil
	case mutationMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.status = msg.status
			m.revealed = false
			return m, loadDashboard(m.client)
		}
		return m, nil
	case adminUserMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.adminEditing = false
			m.adminPass = ""
			m.status = "Utilisateur mis à jour"
			return m, loadDashboard(m.client)
		}
		return m, nil
	case journalSaveMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.journalEditing = false
			m.journalID = ""
			m.journalTitle = ""
			m.journalText = ""
			m.status = "Entrée du journal modifiée"
			if msg.created {
				m.status = "Entrée du journal créée et corrigée"
			}
			return m, loadDashboard(m.client)
		}
		return m, nil
	case searchMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.search = msg.result
			m.status = fmt.Sprintf("%d résultat(s)", len(msg.result.Cards)+len(msg.result.Decks))
		}
		return m, nil
	case checkMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.revealed = true
			if msg.result.Correct {
				m.status = "Bonne réponse"
			} else {
				m.status = "Réponse attendue : " + msg.result.Expected
			}
		}
		return m, nil
	case backupMsg:
		m.loading = false
		m.err = msg.err
		if msg.err != nil {
			return m, nil
		}
		m.lastBackupAt = msg.backup.UpdatedAt
		if msg.action == "upload" {
			m.status = "Configuration et état sauvegardés sur le serveur"
			return m, nil
		}

		configValue, err := normalizeConfig(msg.backup.Config)
		if err != nil {
			m.err = err
			return m, nil
		}
		stateValue := normalizeState(msg.backup.State)
		if err := saveUserConfig(m.activeUserID, configValue); err != nil {
			m.err = err
			return m, nil
		}
		if err := saveUserState(m.activeUserID, stateValue); err != nil {
			m.err = err
			return m, nil
		}
		_ = saveConfig(configValue)
		m.config = configValue
		m.client.BaseURL = configValue.APIURL
		m.libraryCards = stateValue.LibraryCards
		m.studyDirection = stateValue.StudyDirection
		m.tab = tabIndexForID(stateValue.ActiveView)
		if m.tab == tabAdmin && !m.data.User.IsAdmin {
			m.tab = tabHome
		}
		m.resetCursorForTab()
		m.detailScroll = 0
		applyTheme(configValue.Theme)
		m.status = "Backup restauré depuis le serveur"
		return m, nil
	case tea.KeyMsg:
		if !m.loggedIn {
			return m.updateLogin(msg)
		}
		if m.profileEditing {
			return m.updateProfile(msg)
		}
		if m.adminEditing {
			return m.updateAdminUser(msg)
		}
		if m.journalEditing {
			return m.updateJournalEditor(msg)
		}
		if m.inputMode != "" {
			return m.updateInput(msg)
		}
		return m.updateNavigation(msg)
	}
	return m, nil
}

func (m model) updateJournalEditor(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEsc:
		m.journalEditing = false
		m.journalID = ""
		m.journalTitle = ""
		m.journalText = ""
		m.err = nil
	case tea.KeyTab, tea.KeyDown, tea.KeyUp:
		m.journalField = (m.journalField + 1) % 2
	case tea.KeyEnter:
		if m.journalField == 0 {
			m.journalField = 1
			return m, nil
		}
		if strings.TrimSpace(m.journalText) == "" {
			m.err = fmt.Errorf("le texte du journal est obligatoire")
			return m, nil
		}
		m.loading, m.err = true, nil
		return m, saveJournalCommand(m.client, m.journalID, m.journalTitle, m.journalText)
	case tea.KeyBackspace:
		m.err = nil
		if m.journalField == 0 {
			m.journalTitle = trimLastRune(m.journalTitle)
		} else {
			m.journalText = trimLastRune(m.journalText)
		}
	case tea.KeyRunes:
		m.err = nil
		m.appendJournalInput(string(key.Runes))
	case tea.KeySpace:
		m.err = nil
		m.appendJournalInput(" ")
	}
	return m, nil
}

func (m *model) appendJournalInput(value string) {
	if m.journalField == 0 {
		m.journalTitle += value
	} else {
		m.journalText += value
	}
}

func (m model) updateAdminUser(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEsc:
		m.adminEditing = false
		m.adminPass = ""
		m.err = nil
	case tea.KeyTab, tea.KeyDown:
		m.adminField = (m.adminField + 1) % 3
	case tea.KeyUp:
		m.adminField = (m.adminField + 2) % 3
	case tea.KeyEnter:
		if m.adminField < 2 {
			m.adminField++
			return m, nil
		}
		name, email := cleanNameInput(m.adminName), cleanEmailInput(m.adminEmail)
		if len([]rune(name)) < 2 {
			m.err = fmt.Errorf("le nom doit contenir au moins 2 caractères")
			return m, nil
		}
		if !validEmailInput(email) {
			m.err = fmt.Errorf("l'adresse email n'est pas valide")
			return m, nil
		}
		if m.adminPass != "" && len([]rune(m.adminPass)) < 8 {
			m.err = fmt.Errorf("le mot de passe doit contenir au moins 8 caractères")
			return m, nil
		}
		m.loading, m.err = true, nil
		return m, adminUpdateUserCommand(m.client, m.adminUserID, name, email, m.adminPass)
	case tea.KeyBackspace:
		m.err = nil
		switch m.adminField {
		case 0:
			m.adminName = trimLastRune(m.adminName)
		case 1:
			m.adminEmail = trimLastRune(m.adminEmail)
		case 2:
			m.adminPass = trimLastRune(m.adminPass)
		}
	case tea.KeyRunes:
		m.err = nil
		m.appendAdminInput(string(key.Runes))
	case tea.KeySpace:
		m.err = nil
		m.appendAdminInput(" ")
	}
	return m, nil
}

func (m *model) appendAdminInput(value string) {
	switch m.adminField {
	case 0:
		m.adminName += value
	case 1:
		m.adminEmail += value
	case 2:
		m.adminPass += value
	}
}

func (m model) updateProfile(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEsc:
		m.profileEditing = false
		m.profilePass = ""
	case tea.KeyTab, tea.KeyDown:
		m.profileField = (m.profileField + 1) % 3
	case tea.KeyUp:
		m.profileField = (m.profileField + 2) % 3
	case tea.KeyEnter:
		if m.profileField < 2 {
			m.profileField++
			return m, nil
		}
		if strings.TrimSpace(m.profileName) == "" || strings.TrimSpace(m.profileEmail) == "" {
			m.err = fmt.Errorf("le nom et l'email sont obligatoires")
			return m, nil
		}
		if m.profilePass != "" && len([]rune(m.profilePass)) < 8 {
			m.err = fmt.Errorf("le mot de passe doit contenir au moins 8 caractères")
			return m, nil
		}
		name, email, password := m.profileName, m.profileEmail, m.profilePass
		m.profileEditing = false
		m.profilePass = ""
		m.loading, m.err = true, nil
		return m, updateProfileCommand(m.client, name, email, password)
	case tea.KeyBackspace:
		switch m.profileField {
		case 0:
			m.profileName = trimLastRune(m.profileName)
		case 1:
			m.profileEmail = trimLastRune(m.profileEmail)
		case 2:
			m.profilePass = trimLastRune(m.profilePass)
		}
	case tea.KeyRunes:
		m.appendProfileInput(string(key.Runes))
	case tea.KeySpace:
		m.appendProfileInput(" ")
	}
	return m, nil
}

func (m *model) appendProfileInput(value string) {
	switch m.profileField {
	case 0:
		m.profileName += value
	case 1:
		m.profileEmail += value
	case 2:
		m.profilePass += value
	}
}

func (m model) updateLogin(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.String() == "ctrl+r" {
		m.registering = !m.registering
		m.loginField = 0
		m.loginName = ""
		m.loginEmail = ""
		m.loginPass = ""
		m.err = nil
		return m, nil
	}

	fieldCount := 2
	if m.registering {
		fieldCount = 3
	}
	switch key.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit
	case tea.KeyTab, tea.KeyDown, tea.KeyUp:
		m.loginField = (m.loginField + 1) % fieldCount
	case tea.KeyEnter:
		if m.loginField < fieldCount-1 {
			m.loginField++
			return m, nil
		}
		m.loading, m.err = true, nil
		if m.registering {
			return m, registerCommand(m.client, m.loginName, m.loginEmail, m.loginPass)
		}
		return m, loginCommand(m.client, m.loginEmail, m.loginPass)
	case tea.KeyBackspace:
		m.err = nil
		if m.registering && m.loginField == 0 {
			m.loginName = trimLastRune(m.loginName)
		} else if (m.registering && m.loginField == 1) || (!m.registering && m.loginField == 0) {
			m.loginEmail = trimLastRune(m.loginEmail)
		} else {
			m.loginPass = trimLastRune(m.loginPass)
		}
	case tea.KeyRunes:
		m.err = nil
		if m.registering && m.loginField == 0 {
			m.loginName += string(key.Runes)
		} else if (m.registering && m.loginField == 1) || (!m.registering && m.loginField == 0) {
			m.loginEmail += string(key.Runes)
		} else {
			m.loginPass += string(key.Runes)
		}
	}
	return m, nil
}

func (m model) updateInput(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEsc:
		m.inputMode, m.input = "", ""
	case tea.KeyBackspace:
		m.input = trimLastRune(m.input)
	case tea.KeyEnter:
		value, mode := strings.TrimSpace(m.input), m.inputMode
		m.inputMode, m.input = "", ""
		if value == "" {
			return m, nil
		}
		if mode == "api-url" {
			apiURL, err := normalizeAPIURL(value)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.config.APIURL = apiURL
			if err := saveUserConfig(m.activeUserID, m.config); err != nil {
				m.err = err
				return m, nil
			}
			_ = saveConfig(m.config)
			m.client.BaseURL = apiURL
			m.status = "URL de l'API enregistrée"
			m.err = nil
			return m, nil
		}
		m.loading, m.err = true, nil
		if mode == "answer" && len(m.data.Due) > 0 {
			return m, checkCommand(m.client, m.data.Due[m.cursor].ID, value, m.studyDirection)
		}
		if mode == "search" {
			m.tab = tabSearch
			m.cursor = 0
			return m, searchCommand(m.client, value)
		}
		return m, executeCommand(m.client, value)
	case tea.KeyRunes:
		m.input += string(key.Runes)
	case tea.KeySpace:
		m.input += " "
	}
	return m, nil
}

func (m model) updateNavigation(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		m.persistState()
		return m, tea.Quit
	case "D":
		return m.logout(), nil
	case "?":
		m.showHelp = !m.showHelp
	case ":":
		m.inputMode = "command"
		m.status = commandHint(m.tab)
	case "/":
		m.inputMode = "search"
		m.input = ""
	case "h", "left":
		m.tab = (m.tab - 1 + len(m.visibleTabs())) % len(m.visibleTabs())
		m.resetCursorForTab()
		m.detailScroll = 0
		m.revealed = false
	case "l", "right":
		m.tab = (m.tab + 1) % len(m.visibleTabs())
		m.resetCursorForTab()
		m.detailScroll = 0
		m.revealed = false
	case "j", "down":
		m.cursor = clamp(m.cursor+1, 0, max(0, m.itemCount()-1))
		m.revealed = false
		m.detailScroll = 0
	case "k", "up":
		m.cursor = clamp(m.cursor-1, 0, max(0, m.itemCount()-1))
		m.revealed = false
		m.detailScroll = 0
	case "pgdown", "ctrl+d":
		m.detailScroll += max(3, m.height/3)
	case "pgup", "ctrl+u":
		m.detailScroll = max(0, m.detailScroll-max(3, m.height/3))
	case "r":
		m.loading = true
		return m, loadDashboard(m.client)
	case "tab":
		if m.tab == tabLibrary {
			m.libraryCards = !m.libraryCards
			m.cursor = 0
			m.detailScroll = 0
		}
	case " ":
		if m.tab == tabStudy {
			m.revealed = true
		}
	case "a":
		if m.tab == tabStudy && len(m.data.Due) > 0 {
			m.inputMode = "answer"
			m.input = ""
			m.status = "Écris ta réponse"
		}
	case "v":
		if m.tab == tabStudy {
			if m.studyDirection == "korean-to-french" {
				m.studyDirection = "french-to-korean"
			} else {
				m.studyDirection = "korean-to-french"
			}
			m.revealed = false
		}
	case "1", "2", "3", "4":
		if m.tab == tabStudy && m.revealed && len(m.data.Due) > 0 {
			ratings := map[string]string{"1": "again", "2": "hard", "3": "good", "4": "easy"}
			m.loading = true
			return m, answerCommand(m.client, m.data.Due[m.cursor].ID, ratings[key.String()])
		}
	case "enter":
		if m.tab == tabHome && len(homeOptions) > 0 {
			m.tab = homeOptions[clamp(m.cursor, 0, len(homeOptions)-1)].Tab
			m.resetCursorForTab()
			m.persistState()
			return m, nil
		}
		if m.tab == tabSettings && len(themes) > 0 {
			theme := themes[clamp(m.cursor, 0, len(themes)-1)]
			m.config.Theme = applyTheme(theme.ID)
			if err := saveUserConfig(m.activeUserID, m.config); err != nil {
				m.err = err
			} else {
				m.status = "Thème " + theme.Name + " enregistré"
				m.err = nil
			}
		}
		if m.tab == tabLessons && len(m.data.Lessons) > 0 && !m.data.Lessons[m.cursor].Progress.Completed {
			m.loading = true
			return m, executeCommand(m.client, "lesson-complete "+m.data.Lessons[m.cursor].ID)
		}
	case "n":
		switch m.tab {
		case tabLibrary:
			m.inputMode = "command"
			if m.libraryCards {
				m.input = "card-add "
			} else {
				m.input = "deck-add "
			}
		case tabJournal:
			m.journalEditing = true
			m.journalField = 0
			m.journalID = ""
			m.journalTitle = ""
			m.journalText = ""
			m.err = nil
		}
	case "e":
		if m.tab == tabSettings {
			m.inputMode = "api-url"
			m.input = m.config.APIURL
			m.status = "Nouvelle URL de l'API"
		} else if m.tab == tabProfile {
			m.profileEditing = true
			m.profileField = 0
			m.profileName = m.data.User.Name
			m.profileEmail = m.data.User.Email
			m.profilePass = ""
			m.err = nil
		} else if m.tab == tabAdmin && len(m.data.Users) > 0 {
			user := m.data.Users[m.cursor]
			m.adminEditing = true
			m.adminField = 0
			m.adminUserID = user.ID
			m.adminName = user.Name
			m.adminEmail = user.Email
			m.adminPass = ""
			m.err = nil
		} else if m.tab == tabJournal && len(m.data.Journal) > 0 {
			entry := m.data.Journal[m.cursor]
			m.journalEditing = true
			m.journalField = 0
			m.journalID = entry.ID
			m.journalTitle = entry.Title
			m.journalText = entry.OriginalText
			m.err = nil
		}
	case "x":
		if m.tab == tabAdmin && m.data.User.IsAdmin {
			m.inputMode = "command"
			m.input = "reset CONFIRM"
		}
	case "i":
		if m.tab == tabAdmin && m.data.User.IsAdmin && m.data.RAG.Enabled {
			m.loading = true
			return m, executeCommand(m.client, "rag-reindex")
		}
	case "d":
		if command := m.deleteCommand(); command != "" {
			m.loading = true
			return m, executeCommand(m.client, command)
		}
	case "u":
		if m.tab == tabSettings {
			state := m.appState()
			if err := saveUserState(m.activeUserID, state); err != nil {
				m.err = err
				break
			}
			m.loading = true
			m.err = nil
			return m, uploadBackupCommand(m.client, m.config, state)
		}
	case "o":
		if m.tab == tabSettings {
			m.loading = true
			m.err = nil
			return m, downloadBackupCommand(m.client)
		}
	}
	m.persistState()
	return m, nil
}

func (m model) logout() model {
	m.persistState()
	m.client.Token = ""
	_ = os.Remove(tokenPath())
	m.loggedIn = false
	m.activeUserID = ""
	m.data = DashboardData{}
	m.search = SearchResult{}
	m.tab = tabHome
	m.cursor = 0
	m.loading = false
	m.profileEditing = false
	m.journalEditing = false
	m.adminEditing = false
	m.inputMode = ""
	m.input = ""
	m.registering = false
	m.loginField = 0
	m.loginName = ""
	m.loginEmail = ""
	m.loginPass = ""
	m.status = "Déconnecté"
	m.err = nil
	return m
}

func (m model) View() string {
	if !m.loggedIn {
		return m.loginView()
	}
	width := max(80, m.width)
	height := max(24, m.height)

	header := m.headerView(width)
	footer := m.footerView(width)
	bodyHeight := max(10, height-lipgloss.Height(header)-lipgloss.Height(footer)-1)
	body := m.bodyView(width, bodyHeight)
	if m.showHelp {
		body = m.helpView(width, bodyHeight)
	}
	return baseStyle.Render(header + "\n" + body + "\n" + footer)
}

func (m model) loginView() string {
	field := func(label, value string, active bool) string {
		style := panelStyle.Width(48)
		if active {
			style = style.BorderForeground(brightGreen)
		}
		return style.Render(mutedStyle.Render(label) + "\n" + value)
	}
	password := strings.Repeat("•", len([]rune(m.loginPass)))
	modeLabel := "Connexion au backend Gin / SQLite"
	fields := ""
	if m.registering {
		modeLabel = "Création d'un compte personnel"
		fields += field("NOM", m.loginName, m.loginField == 0) + "\n"
	}
	emailField, passwordField := 0, 1
	if m.registering {
		emailField, passwordField = 1, 2
	}
	fields += field("EMAIL", m.loginEmail, m.loginField == emailField) + "\n" + field("MOT DE PASSE", password, m.loginField == passwordField)
	content := activeStyle.Render("한  KOREAN LEARNING") + "\n\n" + mutedStyle.Render(modeLabel) + "\n\n" + fields + "\n\n" + mutedStyle.Render("tab changer  ·  entrée valider  ·  ctrl+r connexion/inscription  ·  esc quitter")
	if m.loading {
		content += "\n" + activeStyle.Render("Connexion...")
	}
	if m.err != nil {
		content += "\n" + errorStyle.Render(m.err.Error())
	}
	return lipgloss.Place(max(70, m.width), max(22, m.height), lipgloss.Center, lipgloss.Center, content)
}

func (m model) headerView(width int) string {
	brand := activeStyle.Render("한 KOREAN LEARNING") + "  " + mutedStyle.Render(m.data.User.Name)
	visibleTabs := m.visibleTabs()
	items := make([]string, 0, len(visibleTabs))
	for index, tab := range visibleTabs {
		if index == m.tab {
			items = append(items, selectedStyle.Padding(0, 1).Render(tab))
		} else {
			items = append(items, mutedStyle.Padding(0, 1).Render(tab))
		}
	}
	return lipgloss.NewStyle().Width(width).Padding(0, 1).Render(brand + "\n" + wrapHeaderItems(items, max(20, width-4)))
}

func (m model) bodyView(width int, height int) string {
	if m.loading && len(m.data.Decks) == 0 {
		return panelStyle.Width(width - 6).Height(height - 2).Render("Synchronisation avec l'API...")
	}
	switch m.tab {
	case tabHome:
		return m.homeView(width, height)
	case tabStudy:
		return m.studyView(width, height)
	case tabLibrary:
		return m.libraryView(width, height)
	case tabLessons:
		return m.lessonsView(width, height)
	case tabJournal:
		return m.journalView(width, height)
	case tabStats:
		return m.statsView(width, height)
	case tabSearch:
		return m.searchView(width, height)
	case tabSettings:
		return m.settingsView(width, height)
	case tabProfile:
		return m.profileView(width, height)
	default:
		return m.adminView(width, height)
	}
}

type homeOption struct {
	Title       string
	Description string
	Tab         int
}

var homeOptions = []homeOption{
	{Title: "Réviser maintenant", Description: "Travailler les cartes dues", Tab: tabStudy},
	{Title: "Gérer ma bibliothèque", Description: "Cartes, decks et import CSV", Tab: tabLibrary},
	{Title: "Continuer les leçons", Description: "Parcours guidé de coréen", Tab: tabLessons},
	{Title: "Écrire dans le journal", Description: "Pratiquer avec des corrections", Tab: tabJournal},
	{Title: "Voir ma progression", Description: "Statistiques et cartes difficiles", Tab: tabStats},
	{Title: "Configurer l'application", Description: "Thème, API et backup", Tab: tabSettings},
}

func (m model) homeView(width, height int) string {
	leftWidth := max(38, width/2)
	menu := "QUE VEUX-TU FAIRE ?\n\n"
	for index, option := range homeOptions {
		menu += listLine(index == m.cursor, option.Title, option.Description) + "\n"
	}
	menu += "\n" + mutedStyle.Render("j/k choisir · entrée ouvrir")

	summary := "BONJOUR " + strings.ToUpper(m.data.User.Name) + "\n\n"
	summary += fmt.Sprintf("Cartes dues aujourd'hui    %d\n", m.data.Stats.DueCards)
	summary += fmt.Sprintf("Cartes maîtrisées         %d\n", m.data.Stats.MasteredCards)
	summary += fmt.Sprintf("Série actuelle            %d jours\n\n", m.data.Stats.CurrentStreak)
	summary += activeStyle.Render("Objectif du jour") + "\n"
	if m.data.Stats.DueCards > 0 {
		summary += "Révise quelques cartes, puis avance dans une leçon."
	} else {
		summary += "La file est terminée : profite-en pour apprendre ou écrire."
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		panelStyle.Width(leftWidth).Height(height-2).Render(menu),
		panelStyle.Width(width-leftWidth-9).Height(height-2).Render(summary),
	)
}

func (m model) settingsView(width, height int) string {
	leftWidth := max(34, width/3)
	list := "THÈMES DE COULEURS\n\n"
	for index, theme := range themes {
		state := theme.Description
		if theme.ID == m.config.Theme {
			state = "Actif · " + state
		}
		list += listLine(index == m.cursor, theme.Name, state) + "\n"
	}
	list += "\n" + mutedStyle.Render("j/k choisir · entrée appliquer")

	backupStatus := "Aucun backup durant cette session"
	if !m.lastBackupAt.IsZero() {
		backupStatus = "Dernier backup : " + m.lastBackupAt.Local().Format("02/01/2006 15:04")
	}
	detailWidth := max(24, width-leftWidth-13)
	detail := "PARAMÈTRES\n\n"
	detail += activeStyle.Render("Thème actuel") + "\n" + themeName(m.config.Theme) + "  " + activeStyle.Render("██") + redStyle.Render("██") + "\n\n"
	detail += activeStyle.Render("Serveur API") + "\n" + truncate(m.config.APIURL, detailWidth) + "\n" + mutedStyle.Render("e modifier l'URL") + "\n\n"
	detail += activeStyle.Render("Fichiers locaux JSON") + "\n"
	detail += "config.json · state.json\n"
	detail += truncate(userConfigDirectory(m.activeUserID), detailWidth) + "\n\n"
	detail += activeStyle.Render("Backup personnel") + "\n" + backupStatus + "\n"
	detail += mutedStyle.Render("u envoyer vers le serveur · o restaurer depuis le serveur") + "\n\n"
	detail += mutedStyle.Render("Le token JWT est volontairement exclu du backup.")

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		panelStyle.Width(leftWidth).Height(height-2).Render(list),
		panelStyle.Width(width-leftWidth-9).Height(height-2).Render(detail),
	)
}

func (m model) studyView(width, height int) string {
	leftWidth := max(26, width/3)
	list := "CARTES DUES\n\n"
	start, end := visibleBounds(len(m.data.Due), m.cursor, height-6)
	for index, card := range m.data.Due[start:end] {
		index += start
		prompt := card.Korean
		if m.studyDirection == "french-to-korean" {
			prompt = card.Translation
		}
		list += listLine(index == m.cursor, prompt, "À réviser") + "\n"
	}
	if len(m.data.Due) == 0 {
		list += activeStyle.Render("Session terminée")
	}
	detail := "RÉVISION\n\n"
	if len(m.data.Due) > 0 {
		card := m.data.Due[m.cursor]
		prompt, answer := card.Korean, card.Translation
		direction := "KO → FR"
		if m.studyDirection == "french-to-korean" {
			prompt, answer, direction = card.Translation, card.Korean, "FR → KO"
		}
		detail += mutedStyle.Render(direction+"  ·  v inverser") + "\n\n" + lipgloss.NewStyle().Foreground(text).Bold(true).Render(prompt) + "\n"
		detail += "\n"
		if m.revealed {
			detail += activeStyle.Render(answer) + "\n"
			if card.Romanization != "" {
				detail += mutedStyle.Render(card.Romanization) + "\n"
			}
			detail += "\n" + card.ExampleKorean + "\n" + mutedStyle.Render(card.ExampleTranslation) + "\n\n" + "1 À revoir   2 Avec hésitation   3 Bien retenue   4 Maîtrisée"
		} else {
			detail += mutedStyle.Render("a écrire une réponse  ·  espace révéler")
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Width(leftWidth).Height(height-2).Render(list), panelStyle.Width(width-leftWidth-9).Height(height-2).Render(detail))
}

func (m model) libraryView(width, height int) string {
	leftWidth := max(34, width/2)
	title := "CARTES"
	if !m.libraryCards {
		title = "DECKS"
	}
	list := title + "  " + mutedStyle.Render("tab pour changer") + "\n\n"
	if m.libraryCards {
		start, end := visibleBounds(len(m.data.Cards), m.cursor, height-6)
		for index, card := range m.data.Cards[start:end] {
			index += start
			list += listLine(index == m.cursor, card.Korean, card.Translation) + "\n"
		}
	} else {
		start, end := visibleBounds(len(m.data.Decks), m.cursor, height-6)
		for index, deck := range m.data.Decks[start:end] {
			index += start
			list += listLine(index == m.cursor, deck.Name, deck.Description) + "\n"
		}
	}
	detail := "DETAIL\n\n"
	if m.libraryCards && len(m.data.Cards) > 0 {
		card := m.data.Cards[m.cursor]
		detail += activeStyle.Render(card.Korean) + "\n" + card.Translation + "\n" + mutedStyle.Render(card.Romanization) + "\n\nDeck: " + card.DeckID + "\nTags: " + strings.Join(card.Tags, ", ") + "\nProchaine révision : " + card.ReviewState.NextReviewAt.Local().Format("02/01 15:04")
	}
	if !m.libraryCards && len(m.data.Decks) > 0 {
		deck := m.data.Decks[m.cursor]
		detail += activeStyle.Render(deck.Name) + "\n" + deck.Description + "\n\nID: " + deck.ID
	}
	detail += "\n\n" + mutedStyle.Render("n nouveau  ·  d supprimer  ·  : commandes avancées")
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Width(leftWidth).Height(height-2).Render(list), panelStyle.Width(width-leftWidth-9).Height(height-2).Render(detail))
}

func (m model) lessonsView(width, height int) string {
	leftWidth := max(30, width/3)
	list := "PARCOURS\n\n"
	start, end := visibleBounds(len(m.data.Lessons), m.cursor, height-6)
	for index, lesson := range m.data.Lessons[start:end] {
		index += start
		state := lesson.Level + " · À faire"
		if lesson.Progress.Completed {
			state = lesson.Level + " · Terminée ✓"
		}
		list += listLine(index == m.cursor, lesson.Title, state) + "\n"
	}
	detail := "LEÇON\n\n"
	if len(m.data.Lessons) > 0 {
		lesson := m.data.Lessons[m.cursor]
		detail += lesson.Title + "\n" + lesson.Description + "\n\n" + lesson.Content + "\n\n"
		if lesson.Progress.Completed {
			detail += "✓ Leçon terminée"
		} else {
			detail += "Entrée : valider la leçon"
		}
	}
	detailWidth := max(24, width-leftWidth-13)
	detail = scrollableText(detail, detailWidth, max(5, height-6), m.detailScroll)
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Width(leftWidth).Height(height-2).Render(list), panelStyle.Width(width-leftWidth-9).Height(height-2).Render(detail))
}

func (m model) journalView(width, height int) string {
	if m.journalEditing {
		field := func(label, value string, active bool) string {
			style := panelStyle.Width(min(76, width-14))
			if active {
				style = style.BorderForeground(brightGreen)
			}
			return style.Render(mutedStyle.Render(label) + "\n" + value)
		}
		heading := "NOUVELLE ENTRÉE DU JOURNAL"
		if m.journalID != "" {
			heading = "MODIFIER L'ENTRÉE DU JOURNAL"
		}
		content := heading + "\n\n"
		content += field("TITRE (FACULTATIF)", m.journalTitle, m.journalField == 0) + "\n"
		content += field("TEXTE EN CORÉEN", m.journalText, m.journalField == 1)
		content += "\n\n" + mutedStyle.Render("tab changer de champ  ·  entrée continuer/enregistrer  ·  échap annuler")
		return panelStyle.Width(width - 6).Height(height - 2).Render(content)
	}

	leftWidth := max(32, width/3)
	list := "MON JOURNAL\n\n"
	start, end := visibleBounds(len(m.data.Journal), m.cursor, height-6)
	for index, entry := range m.data.Journal[start:end] {
		index += start
		list += listLine(index == m.cursor, entry.Title, entry.CreatedAt.Local().Format("02/01/2006")) + "\n"
	}
	if len(m.data.Journal) == 0 {
		list += mutedStyle.Render("Aucune entrée")
	}
	detail := "TEXTE ET SUGGESTIONS\n\n"
	if len(m.data.Journal) > 0 {
		entry := m.data.Journal[m.cursor]
		detail += entry.CorrectedText + "\n\nTexte original : " + entry.OriginalText + "\n\n"
		for _, correction := range entry.Corrections {
			detail += correction.Original + " → " + correction.Replacement + "\n" + correction.Reason + "\n"
		}
		if len(entry.Sources) > 0 {
			detail += "\nLEÇONS UTILISÉES\n"
			for _, source := range entry.Sources {
				detail += fmt.Sprintf("\n%s · %s\n%s\n", source.Level, source.Title, source.Excerpt)
			}
		}
	}
	detail += "\n" + mutedStyle.Render("n nouvelle entrée  ·  e modifier l'entrée  ·  d supprimer")
	detail = scrollableText(detail, max(24, width-leftWidth-13), max(5, height-6), m.detailScroll)
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Width(leftWidth).Height(height-2).Render(list), panelStyle.Width(width-leftWidth-9).Height(height-2).Render(detail))
}

func (m model) statsView(width, height int) string {
	s := m.data.Stats
	metrics := fmt.Sprintf("PROGRESSION\n\nCartes             %d\nÀ réviser          %d\nNouvelles          %d\nMaîtrisées         %d\nDifficiles         %d\n\nRévisions aujourd'hui  %d\nPrécision              %.0f%%\nSérie actuelle         %d jours\nRecord                  %d jours", s.TotalCards, s.DueCards, s.NewCards, s.MasteredCards, s.DifficultCards, s.ReviewsToday, s.AccuracyPercent, s.CurrentStreak, s.LongestStreak)
	difficult := "À RENFORCER\n\n"
	for _, card := range m.data.Difficult {
		difficult += redStyle.Render(card.Korean) + "  " + card.Translation + fmt.Sprintf("  (%d oublis)\n", card.ReviewState.LapseCount)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Width(width/2-4).Height(height-2).Render(metrics), panelStyle.Width(width/2-5).Height(height-2).Render(difficult))
}

func (m model) searchView(width, height int) string {
	left := "DECKS\n\n"
	for _, deck := range m.search.Decks {
		left += activeStyle.Render(deck.Name) + "\n" + mutedStyle.Render(deck.Description) + "\n\n"
	}
	right := "CARTES\n\n"
	for _, card := range m.search.Cards {
		right += activeStyle.Render(card.Korean) + "  " + card.Translation + "\n"
	}
	if len(m.search.Decks)+len(m.search.Cards) == 0 {
		right += mutedStyle.Render("Appuie sur / pour rechercher dans tous les champs.")
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Width(width/2-4).Height(height-2).Render(left), panelStyle.Width(width/2-5).Height(height-2).Render(right))
}

func (m model) profileView(width, height int) string {
	if m.profileEditing {
		field := func(label, value string, active bool) string {
			style := panelStyle.Width(min(64, width-14))
			if active {
				style = style.BorderForeground(brightGreen)
			}
			return style.Render(mutedStyle.Render(label) + "\n" + value)
		}
		password := strings.Repeat("•", len([]rune(m.profilePass)))
		content := "MODIFIER MES INFORMATIONS\n\n"
		content += field("NOM", m.profileName, m.profileField == 0) + "\n"
		content += field("EMAIL", m.profileEmail, m.profileField == 1) + "\n"
		content += field("NOUVEAU MOT DE PASSE (FACULTATIF)", password, m.profileField == 2)
		content += "\n\n" + mutedStyle.Render("tab champ suivant  ·  entrée continuer/valider  ·  échap annuler")
		return panelStyle.Width(width - 6).Height(height - 2).Render(content)
	}

	content := "PROFIL\n\n" + activeStyle.Render(m.data.User.Name) + "\n" + m.data.User.Email + "\nID: " + m.data.User.ID
	content += "\n\n" + activeStyle.Render("e modifier mes informations")
	content += "\n" + redStyle.Render("D se déconnecter")
	return panelStyle.Width(width - 6).Height(height - 2).Render(content)
}

func (m model) adminView(width, height int) string {
	if m.adminEditing {
		field := func(label, value string, active bool) string {
			style := panelStyle.Width(min(64, width-14))
			if active {
				style = style.BorderForeground(brightGreen)
			}
			return style.Render(mutedStyle.Render(label) + "\n" + value)
		}
		password := strings.Repeat("•", len([]rune(m.adminPass)))
		content := "MODIFIER UN UTILISATEUR\n\n"
		content += mutedStyle.Render("Compte : ") + activeStyle.Render(m.adminName) + "  " + m.adminEmail + "\n\n"
		content += field("NOM", m.adminName, m.adminField == 0) + "\n"
		content += field("EMAIL", m.adminEmail, m.adminField == 1) + "\n"
		content += field("NOUVEAU MOT DE PASSE (FACULTATIF)", password, m.adminField == 2)
		content += "\n\n" + mutedStyle.Render("tab champ suivant  ·  entrée continuer/valider  ·  échap annuler")
		return panelStyle.Width(width - 6).Height(height - 2).Render(content)
	}

	leftWidth := max(34, width/2)
	list := "UTILISATEURS\n\n"
	start, end := visibleBounds(len(m.data.Users), m.cursor, height-7)
	for index, user := range m.data.Users[start:end] {
		index += start
		list += listLine(index == m.cursor, user.Name, user.Email) + "\n"
	}
	if len(m.data.Users) == 0 {
		list += mutedStyle.Render("Aucun compte non administrateur")
	}

	detail := "ADMINISTRATION\n\n"
	if len(m.data.Users) > 0 {
		user := m.data.Users[m.cursor]
		detail += activeStyle.Render(user.Name) + "\n" + user.Email + "\nID : " + user.ID + "\n\n" + mutedStyle.Render("e modifier cet utilisateur")
	}
	detail += "\n\nINDEX PÉDAGOGIQUE\n"
	if !m.data.RAG.Enabled {
		detail += mutedStyle.Render("RAG non configuré")
	} else if m.data.RAG.Ready {
		detail += activeStyle.Render(fmt.Sprintf("Prêt · %d passages", m.data.RAG.ChunkCount))
		detail += "\n" + mutedStyle.Render("i reconstruire l'index")
	} else {
		detail += mutedStyle.Render("À construire · i lancer l'indexation")
	}
	detail += "\n\n" + redStyle.Render("x préparer le reset de la base") + "\n\nSwagger : " + m.client.BaseURL + "/swagger/index.html"
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Width(leftWidth).Height(height-2).Render(list), panelStyle.Width(width-leftWidth-9).Height(height-2).Render(detail))
}

func (m model) helpView(width, height int) string {
	help := "AIDE\n\n h/l ou ←/→   changer d'onglet\n j/k ou ↑/↓   naviguer\n PgUp/PgDn    faire défiler le détail\n a             saisir une réponse\n v             inverser KO/FR\n espace        révéler une carte\n 1..4          indiquer la mémorisation\n n             créer dans la vue active\n d             supprimer l'élément actif\n e             modifier l'élément actif, l'URL API ou le profil\n i             reconstruire l'index RAG (admin)\n u             envoyer config + état vers le serveur\n o             restaurer config + état depuis le serveur\n D             se déconnecter\n x             préparer le reset administrateur\n tab           decks/cartes dans la bibliothèque\n /             recherche globale\n :             palette de commandes\n r             actualiser\n ?             fermer l'aide\n q             quitter\n\nCOMMANDES AVANCÉES\n deck-add NOM | DESCRIPTION\n deck-update ID | NOM | DESCRIPTION\n decks-description ID1,ID2 | DESCRIPTION\n card-add DECK_ID | CORÉEN | TRADUCTION | ROMANISATION\n card-update ID | CORÉEN | TRADUCTION | ROMANISATION\n cards-move ID1,ID2 | DECK_ID\n lesson-complete ID\n import DECK_ID | FICHIER.csv\n export FICHIER.csv\n rag-reindex"
	return panelStyle.BorderForeground(brightGreen).Width(width - 6).Height(height - 2).Render(help)
}

func (m model) footerView(width int) string {
	status := m.status
	if m.err != nil {
		status = errorStyle.Render(m.err.Error())
	}
	if m.loading {
		status = activeStyle.Render("Synchronisation...")
	}
	if m.inputMode != "" {
		prefix := ":"
		if m.inputMode == "search" {
			prefix = "/"
		} else if m.inputMode == "api-url" {
			prefix = "API> "
		}
		status = activeStyle.Render(prefix) + m.input + "█"
	}
	hints := mutedStyle.Render(" h/l onglets  j/k naviguer  / chercher  : commandes  D déconnexion  ? aide  q quitter ")
	space := max(1, width-lipgloss.Width(status)-lipgloss.Width(hints)-2)
	return lipgloss.NewStyle().Background(lipgloss.Color("#0D1512")).Width(width).Render(" " + status + strings.Repeat(" ", space) + hints)
}

func (m model) itemCount() int {
	switch m.tab {
	case tabHome:
		return len(homeOptions)
	case tabStudy:
		return len(m.data.Due)
	case tabLibrary:
		if m.libraryCards {
			return len(m.data.Cards)
		}
		return len(m.data.Decks)
	case tabLessons:
		return len(m.data.Lessons)
	case tabJournal:
		return len(m.data.Journal)
	case tabSettings:
		return len(themes)
	case tabAdmin:
		return len(m.data.Users)
	}
	return 0
}

func (m model) deleteCommand() string {
	switch m.tab {
	case tabLibrary:
		if m.libraryCards && len(m.data.Cards) > 0 {
			return "card-delete " + m.data.Cards[m.cursor].ID
		}
		if !m.libraryCards && len(m.data.Decks) > 0 {
			return "deck-delete " + m.data.Decks[m.cursor].ID
		}
	case tabJournal:
		if len(m.data.Journal) > 0 {
			return "journal-delete " + m.data.Journal[m.cursor].ID
		}
	}
	return ""
}

func (m model) visibleTabs() []string {
	if !m.data.User.IsAdmin {
		return tabs
	}
	return append(append([]string{}, tabs...), "ADMIN")
}

func (m *model) resetCursorForTab() {
	m.cursor = 0
	if m.tab == tabSettings {
		m.cursor = themeIndex(m.config.Theme)
	}
}

func (m model) appState() AppState {
	return AppState{
		Version:        localDataVersion,
		ActiveView:     m.activeTabID(),
		StudyDirection: m.studyDirection,
		LibraryCards:   m.libraryCards,
		UpdatedAt:      time.Now().UTC(),
	}
}

func (m model) persistState() {
	if m.activeUserID != "" {
		_ = saveUserState(m.activeUserID, m.appState())
	}
}

func (m *model) activateUserProfile(userID string) {
	if strings.TrimSpace(userID) == "" {
		return
	}

	configFallback := defaultConfig()
	configFallback.APIURL = m.client.BaseURL
	stateFallback := defaultState()
	if !hasUserProfiles() {
		configFallback = m.config
		if legacyState, err := loadState(); err == nil {
			stateFallback = legacyState
		}
	}

	configValue, configErr := loadUserConfig(userID, configFallback)
	stateValue, stateErr := loadUserState(userID, stateFallback)
	if configErr != nil {
		configValue = configFallback
	}
	if stateErr != nil {
		stateValue = stateFallback
	}
	configValue.APIURL = m.client.BaseURL
	m.activeUserID = userID
	m.config = configValue
	m.libraryCards = stateValue.LibraryCards
	m.studyDirection = stateValue.StudyDirection
	m.tab = tabIndexForID(stateValue.ActiveView)
	m.resetCursorForTab()
	m.detailScroll = 0
	applyTheme(configValue.Theme)
	_ = saveUserConfig(userID, configValue)
	_ = saveUserState(userID, stateValue)
	if configErr != nil {
		m.err = configErr
	} else if stateErr != nil {
		m.err = stateErr
	}
}

func (m model) activeTabID() string {
	if m.tab == tabAdmin {
		return "admin"
	}
	if m.tab >= 0 && m.tab < len(tabIDs) {
		return tabIDs[m.tab]
	}
	return "home"
}

func tabIndexForID(id string) int {
	if id == "admin" {
		return tabAdmin
	}
	for index, candidate := range tabIDs {
		if candidate == id {
			return index
		}
	}
	return tabHome
}

func isViewSupported(id string) bool {
	if id == "admin" {
		return true
	}
	for _, candidate := range tabIDs {
		if candidate == id {
			return true
		}
	}
	return false
}

func isThemeSupported(id string) bool {
	for _, theme := range themes {
		if theme.ID == id {
			return true
		}
	}
	return false
}

func themeIndex(id string) int {
	for index, theme := range themes {
		if theme.ID == id {
			return index
		}
	}
	return 0
}

func themeName(id string) string {
	return themes[themeIndex(id)].Name
}

func applyTheme(id string) string {
	theme := themes[themeIndex(id)]
	green = lipgloss.Color(theme.Accent)
	brightGreen = lipgloss.Color(theme.Bright)
	red = lipgloss.Color(theme.Danger)
	muted = lipgloss.Color(theme.Muted)
	panel = lipgloss.Color(theme.Panel)
	border = lipgloss.Color(theme.Border)
	text = lipgloss.Color(theme.Text)

	baseStyle = lipgloss.NewStyle().Foreground(text)
	activeStyle = lipgloss.NewStyle().Foreground(brightGreen).Bold(true)
	mutedStyle = lipgloss.NewStyle().Foreground(muted)
	redStyle = lipgloss.NewStyle().Foreground(red)
	selectedStyle = lipgloss.NewStyle().Background(green).Foreground(lipgloss.Color(theme.Selection)).Bold(true)
	panelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).Background(panel).Padding(1, 2)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8B78")).Bold(true)
	return theme.ID
}

func loadDashboard(client *APIClient) tea.Cmd {
	return func() tea.Msg { data, err := client.LoadDashboard(); return dashboardMsg{data: data, err: err} }
}
func loginCommand(client *APIClient, email, password string) tea.Cmd {
	return func() tea.Msg {
		result, err := client.Login(email, password)
		return loginMsg{result: result, err: err}
	}
}
func registerCommand(client *APIClient, name, email, password string) tea.Cmd {
	return func() tea.Msg {
		result, err := client.Register(name, email, password)
		return loginMsg{result: result, err: err}
	}
}
func answerCommand(client *APIClient, id, rating string) tea.Cmd {
	return func() tea.Msg {
		err := client.Answer(id, rating)
		labels := map[string]string{
			"again": "À revoir",
			"hard":  "Avec hésitation",
			"good":  "Bien retenue",
			"easy":  "Maîtrisée",
		}
		return mutationMsg{status: "Mémorisation : " + labels[rating], err: err}
	}
}
func checkCommand(client *APIClient, id, answer, direction string) tea.Cmd {
	return func() tea.Msg {
		result, err := client.Check(id, answer, direction)
		return checkMsg{result: result, err: err}
	}
}
func executeCommand(client *APIClient, command string) tea.Cmd {
	return func() tea.Msg { status, err := client.Execute(command); return mutationMsg{status: status, err: err} }
}
func updateProfileCommand(client *APIClient, name, email, password string) tea.Cmd {
	return func() tea.Msg {
		err := client.UpdateProfile(name, email, password)
		return mutationMsg{status: "Profil modifié", err: err}
	}
}
func saveJournalCommand(client *APIClient, id, title, text string) tea.Cmd {
	return func() tea.Msg {
		err := client.SaveJournalEntry(id, title, text)
		return journalSaveMsg{created: strings.TrimSpace(id) == "", err: err}
	}
}
func adminUpdateUserCommand(client *APIClient, userID, name, email, password string) tea.Cmd {
	return func() tea.Msg {
		return adminUserMsg{err: client.AdminUpdateUser(userID, name, email, password)}
	}
}
func searchCommand(client *APIClient, query string) tea.Cmd {
	return func() tea.Msg { result, err := client.Search(query); return searchMsg{result: result, err: err} }
}
func uploadBackupCommand(client *APIClient, config AppConfig, state AppState) tea.Cmd {
	return func() tea.Msg {
		backup, err := client.UploadBackup(config, state)
		return backupMsg{action: "upload", backup: backup, err: err}
	}
}
func downloadBackupCommand(client *APIClient) tea.Cmd {
	return func() tea.Msg {
		backup, err := client.DownloadBackup()
		return backupMsg{action: "download", backup: backup, err: err}
	}
}

func listLine(selected bool, primary, secondary string) string {
	line := fmt.Sprintf("%-22s %s", truncate(primary, 22), truncate(secondary, 28))
	if selected {
		return selectedStyle.Width(54).Render(line)
	}
	return line
}
func truncate(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:max(0, limit-1)]) + "…"
}
func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}
	return string(runes[:len(runes)-1])
}
func clamp(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
func visibleBounds(total, cursor, limit int) (int, int) {
	if total == 0 {
		return 0, 0
	}
	limit = max(1, limit)
	start := max(0, cursor-limit/2)
	end := min(total, start+limit)
	start = max(0, end-limit)
	return start, end
}
func wrapHeaderItems(items []string, width int) string {
	rows := make([]string, 0, 2)
	current := ""
	for _, item := range items {
		candidate := item
		if current != "" {
			candidate = current + " " + item
		}
		if current != "" && lipgloss.Width(candidate) > width {
			rows = append(rows, current)
			current = item
			continue
		}
		current = candidate
	}
	if current != "" {
		rows = append(rows, current)
	}
	return strings.Join(rows, "\n")
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func scrollableText(value string, width, height, offset int) string {
	lines := wrapText(value, width)
	if len(lines) <= height {
		return strings.Join(lines, "\n")
	}

	contentHeight := max(1, height-1)
	maxOffset := max(0, len(lines)-contentHeight)
	offset = clamp(offset, 0, maxOffset)
	end := min(len(lines), offset+contentHeight)
	position := fmt.Sprintf("PgUp/PgDn · %d-%d/%d", offset+1, end, len(lines))
	return strings.Join(lines[offset:end], "\n") + "\n" + mutedStyle.Render(position)
}

func wrapText(value string, width int) []string {
	if width <= 0 {
		return []string{value}
	}
	result := make([]string, 0)
	for _, sourceLine := range strings.Split(value, "\n") {
		words := strings.Fields(sourceLine)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		line := words[0]
		for _, word := range words[1:] {
			candidate := line + " " + word
			if lipgloss.Width(candidate) <= width {
				line = candidate
				continue
			}
			result = append(result, line)
			line = word
		}
		result = append(result, line)
	}
	return result
}
func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
func commandHint(tab int) string {
	hints := map[int]string{
		tabLibrary:  "deck-add / card-add",
		tabLessons:  "lesson-complete",
		tabJournal:  "utilise n pour écrire ou e pour modifier",
		tabSettings: "e URL API / u envoyer / o restaurer",
		tabAdmin:    "admin-user / reset",
	}
	return "Commande : " + hints[tab]
}
