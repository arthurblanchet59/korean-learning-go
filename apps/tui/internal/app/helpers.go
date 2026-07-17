package app

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
)

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
		tabLibrary:  "deck-add / card-add / import / export",
		tabLessons:  "lesson-complete",
		tabJournal:  "utilise n pour écrire ou e pour modifier",
		tabSettings: "e URL API / u envoyer / o restaurer",
		tabAdmin:    "admin-user / reset",
	}
	hint := hints[tab]
	if hint == "" {
		hint = "saisis une commande avancée (? pour consulter l'aide)"
	}
	return "Commande : " + hint
}
