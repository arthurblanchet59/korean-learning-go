package app

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

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
	case tea.KeyTab, tea.KeyDown:
		m.loginField = (m.loginField + 1) % fieldCount
	case tea.KeyUp:
		m.loginField = (m.loginField - 1 + fieldCount) % fieldCount
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
