package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	accentColor = lipgloss.Color("99")  // purple
	dimColor    = lipgloss.Color("240") // gray
	goodColor   = lipgloss.Color("42")  // green
	warnColor   = lipgloss.Color("220") // yellow
	errorColor  = lipgloss.Color("196") // red
	baseStyle   = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(accentColor)
	headerStyle = lipgloss.NewStyle().
			Foreground(accentColor).Bold(true)
	highlightStyle = lipgloss.NewStyle().
			Foreground(goodColor).Bold(true)
	dimStyle = lipgloss.NewStyle().
			Foreground(dimColor)
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).Bold(true)
)

// FeedTable renders a table of feeds.
func FeedTable(headers []string, rows [][]string) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(dimColor)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0: // header
				return headerStyle
			case col == 0: // first column (ID/title)
				return highlightStyle
			default:
				return lipgloss.NewStyle()
			}
		}).
		Headers(headers...).
		Rows(rows...)
	return t.Render()
}

// CategoryTable renders a table of categories.
func CategoryTable(headers []string, rows [][]string) string {
	return FeedTable(headers, rows)
}

// EntryTable renders a table of entries.
func EntryTable(headers []string, rows [][]string) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(dimColor)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return headerStyle
			case col == 1: // title column - highlight
				return highlightStyle
			default:
				return lipgloss.NewStyle()
			}
		}).
		Headers(headers...).
		Rows(rows...)
	return t.Render()
}

// InfoPanel renders a detail panel for a single item.
func InfoPanel(sections map[string]string) string {
	var b strings.Builder
	b.WriteString(baseStyle.Render(""))
	b.WriteString("\n")

	for label, value := range sections {
		if value == "" {
			continue
		}
		b.WriteString(headerStyle.Render(fmt.Sprintf("  %-16s", label)))
		b.WriteString("  ")
		// Truncate long URLs for display
		if len(value) > 100 {
			value = value[:97] + "..."
		}
		b.WriteString(dimStyle.Render(value))
		b.WriteString("\n")
	}
	return b.String()
}

// Success prints a success message.
func Success(msg string) string {
	return highlightStyle.Render("✓") + " " + msg
}

// ErrorMsg prints an error message.
func ErrorMsg(msg string) string {
	return errorStyle.Render("✗") + " " + msg
}

// Warn prints a warning message.
func Warn(msg string) string {
	return lipgloss.NewStyle().Foreground(warnColor).Render("!") + " " + msg
}
