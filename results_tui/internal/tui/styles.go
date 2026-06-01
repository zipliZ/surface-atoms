package tui

import "github.com/charmbracelet/lipgloss"

var (
	focusedBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1, 2)
	panelBorder   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(1, 2)
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	helpStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
)
