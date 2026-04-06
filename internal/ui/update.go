package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/i18n"
)

type clearStatusMsg struct{}
type resetGPendingMsg struct{}
type resetDPendingMsg struct{}

// getHelpContent returns the translated help text for the help viewport.
func getHelpContent() string {
	return fmt.Sprintf(`%s
  j / k          %s / %s
  gg             %s
  G              %s

%s
  a              %s
  e              %s
  dd             %s
  Space          %s
  u              %s
  o              %s

%s
  /              %s
  f              %s
  Esc            %s

%s
  s              %s
  ?              %s
  q / Ctrl+C     %s`,
		i18n.T("help_screen.navigation"),
		i18n.T("help_screen.row_down"), i18n.T("help_screen.row_up"),
		i18n.T("help_screen.go_top"),
		i18n.T("help_screen.go_bottom"),
		i18n.T("help_screen.actions"),
		i18n.T("help_screen.add_reminder"),
		i18n.T("help_screen.edit_selected"),
		i18n.T("help_screen.delete_selected"),
		i18n.T("help_screen.mark_paid"),
		i18n.T("help_screen.undo_last"),
		i18n.T("help_screen.open_url"),
		i18n.T("help_screen.search_filter"),
		i18n.T("help_screen.search"),
		i18n.T("help_screen.filter_tags"),
		i18n.T("help_screen.reset_filter"),
		i18n.T("help_screen.general"),
		i18n.T("help_screen.settings"),
		i18n.T("help_screen.toggle_help"),
		i18n.T("help_screen.quit"),
	)
}

// helpMaxWidth is the maximum width of the help overlay box.
const helpMaxWidth = 60

// initHelpViewport creates and configures the help viewport sized to terminal dimensions.
func (m *Model) initHelpViewport() {
	content := getHelpContent()
	contentLines := strings.Count(content, "\n") + 1
	// Border (2) + padding top/bottom (2) = 4 lines overhead
	borderOverhead := 4
	maxHeight := m.height - 4
	vpHeight := contentLines
	if vpHeight > maxHeight-borderOverhead {
		vpHeight = maxHeight - borderOverhead
	}
	if vpHeight < 1 {
		vpHeight = 1
	}

	// Inner content width = max width - border (2) - padding left/right (6)
	vpWidth := helpMaxWidth - 8
	if vpWidth < 20 {
		vpWidth = 20
	}

	vp := viewport.New(
		viewport.WithWidth(vpWidth),
		viewport.WithHeight(vpHeight),
	)
	vp.SetContent(content)
	m.helpViewport = vp
}

// Update is the central message dispatcher following the Elm Architecture.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Reserve: 4 lines header (title + blank + table header + separator)
		// + 5 lines footer (status/filter/search + help bar with border + padding)
		tableHeight := msg.Height - 9
		if tableHeight < 3 {
			tableHeight = 3
		}
		m.table.SetHeight(tableHeight)
		m.table.SetWidth(msg.Width - 4)
		// Re-init help viewport if currently viewing help
		if m.viewMode == ViewHelp {
			m.initHelpViewport()
		}
		return m, nil

	case loadRemindersMsg:
		reminders, err := m.storage.Load(context.Background())
		if err != nil {
			m.err = err
			return m, nil
		}
		m.reminders = reminders
		m.sortReminders()
		m.updateListItems()
		return m, nil

	case settingsLoadMsg:
		settings, err := m.storage.LoadSettings(context.Background())
		if err != nil {
			// Non-blocking: use system detection as fallback
			i18n.SetLang(i18n.DetectSystemLang())
			return m, nil
		}
		return m, func() tea.Msg {
			return settingsLoadedMsg{settings: settings}
		}

	case settingsLoadedMsg:
		if msg.settings.Language != "" {
			i18n.SetLang(i18n.Lang(msg.settings.Language))
		} else {
			i18n.SetLang(i18n.DetectSystemLang())
		}
		// Update table columns with translated headers
		m.table.SetColumns(getTableColumns())
		return m, nil

	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil

	case resetGPendingMsg:
		m.gPending = false
		return m, nil

	case resetDPendingMsg:
		m.dPending = false
		return m, nil

	case tea.KeyPressMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "?":
			if m.viewMode == ViewHelp {
				m.viewMode = ViewList
			} else if m.viewMode == ViewList {
				m.viewMode = ViewHelp
				m.initHelpViewport()
			}
			return m, nil
		}

		// Delegate to view-specific handler
		switch m.viewMode {
		case ViewList:
			return m.handleListKeys(msg)
		case ViewWizard:
			return m.handleWizardUpdate(msg)
		case ViewEdit:
			return m.handleEditKeys(msg)
		case ViewDelete:
			return m.handleDeleteKeys(msg)
		case ViewConfirmPaid:
			return m.handlePaidKeys(msg)
		case ViewFilter:
			return m.handleFilterKeys(msg)
		case ViewHelp:
			return m.handleHelpKeys(msg)
		case ViewSettings:
			return m.handleSettingsKeys(msg)
		}
	}

	// Forward non-key messages to wizard (cursor blink, etc.)
	if m.viewMode == ViewWizard && m.wizard != nil {
		return m.handleWizardUpdate(msg)
	}

	// Forward non-key messages to edit inputs (cursor blink, etc.)
	if m.viewMode == ViewEdit && m.edit != nil {
		var cmd tea.Cmd
		actual := m.edit.actualField()
		switch actual {
		case editFieldName:
			m.edit.nameInput, cmd = m.edit.nameInput.Update(msg)
		case editFieldURL:
			m.edit.urlInput, cmd = m.edit.urlInput.Update(msg)
		case editFieldTags:
			m.edit.tagsInput, cmd = m.edit.tagsInput.Update(msg)
		case editFieldCustomDays:
			m.edit.customInput, cmd = m.edit.customInput.Update(msg)
		case editFieldNextDue:
			m.edit.nextDueInput, cmd = m.edit.nextDueInput.Update(msg)
		case editFieldRemindDays:
			m.edit.remindInput, cmd = m.edit.remindInput.Update(msg)
		case editFieldEmailAddr:
			m.edit.emailInput, cmd = m.edit.emailInput.Update(msg)
		}
		return m, cmd
	}

	// Forward non-key messages to help viewport when in ViewHelp
	if m.viewMode == ViewHelp {
		var cmd tea.Cmd
		m.helpViewport, cmd = m.helpViewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) sortReminders() {
	sort.Slice(m.reminders, func(i, j int) bool {
		return m.reminders[i].NextDue.Before(m.reminders[j].NextDue)
	})
}

func (m *Model) updateListItems() {
	filtered := m.filteredReminders()
	rows := make([]TableRow, len(filtered))
	for i, r := range filtered {
		rows[i] = reminderToRow(r)
	}
	m.table.SetRows(rows)
}

// reminderToRow converts a domain.Reminder into a TableRow with per-row style.
func reminderToRow(r domain.Reminder) TableRow {
	// Row-level style (applied to entire line as one Render call)
	var rowStyle lipgloss.Style
	switch {
	case r.IsOverdue() || r.DaysUntilDue() == 0:
		rowStyle = styleUrgent
	case r.IsDueSoon(r.RemindDaysBefore):
		rowStyle = styleSoon
	default:
		rowStyle = lipgloss.NewStyle()
	}

	// Status symbol
	var symbol string
	switch {
	case r.IsOverdue() || r.DaysUntilDue() == 0:
		symbol = "✸"
	case r.IsDueSoon(r.RemindDaysBefore):
		symbol = "●"
	default:
		symbol = "○"
	}

	// Date text
	days := r.DaysUntilDue()
	var dateText string
	switch {
	case days < 0:
		dateText = fmt.Sprintf(i18n.T("list.overdue"), -days)
	case days == 0:
		dateText = i18n.T("list.today")
	case days == 1:
		dateText = i18n.T("list.tomorrow")
	default:
		dateText = fmt.Sprintf(i18n.T("list.in_days"), days)
	}

	// Tags
	tags := ""
	if len(r.Tags) > 0 {
		parts := make([]string, len(r.Tags))
		for i, t := range r.Tags {
			parts[i] = "#" + t
		}
		tags = strings.Join(parts, " ")
	}

	// Link indicator
	link := ""
	if r.URL != "" {
		link = "🔗"
	}

	return TableRow{
		Cells: []string{symbol, r.Name, dateText, r.Interval.String(), tags, link},
		Style: rowStyle,
	}
}

func (m Model) filteredReminders() []domain.Reminder {
	result := m.reminders

	// Filter by tags
	if len(m.activeFilters) > 0 {
		var filtered []domain.Reminder
		for _, r := range result {
			for _, tag := range r.Tags {
				for _, f := range m.activeFilters {
					if tag == f {
						filtered = append(filtered, r)
						goto next
					}
				}
			}
		next:
		}
		result = filtered
	}

	// Filter by search query
	if query := m.searchInput.Value(); query != "" && (m.searchOpen || m.searchActive) {
		var filtered []domain.Reminder
		for _, r := range result {
			if matchesQuery(r, query) {
				filtered = append(filtered, r)
			}
		}
		result = filtered
	}

	return result
}

func matchesQuery(r domain.Reminder, query string) bool {
	q := strings.ToLower(query)
	if strings.Contains(strings.ToLower(r.Name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(r.URL), q) {
		return true
	}
	for _, tag := range r.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

func (m Model) setStatusMsg(msg string) (Model, tea.Cmd) {
	m.statusMsg = msg
	return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}
