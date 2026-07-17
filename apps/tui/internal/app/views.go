package app

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

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
	if m.client != nil && m.client.BaseURL != m.config.APIURL {
		detail += activeStyle.Render("Repli local actif") + "\n" + m.client.BaseURL + "\n\n"
	}
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
	detail += "\n\n" + mutedStyle.Render("n nouveau  ·  e modifier  ·  d supprimer  ·  c options avancées")
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
	detail += "\n\n" + redStyle.Render("x préparer le reset de la base") + "\n\nSwagger : " + m.client.BaseURL + "/swagger/index.html"
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Width(leftWidth).Height(height-2).Render(list), panelStyle.Width(width-leftWidth-9).Height(height-2).Render(detail))
}

func (m model) helpView(width, height int) string {
	help := "AIDE\n\n h/l ou ←/→   changer d'onglet\n j/k ou ↑/↓   naviguer\n PgUp/PgDn    faire défiler le détail\n a             saisir une réponse\n v             inverser KO/FR\n espace        révéler une carte\n 1..4          indiquer la mémorisation\n n             créer dans la bibliothèque ou le journal\n d             supprimer dans la bibliothèque ou le journal\n e             modifier l'élément actif, l'URL API ou le profil\n u             envoyer config + état vers le serveur\n o             restaurer config + état depuis le serveur\n D             se déconnecter\n x             préparer le reset administrateur\n tab           decks/cartes dans la bibliothèque\n /             recherche globale\n c ou :        ouvrir les commandes avancées\n r             actualiser\n ? ou échap    fermer l'aide\n q             quitter\n\nCOMMANDES AVANCÉES\n deck-add NOM | DESCRIPTION\n deck-update ID | NOM | DESCRIPTION\n decks-delete ID1,ID2\n decks-description ID1,ID2 | DESCRIPTION\n card-add DECK_ID | CORÉEN | TRADUCTION | ROMANISATION\n card-update ID | CORÉEN | TRADUCTION | ROMANISATION\n cards-delete ID1,ID2\n cards-move ID1,ID2 | DECK_ID\n lesson-complete ID\n import DECK_ID | FICHIER.csv\n export FICHIER.csv"
	content := scrollableText(help, max(24, width-12), max(5, height-6), m.detailScroll)
	return panelStyle.BorderForeground(brightGreen).Width(width - 6).Height(height - 2).Render(content)
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
	hintText := " h/l onglets  j/k naviguer  / chercher  c commandes  D déconnexion  ? aide  q quitter "
	compactHint := " h/l onglets  j/k naviguer  c commandes  ? aide  q quitter "
	available := width - lipgloss.Width(status) - 3
	if lipgloss.Width(hintText) > available {
		hintText = compactHint
	}
	if lipgloss.Width(hintText) > available {
		hintText = ""
	}
	hints := mutedStyle.Render(hintText)
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
