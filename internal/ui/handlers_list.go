package ui

import (
	"context"
	"fmt"
	"sort"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/i18n"
)

func (m Model) handleListKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// If search is open, handle search input
	if m.searchOpen {
		return m.handleSearchKeys(msg)
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "a":
		return m.initWizard()

	case "e":
		selected := m.selectedReminder()
		if selected == nil {
			return m, nil
		}
		m.edit = newEditState(selected)
		m.viewMode = ViewEdit
		return m, nil

	case "j", "down", "ctrl+j":
		m.table.MoveDown(1)
		return m, nil

	case "k", "up", "ctrl+k":
		m.table.MoveUp(1)
		return m, nil

	case "g":
		if m.gPending {
			m.gPending = false
			m.table.GotoTop()
			return m, nil
		}
		m.gPending = true
		return m, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
			return resetGPendingMsg{}
		})

	case "G":
		m.table.GotoBottom()
		return m, nil

	case "d":
		if m.dPending {
			m.dPending = false
			m.viewMode = ViewDelete
			return m, nil
		}
		m.dPending = true
		return m, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
			return resetDPendingMsg{}
		})

	case "space":
		if m.selectedReminder() != nil {
			m.viewMode = ViewConfirmPaid
		}
		return m, nil

	case "o":
		return m.openURL()

	case "u":
		return m.undo()

	case "/":
		m.searchOpen = true
		cmd := m.searchInput.Focus()
		return m, cmd

	case "f":
		m.viewMode = ViewFilter
		m.allTags = m.collectAllTags()
		m.pendingFilters = make([]string, len(m.activeFilters))
		copy(m.pendingFilters, m.activeFilters)
		m.filterCursor = 0
		return m, nil

	case "s":
		m.viewMode = ViewSettings
		m.settingsCursor = 0
		m.settingsSection = settingsSectionLang
		m.ntfyTopicInput.Blur()
		return m, nil

	case "esc":
		if m.searchActive {
			m.searchActive = false
			m.searchInput.SetValue("")
			m.updateListItems()
			return m, nil
		}
		if len(m.activeFilters) > 0 {
			m.activeFilters = nil
			m.updateListItems()
			return m, nil
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleSearchKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.searchOpen = false
		m.searchActive = true
		m.searchInput.Blur()
		return m, nil

	case "esc":
		m.searchOpen = false
		m.searchActive = false
		m.searchInput.SetValue("")
		m.searchInput.Blur()
		m.updateListItems()
		return m, nil
	}

	// Update search input
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.updateListItems()
	return m, cmd
}

func (m Model) confirmPaid() (Model, tea.Cmd) {
	selected := m.selectedReminder()
	if selected == nil {
		m.viewMode = ViewList
		return m, nil
	}

	// Save for undo
	m.lastAction = &UndoAction{
		Type:     UndoPaid,
		Reminder: *selected,
	}

	name := selected.Name

	if selected.Interval == domain.IntervalOnce {
		// One-time: remove from list
		for i := range m.reminders {
			if m.reminders[i].ID == selected.ID {
				m.reminders = append(m.reminders[:i], m.reminders[i+1:]...)
				break
			}
		}
		m.sortReminders()
		m.updateListItems()

		if err := m.storage.Save(context.Background(), m.reminders); err != nil {
			m.err = err
		}

		m.viewMode = ViewList
		return m.setStatusMsg(fmt.Sprintf(i18n.T("status.paid_once"), name))
	}

	// Recurring: advance NextDue
	nextDue := calcNextDue(selected.NextDue, selected.Interval, selected.CustomDays)
	selected.NextDue = nextDue
	selected.PaidAt = nil // reset paid status
	selected.ResetNotifyStage()

	m.sortReminders()
	m.updateListItems()

	if err := m.storage.Save(context.Background(), m.reminders); err != nil {
		m.err = err
	}

	m.viewMode = ViewList
	return m.setStatusMsg(fmt.Sprintf(i18n.T("status.paid_done"), name, nextDue.Format("02.01.2006")))
}

func (m Model) openURL() (Model, tea.Cmd) {
	selected := m.selectedReminder()
	if selected == nil || selected.URL == "" {
		return m, nil
	}

	err := m.platform.OpenURL(context.Background(), selected.URL)
	if err != nil {
		m.err = err
		return m, nil
	}
	return m, nil
}

func (m Model) undo() (Model, tea.Cmd) {
	if m.lastAction == nil {
		return m, nil
	}

	action := m.lastAction
	m.lastAction = nil

	switch action.Type {
	case UndoDelete:
		m.reminders = append(m.reminders, action.Reminder)
	case UndoPaid:
		// For once items that were removed, re-add
		found := false
		for i := range m.reminders {
			if m.reminders[i].ID == action.Reminder.ID {
				m.reminders[i] = action.Reminder
				found = true
				break
			}
		}
		if !found {
			m.reminders = append(m.reminders, action.Reminder)
		}
	case UndoEdit:
		for i := range m.reminders {
			if m.reminders[i].ID == action.Reminder.ID {
				m.reminders[i] = action.Reminder
				break
			}
		}
	}

	m.sortReminders()
	m.updateListItems()

	if err := m.storage.Save(context.Background(), m.reminders); err != nil {
		m.err = err
		return m, nil
	}

	return m.setStatusMsg(i18n.T("status.undo"))
}

func (m Model) collectAllTags() []string {
	tagSet := make(map[string]bool)
	for _, r := range m.reminders {
		for _, t := range r.Tags {
			tagSet[t] = true
		}
	}
	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	return tags
}

// calcNextDue advances a due date from the current NextDue by one interval period.
// Keeps advancing until the result is strictly after today.
// This anchors to the original schedule (e.g. 1st of month stays 1st of month).
func calcNextDue(current time.Time, interval domain.Interval, customDays int) time.Time {
	today := time.Now().Truncate(24 * time.Hour)
	next := current
	for {
		switch interval {
		case domain.IntervalWeekly:
			next = next.AddDate(0, 0, 7)
		case domain.IntervalMonthly:
			next = next.AddDate(0, 1, 0)
		case domain.IntervalYearly:
			next = next.AddDate(1, 0, 0)
		case domain.IntervalCustom:
			if customDays <= 0 {
				customDays = 30
			}
			next = next.AddDate(0, 0, customDays)
		default:
			next = next.AddDate(0, 1, 0)
		}
		if next.Truncate(24 * time.Hour).After(today) {
			break
		}
	}
	return next
}
