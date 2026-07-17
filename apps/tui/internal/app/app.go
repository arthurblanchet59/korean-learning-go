package app

import (
	"flag"
	"fmt"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
	"time"
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

var version = "dev"

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

func (m model) updateNavigation(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showHelp {
		switch key.String() {
		case "?", "esc":
			m.showHelp = false
			m.detailScroll = 0
		case "pgdown", "ctrl+d":
			m.detailScroll += max(3, m.height/3)
		case "pgup", "ctrl+u":
			m.detailScroll = max(0, m.detailScroll-max(3, m.height/3))
		case "q", "ctrl+c":
			m.persistState()
			return m, tea.Quit
		}
		return m, nil
	}

	switch key.String() {
	case "q", "ctrl+c":
		m.persistState()
		return m, tea.Quit
	case "D":
		return m.logout(), nil
	case "?":
		m.showHelp = !m.showHelp
		m.detailScroll = 0
	case ":", "c":
		m.inputMode = "command"
		m.input = ""
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
		} else if m.tab == tabLibrary {
			m.inputMode = "command"
			if m.libraryCards && len(m.data.Cards) > 0 {
				card := m.data.Cards[m.cursor]
				m.input = "card-update " + strings.Join([]string{
					card.ID,
					card.Korean,
					card.Translation,
					card.Romanization,
				}, " | ")
			} else if !m.libraryCards && len(m.data.Decks) > 0 {
				deck := m.data.Decks[m.cursor]
				m.input = "deck-update " + strings.Join([]string{
					deck.ID,
					deck.Name,
					deck.Description,
				}, " | ")
			}
		}
	case "x":
		if m.tab == tabAdmin && m.data.User.IsAdmin {
			m.inputMode = "command"
			m.input = "reset CONFIRM"
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
	configFallback.APIURL = m.config.APIURL
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
	configValue.APIURL = m.config.APIURL
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

func Run() {
	apiURL := flag.String("api", "", "URL du backend (prioritaire sur config.json)")
	showVersion := flag.Bool("version", false, "Affiche la version et quitte")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	configValue, configErr := loadConfig()
	stateValue, stateErr := loadState()
	explicitAPIURL := false
	if envURL := strings.TrimSpace(os.Getenv("KOREAN_API_URL")); envURL != "" {
		configValue.APIURL = envURL
		explicitAPIURL = true
	}
	if strings.TrimSpace(*apiURL) != "" {
		configValue.APIURL = *apiURL
		explicitAPIURL = true
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

	runtimeAPIURL := configValue.APIURL
	usingLocalFallback := false
	if !explicitAPIURL {
		runtimeAPIURL, usingLocalFallback = resolveAvailableAPIURL(configValue.APIURL, localAPIURL)
	}
	client := NewAPIClient(runtimeAPIURL)
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
	} else if usingLocalFallback {
		application.status = "API distante indisponible : backend local utilisé"
	}
	if _, err := tea.NewProgram(application, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
