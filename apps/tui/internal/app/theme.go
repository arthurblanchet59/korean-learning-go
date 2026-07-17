package app

import (
	"github.com/charmbracelet/lipgloss"
)

func init() {
	applyTheme("emerald")
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
