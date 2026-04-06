package ui

import "charm.land/lipgloss/v2"

var (
	colorUrgent  = lipgloss.Color("#FF5F5F")
	colorSoon    = lipgloss.Color("#FFCC00")
	colorPaid    = lipgloss.Color("#666666")
	colorSuccess = lipgloss.Color("#50FA7B")
	colorWarning = lipgloss.Color("#FFB86C")
	colorError   = lipgloss.Color("#FF5555")
	colorAccent  = lipgloss.Color("#BD93F9")
	colorHint    = lipgloss.Color("#6272A4")
	colorBorder  = lipgloss.Color("#44475A")
	colorKeyHelp = lipgloss.Color("#8BE9FD")

	styleUrgent  = lipgloss.NewStyle().Foreground(colorUrgent)
	styleSoon    = lipgloss.NewStyle().Foreground(colorSoon)
	stylePaid    = lipgloss.NewStyle().Foreground(colorPaid)
	styleSuccess = lipgloss.NewStyle().Foreground(colorSuccess)
	styleWarning = lipgloss.NewStyle().Foreground(colorWarning)
	styleError   = lipgloss.NewStyle().Foreground(colorError).Bold(true)
	styleAccent  = lipgloss.NewStyle().Foreground(colorAccent)
	styleHint    = lipgloss.NewStyle().Foreground(colorHint)

	styleHelpKey = lipgloss.NewStyle().Foreground(colorKeyHelp).Bold(true)
	styleHelpDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#BFBFBF"))
	styleHelpSep = lipgloss.NewStyle().Foreground(colorBorder)
	styleHelpBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)
)
