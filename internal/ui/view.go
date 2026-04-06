package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/i18n"
)

// View renders the current screen based on viewMode. Pure function, no side effects.
func (m Model) View() tea.View {
	var content string

	switch m.viewMode {
	case ViewList:
		content = m.viewList()
	case ViewWizard:
		content = m.viewWizard()
	case ViewEdit:
		content = m.viewEdit()
	case ViewDelete:
		content = m.viewDelete()
	case ViewConfirmPaid:
		content = m.viewConfirmPaid()
	case ViewFilter:
		content = m.viewFilter()
	case ViewHelp:
		content = m.viewHelp()
	case ViewSettings:
		content = m.viewSettings()
	default:
		content = m.viewList()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m Model) viewList() string {
	// Header
	var content strings.Builder

	title := styleAccent.Bold(true).Render("billmind")
	itemCount := len(m.table.Rows())

	if itemCount > 0 {
		counter := styleHint.Render(fmt.Sprintf(i18n.T("list.count"), itemCount))
		content.WriteString(title + "  " + counter + "\n\n")
		content.WriteString(m.table.View())
	} else {
		content.WriteString(title + "\n\n")
		content.WriteString(styleHint.Render("  " + i18n.T("list.no_items")) + "\n")
		content.WriteString(styleHint.Render("  " + i18n.T("list.empty_hint")) + "\n")
	}

	// Footer parts
	var footer strings.Builder

	// Filter indicator
	if len(m.activeFilters) > 0 {
		tags := ""
		for _, t := range m.activeFilters {
			tags += fmt.Sprintf(" [%s]", t)
		}
		footer.WriteString(styleHint.Render(tags+"   f "+i18n.T("filter.change")+"   esc "+i18n.T("filter.reset")) + "\n")
	}

	// Search indicator (when fixed but not open)
	if m.searchActive && !m.searchOpen {
		footer.WriteString(styleHint.Render(fmt.Sprintf("  /%s   esc %s", m.searchInput.Value(), i18n.T("search.reset"))) + "\n")
	}

	// Search input (when open)
	if m.searchOpen {
		footer.WriteString(m.searchInput.View() + "\n")
	}

	// Status message
	if m.statusMsg != "" {
		footer.WriteString("  " + styleSuccess.Render(m.statusMsg) + "\n")
	}

	// Error
	if m.err != nil {
		footer.WriteString("  " + styleError.Render(fmt.Sprintf("✗ %v", m.err)) + "\n")
	}

	// Help bar
	footer.WriteString(m.renderHelpBar())

	// Compose: content at top, footer pinned to bottom
	main := content.String()
	return lipgloss.JoinVertical(lipgloss.Left,
		main,
		lipgloss.PlaceVertical(m.height-lipgloss.Height(main), lipgloss.Bottom, footer.String()),
	)
}

func formatKeyHelp(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, styleHelpKey.Render(pairs[i])+" "+styleHelpDesc.Render(pairs[i+1]))
	}
	sep := " " + styleHelpSep.Render("•") + " "
	inner := strings.Join(parts, sep)
	return styleHelpBox.Render(inner)
}

func (m Model) renderHelpBar() string {
	hasItems := len(m.table.Rows()) > 0

	pairs := []string{"a", i18n.T("help.add")}

	if hasItems {
		pairs = append(pairs,
			"e", i18n.T("help.edit"),
			"dd", i18n.T("help.delete"),
			"space", i18n.T("help.paid"),
			"o", i18n.T("help.open"),
		)
	}

	pairs = append(pairs,
		"/", i18n.T("help.search"),
		"f", i18n.T("help.filter"),
		"u", i18n.T("help.undo"),
		"s", i18n.T("help.settings"),
		"?", i18n.T("help.help"),
		"q", i18n.T("help.quit"),
	)

	return formatKeyHelp(pairs...)
}

func (m Model) viewWizard() string {
	if m.wizard == nil {
		return m.viewList()
	}
	return renderWizard(m.wizard, m.width, m.height)
}

func (m Model) viewEdit() string {
	if m.edit == nil {
		return m.viewList()
	}

	// Render the list view as dimmed background
	listView := m.viewList()
	dimmed := lipgloss.NewStyle().Faint(true).Render(listView)

	// Render the edit form
	formContent := m.edit.renderEditForm(m.width)

	framedForm := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 3).
		Render(formContent)

	return overlayCenter(dimmed, framedForm, m.width, m.height)
}

func (m Model) viewDelete() string {
	selected := m.selectedReminder()
	name := "?"
	if selected != nil {
		name = selected.Name
	}

	// Render the list view as dimmed background
	listView := m.viewList()
	dimmed := lipgloss.NewStyle().Faint(true).Render(listView)

	// Create the delete confirmation modal
	hint := formatKeyHelp("y", i18n.T("delete.yes"), "n/Esc", i18n.T("delete.no"))
	modalContent := fmt.Sprintf(i18n.T("delete.confirm"), name) + "\n\n" + hint

	modal := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(colorWarning).
		Padding(1, 3).
		Render(modalContent)

	return overlayCenter(dimmed, modal, m.width, m.height)
}

func (m Model) viewConfirmPaid() string {
	selected := m.selectedReminder()
	name := "?"
	if selected == nil {
		return m.viewList()
	}
	name = selected.Name

	listView := m.viewList()
	dimmed := lipgloss.NewStyle().Faint(true).Render(listView)

	// Build modal content
	var info string
	if selected.Interval == domain.IntervalOnce {
		info = styleHint.Render(i18n.T("paid.will_remove"))
	} else {
		nextDue := calcNextDue(selected.NextDue, selected.Interval, selected.CustomDays)
		info = styleHint.Render(fmt.Sprintf(i18n.T("paid.next_due"), nextDue.Format("02.01.2006")))
	}

	hint := formatKeyHelp("y", i18n.T("paid.yes"), "n/Esc", i18n.T("paid.no"))
	modalContent := fmt.Sprintf(i18n.T("paid.confirm"), name) + "\n" + info + "\n\n" + hint

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSuccess).
		Padding(1, 3).
		Render(modalContent)

	return overlayCenter(dimmed, modal, m.width, m.height)
}

func (m Model) viewFilter() string {
	// Render the list view as dimmed background
	listView := m.viewList()
	dimmed := lipgloss.NewStyle().Faint(true).Render(listView)

	// Build the filter box content
	var b strings.Builder
	for i, tag := range m.allTags {
		cursor := "  "
		if i == m.filterCursor {
			cursor = "> "
		}
		checked := "[ ]"
		for _, f := range m.pendingFilters {
			if f == tag {
				checked = "[x]"
				break
			}
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checked, tag))
	}
	b.WriteString("\n")
	b.WriteString(formatKeyHelp(
		"j/k", i18n.T("filter.navigate"),
		"space", i18n.T("filter.toggle"),
		"enter", i18n.T("filter.apply"),
		"esc", i18n.T("filter.close"),
	))

	filterBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 2).
		Render(b.String())

	title := styleAccent.Bold(true).Render(i18n.T("filter.title"))
	result := title + "\n" + filterBox
	return overlayCenter(dimmed, result, m.width, m.height)
}

func (m Model) viewHelp() string {
	vpView := m.helpViewport.View()

	// Footer with help hints
	footer := formatKeyHelp("?/esc", i18n.T("help_screen.close"))
	if m.helpViewport.TotalLineCount() > m.helpViewport.VisibleLineCount() {
		pct := fmt.Sprintf("%d%%", int(m.helpViewport.ScrollPercent()*100))
		footer = formatKeyHelp("↑/↓", pct, "?/esc", i18n.T("help_screen.close"))
	}

	innerContent := vpView + "\n" + footer

	title := styleAccent.Bold(true).Render(i18n.T("help_screen.title"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 3).
		Width(helpMaxWidth).
		Render(innerContent)

	result := title + "\n" + box
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, result)
}

func (m Model) viewSettings() string {
	// Render the list view as dimmed background
	listView := m.viewList()
	dimmed := lipgloss.NewStyle().Faint(true).Render(listView)

	currentLang := i18n.CurrentLang()
	sectionLabel := lipgloss.NewStyle().Bold(true)
	activeSectionLabel := sectionLabel.Foreground(colorAccent)
	dimSectionLabel := sectionLabel.Foreground(colorHint)

	var b strings.Builder

	// Section 1: Language
	if m.settingsSection == settingsSectionLang {
		b.WriteString(activeSectionLabel.Render(i18n.T("settings.language")) + "\n\n")
	} else {
		b.WriteString(dimSectionLabel.Render(i18n.T("settings.language")) + "\n\n")
	}

	for idx, lang := range settingsLangs {
		cursor := "  "
		if m.settingsSection == settingsSectionLang && idx == m.settingsCursor {
			cursor = "> "
		}

		var label string
		if lang == i18n.LangRu {
			label = i18n.T("settings.russian")
		} else {
			label = i18n.T("settings.english")
		}

		if lang == currentLang {
			label = styleAccent.Bold(true).Render("[" + label + "]")
		} else {
			label = styleHint.Render(label)
		}

		b.WriteString(cursor + label + "\n")
	}

	// Section 2: ntfy topic
	b.WriteString("\n")
	if m.settingsSection == settingsSectionNtfy {
		b.WriteString(activeSectionLabel.Render(i18n.T("settings.ntfy_topic")) + "\n\n")
	} else {
		b.WriteString(dimSectionLabel.Render(i18n.T("settings.ntfy_topic")) + "\n\n")
	}
	b.WriteString("  " + m.ntfyTopicInput.View() + "\n")
	b.WriteString("  " + styleHint.Render(i18n.T("settings.ntfy_hint")) + "\n")

	b.WriteString("\n")
	b.WriteString(formatKeyHelp(
		"tab", i18n.T("settings.navigate"),
		"space", i18n.T("settings.select"),
		"esc", i18n.T("settings.close"),
	))

	title := styleAccent.Bold(true).Render(i18n.T("settings.title"))

	settingsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 3).
		Render(b.String())

	result := title + "\n" + settingsBox
	return overlayCenter(dimmed, result, m.width, m.height)
}

func (m Model) selectedReminder() *domain.Reminder {
	filtered := m.filteredReminders()
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(filtered) {
		return nil
	}
	id := filtered[cursor].ID
	for i := range m.reminders {
		if m.reminders[i].ID == id {
			return &m.reminders[i]
		}
	}
	return nil
}
