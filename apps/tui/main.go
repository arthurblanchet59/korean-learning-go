package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arthurblanchet59/korean-learning-go/packages/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	green       = lipgloss.Color("#4EA67A")
	brightGreen = lipgloss.Color("#78D7A6")
	red         = lipgloss.Color("#E05A47")
	muted       = lipgloss.Color("#7E8B86")
	panel       = lipgloss.Color("#17211E")
	border      = lipgloss.Color("#395048")
	text        = lipgloss.Color("#F1F5F3")

	baseStyle     = lipgloss.NewStyle().Foreground(text)
	activeStyle   = lipgloss.NewStyle().Foreground(brightGreen).Bold(true)
	mutedStyle    = lipgloss.NewStyle().Foreground(muted)
	redStyle      = lipgloss.NewStyle().Foreground(red)
	selectedStyle = lipgloss.NewStyle().Background(green).Foreground(lipgloss.Color("#08110E")).Bold(true)
	panelStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).Background(panel).Padding(1, 2)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8B78")).Bold(true)
)

var tabs = []string{"REVISION", "BIBLIOTHEQUE", "LECONS", "JOURNAL", "PROGRESSION", "RECHERCHE", "PROFIL"}

type model struct {
	client         *APIClient
	data           DashboardData
	search         SearchResult
	width          int
	height         int
	tab            int
	cursor         int
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
	studyDirection string
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
type searchMsg struct {
	result SearchResult
	err    error
}
type checkMsg struct {
	result core.AnswerCheck
	err    error
}

func main() {
	apiURL := flag.String("api", envOr("KOREAN_API_URL", "http://localhost:8080"), "URL du backend")
	flag.Parse()

	client := NewAPIClient(*apiURL)
	application := model{client: client, loggedIn: client.Token != "", loading: client.Token != "", libraryCards: true, loginEmail: "admin@korean.local", studyDirection: "korean-to-french"}
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
			m.status = "Donnees synchronisees"
			m.cursor = clamp(m.cursor, 0, max(0, m.itemCount()-1))
		} else if strings.Contains(strings.ToLower(msg.err.Error()), "token") || strings.Contains(strings.ToLower(msg.err.Error()), "unauthorized") {
			m.loggedIn = false
			m.client.Token = ""
			_ = os.Remove(tokenPath())
		}
		return m, nil
	case loginMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.loggedIn = true
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
	case searchMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.search = msg.result
			m.status = fmt.Sprintf("%d resultat(s)", len(msg.result.Cards)+len(msg.result.Decks))
		}
		return m, nil
	case checkMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.revealed = true
			if msg.result.Correct {
				m.status = "Bonne reponse"
			} else {
				m.status = "Reponse attendue: " + msg.result.Expected
			}
		}
		return m, nil
	case tea.KeyMsg:
		if !m.loggedIn {
			return m.updateLogin(msg)
		}
		if m.inputMode != "" {
			return m.updateInput(msg)
		}
		return m.updateNavigation(msg)
	}
	return m, nil
}

func (m model) updateLogin(key tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.registering && m.loginField == 0 {
			m.loginName = trimLastRune(m.loginName)
		} else if (m.registering && m.loginField == 1) || (!m.registering && m.loginField == 0) {
			m.loginEmail = trimLastRune(m.loginEmail)
		} else {
			m.loginPass = trimLastRune(m.loginPass)
		}
	case tea.KeyRunes:
		if m.registering && m.loginField == 0 {
			m.loginName += string(key.Runes)
		} else if (m.registering && m.loginField == 1) || (!m.registering && m.loginField == 0) {
			m.loginEmail += string(key.Runes)
		} else {
			m.loginPass += string(key.Runes)
		}
	}
	if key.String() == "ctrl+r" {
		m.registering = !m.registering
		m.loginField = 0
		m.err = nil
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
		m.loading, m.err = true, nil
		if mode == "answer" && len(m.data.Due) > 0 {
			return m, checkCommand(m.client, m.data.Due[m.cursor].ID, value, m.studyDirection)
		}
		if mode == "search" {
			m.tab = 5
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
		return m, tea.Quit
	case "?":
		m.showHelp = !m.showHelp
	case ":":
		m.inputMode = "command"
		m.status = commandHint(m.tab)
	case "/":
		m.inputMode = "search"
		m.input = ""
	case "h", "left":
		m.tab = (m.tab - 1 + len(tabs)) % len(tabs)
		m.cursor = 0
		m.revealed = false
	case "l", "right":
		m.tab = (m.tab + 1) % len(tabs)
		m.cursor = 0
		m.revealed = false
	case "j", "down":
		m.cursor = clamp(m.cursor+1, 0, max(0, m.itemCount()-1))
		m.revealed = false
	case "k", "up":
		m.cursor = clamp(m.cursor-1, 0, max(0, m.itemCount()-1))
		m.revealed = false
	case "r":
		m.loading = true
		return m, loadDashboard(m.client)
	case "tab":
		if m.tab == 1 {
			m.libraryCards = !m.libraryCards
			m.cursor = 0
		}
	case " ":
		if m.tab == 0 {
			m.revealed = true
		}
	case "a":
		if m.tab == 0 && len(m.data.Due) > 0 {
			m.inputMode = "answer"
			m.input = ""
			m.status = "Ecris ta reponse"
		}
	case "v":
		if m.tab == 0 {
			if m.studyDirection == "korean-to-french" {
				m.studyDirection = "french-to-korean"
			} else {
				m.studyDirection = "korean-to-french"
			}
			m.revealed = false
		}
	case "1", "2", "3", "4":
		if m.tab == 0 && m.revealed && len(m.data.Due) > 0 {
			ratings := map[string]string{"1": "again", "2": "hard", "3": "good", "4": "easy"}
			m.loading = true
			return m, answerCommand(m.client, m.data.Due[m.cursor].ID, ratings[key.String()])
		}
	case "enter":
		if m.tab == 2 && len(m.data.Lessons) > 0 {
			return m, executeCommand(m.client, "lesson-complete "+m.data.Lessons[m.cursor].ID+" | 100")
		}
	case "n":
		m.inputMode = "command"
		switch m.tab {
		case 1:
			if m.libraryCards {
				m.input = "card-add "
			} else {
				m.input = "deck-add "
			}
		case 3:
			m.input = "journal-add "
		}
	case "d":
		if command := m.deleteCommand(); command != "" {
			m.loading = true
			return m, executeCommand(m.client, command)
		}
	}
	return m, nil
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
		modeLabel = "Creation d'un compte personnel"
		fields += field("NOM", m.loginName, m.loginField == 0) + "\n"
	}
	emailField, passwordField := 0, 1
	if m.registering {
		emailField, passwordField = 1, 2
	}
	fields += field("EMAIL", m.loginEmail, m.loginField == emailField) + "\n" + field("MOT DE PASSE", password, m.loginField == passwordField)
	content := activeStyle.Render("한  KOREAN LEARNING") + "\n\n" + mutedStyle.Render(modeLabel) + "\n\n" + fields + "\n\n" + mutedStyle.Render("tab changer  ·  entree valider  ·  ctrl+r connexion/inscription  ·  esc quitter")
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
	items := make([]string, 0, len(tabs))
	for index, tab := range tabs {
		if index == m.tab {
			items = append(items, selectedStyle.Padding(0, 1).Render(tab))
		} else {
			items = append(items, mutedStyle.Padding(0, 1).Render(tab))
		}
	}
	return lipgloss.NewStyle().Width(width).Padding(0, 1).Render(brand + "\n" + strings.Join(items, " "))
}

func (m model) bodyView(width int, height int) string {
	if m.loading && len(m.data.Decks) == 0 {
		return panelStyle.Width(width - 6).Height(height - 2).Render("Synchronisation avec l'API...")
	}
	switch m.tab {
	case 0:
		return m.studyView(width, height)
	case 1:
		return m.libraryView(width, height)
	case 2:
		return m.lessonsView(width, height)
	case 3:
		return m.journalView(width, height)
	case 4:
		return m.statsView(width, height)
	case 5:
		return m.searchView(width, height)
	default:
		return m.profileView(width, height)
	}
}

func (m model) studyView(width, height int) string {
	leftWidth := max(26, width/3)
	list := "CARTES DUES\n\n"
	start, end := visibleBounds(len(m.data.Due), m.cursor, height-6)
	for index, card := range m.data.Due[start:end] {
		index += start
		list += listLine(index == m.cursor, card.Korean, card.Translation) + "\n"
	}
	if len(m.data.Due) == 0 {
		list += activeStyle.Render("Session terminee")
	}
	detail := "REVISION\n\n"
	if len(m.data.Due) > 0 {
		card := m.data.Due[m.cursor]
		prompt, answer := card.Korean, card.Translation
		direction := "KO → FR"
		if m.studyDirection == "french-to-korean" {
			prompt, answer, direction = card.Translation, card.Korean, "FR → KO"
		}
		detail += mutedStyle.Render(direction+"  ·  v inverser") + "\n\n" + lipgloss.NewStyle().Foreground(text).Bold(true).Render(prompt) + "\n"
		if m.studyDirection == "korean-to-french" {
			detail += mutedStyle.Render(card.Romanization) + "\n"
		}
		detail += "\n"
		if m.revealed {
			detail += activeStyle.Render(answer) + "\n\n" + card.ExampleKorean + "\n" + mutedStyle.Render(card.ExampleTranslation) + "\n\n" + "1 Encore   2 Difficile   3 Correct   4 Facile"
		} else {
			detail += mutedStyle.Render("a ecrire une reponse  ·  espace reveler")
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
		detail += activeStyle.Render(card.Korean) + "\n" + card.Translation + "\n" + mutedStyle.Render(card.Romanization) + "\n\nDeck: " + card.DeckID + "\nTags: " + strings.Join(card.Tags, ", ") + "\nProchaine revision: " + card.ReviewState.NextReviewAt.Local().Format("02/01 15:04")
	}
	if !m.libraryCards && len(m.data.Decks) > 0 {
		deck := m.data.Decks[m.cursor]
		detail += activeStyle.Render(deck.Name) + "\n" + deck.Description + "\n\nID: " + deck.ID
	}
	detail += "\n\n" + mutedStyle.Render("n nouveau  ·  d supprimer  ·  : commandes avancees")
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Width(leftWidth).Height(height-2).Render(list), panelStyle.Width(width-leftWidth-9).Height(height-2).Render(detail))
}

func (m model) lessonsView(width, height int) string {
	leftWidth := max(30, width/3)
	list := "PARCOURS\n\n"
	start, end := visibleBounds(len(m.data.Lessons), m.cursor, height-6)
	for index, lesson := range m.data.Lessons[start:end] {
		index += start
		state := lesson.Level
		if lesson.Progress.Completed {
			state += " ✓"
		}
		list += listLine(index == m.cursor, lesson.Title, state) + "\n"
	}
	detail := "LECON\n\n"
	if len(m.data.Lessons) > 0 {
		lesson := m.data.Lessons[m.cursor]
		detail += activeStyle.Render(lesson.Title) + "\n" + mutedStyle.Render(lesson.Description) + "\n\n" + lesson.Content + "\n\n" + fmt.Sprintf("Score: %d%%", lesson.Progress.Score) + "\n" + mutedStyle.Render("entree terminer la lecon")
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Width(leftWidth).Height(height-2).Render(list), panelStyle.Width(width-leftWidth-9).Height(height-2).Render(detail))
}

func (m model) journalView(width, height int) string {
	leftWidth := max(32, width/3)
	list := "MON JOURNAL\n\n"
	start, end := visibleBounds(len(m.data.Journal), m.cursor, height-6)
	for index, entry := range m.data.Journal[start:end] {
		index += start
		list += listLine(index == m.cursor, entry.Title, entry.CreatedAt.Local().Format("02/01/2006")) + "\n"
	}
	if len(m.data.Journal) == 0 {
		list += mutedStyle.Render("Aucune entree")
	}
	detail := "CORRECTION\n\n"
	if len(m.data.Journal) > 0 {
		entry := m.data.Journal[m.cursor]
		detail += activeStyle.Render(entry.CorrectedText) + "\n\n" + mutedStyle.Render("Texte original: ") + entry.OriginalText + "\n\n"
		for _, correction := range entry.Corrections {
			detail += redStyle.Render(correction.Original+" → "+correction.Replacement) + "\n" + mutedStyle.Render(correction.Reason) + "\n"
		}
	}
	detail += "\n" + mutedStyle.Render("n ecrire  ·  d supprimer  ·  : journal-update")
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Width(leftWidth).Height(height-2).Render(list), panelStyle.Width(width-leftWidth-9).Height(height-2).Render(detail))
}

func (m model) statsView(width, height int) string {
	s := m.data.Stats
	metrics := fmt.Sprintf("PROGRESSION\n\nCartes             %d\nA reviser          %d\nNouvelles          %d\nMaitrisees         %d\nDifficiles         %d\n\nRevisions aujourd'hui  %d\nPrecision              %.0f%%\nSerie actuelle         %d jours\nRecord                  %d jours", s.TotalCards, s.DueCards, s.NewCards, s.MasteredCards, s.DifficultCards, s.ReviewsToday, s.AccuracyPercent, s.CurrentStreak, s.LongestStreak)
	difficult := "A RENFORCER\n\n"
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
	content := "PROFIL\n\n" + activeStyle.Render(m.data.User.Name) + "\n" + m.data.User.Email + "\nID: " + m.data.User.ID + "\n\n" + mutedStyle.Render(": profile NOM | EMAIL")
	if m.data.User.IsAdmin {
		content += "\n\nADMINISTRATION\n\n: admin-user ID | NOM\n: reset CONFIRM\n\nSwagger: " + m.client.BaseURL + "/swagger/index.html"
	}
	return panelStyle.Width(width - 6).Height(height - 2).Render(content)
}

func (m model) helpView(width, height int) string {
	help := "AIDE\n\n h/l ou ←/→   changer d'onglet\n j/k ou ↑/↓   naviguer\n a             saisir une reponse\n v             inverser KO/FR\n espace        reveler une carte\n 1..4          noter une revision\n n             creer dans la vue active\n d             supprimer l'element actif\n tab           decks/cartes dans la bibliotheque\n /             recherche globale\n :             palette de commandes\n r             actualiser\n ?             fermer l'aide\n q             quitter\n\nCOMMANDES\n deck-add NOM | DESCRIPTION\n deck-update ID | NOM | DESCRIPTION\n decks-description ID1,ID2 | DESCRIPTION\n card-add DECK_ID | COREEN | TRADUCTION | ROMANISATION\n card-update ID | COREEN | TRADUCTION | ROMANISATION\n cards-move ID1,ID2 | DECK_ID\n journal-add TITRE | TEXTE\n journal-update ID | TITRE | TEXTE\n lesson-complete ID | SCORE\n import DECK_ID | FICHIER.csv\n export FICHIER.csv"
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
		}
		status = activeStyle.Render(prefix) + m.input + "█"
	}
	hints := mutedStyle.Render(" h/l onglets  j/k naviguer  / chercher  : commandes  ? aide  q quitter ")
	space := max(1, width-lipgloss.Width(status)-lipgloss.Width(hints)-2)
	return lipgloss.NewStyle().Background(lipgloss.Color("#0D1512")).Width(width).Render(" " + status + strings.Repeat(" ", space) + hints)
}

func (m model) itemCount() int {
	switch m.tab {
	case 0:
		return len(m.data.Due)
	case 1:
		if m.libraryCards {
			return len(m.data.Cards)
		}
		return len(m.data.Decks)
	case 2:
		return len(m.data.Lessons)
	case 3:
		return len(m.data.Journal)
	}
	return 0
}

func (m model) deleteCommand() string {
	switch m.tab {
	case 1:
		if m.libraryCards && len(m.data.Cards) > 0 {
			return "card-delete " + m.data.Cards[m.cursor].ID
		}
		if !m.libraryCards && len(m.data.Decks) > 0 {
			return "deck-delete " + m.data.Decks[m.cursor].ID
		}
	case 3:
		if len(m.data.Journal) > 0 {
			return "journal-delete " + m.data.Journal[m.cursor].ID
		}
	}
	return ""
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
		return mutationMsg{status: "Revision enregistree: " + rating, err: err}
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
func searchCommand(client *APIClient, query string) tea.Cmd {
	return func() tea.Msg { result, err := client.Search(query); return searchMsg{result: result, err: err} }
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
func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
func commandHint(tab int) string {
	hints := []string{"", "deck-add / card-add", "lesson-complete", "journal-add", "export", "", "profile / admin-user / reset"}
	return "Commande: " + hints[tab]
}
