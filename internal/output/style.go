package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).Bold(true)
)

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
		if len(value) > 100 {
			value = value[:97] + "..."
		}
		b.WriteString(lipgloss.NewStyle().Foreground(dimColor).Render(value))
		b.WriteString("\n")
	}
	return b.String()
}

// Success prints a success message.
func Success(msg string) string {
	return highlightStyle.Render("✓") + " " + msg
}

// Warn prints a warning message.
func Warn(msg string) string {
	return lipgloss.NewStyle().Foreground(warnColor).Render("!") + " " + msg
}

// ErrorMsg prints an error message.
func ErrorMsg(msg string) string {
	return errorStyle.Render("✗") + " " + msg
}
