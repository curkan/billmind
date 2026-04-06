package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/i18n"
)

func (m Model) handleEditKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if m.edit == nil {
		m.viewMode = ViewList
		return m, nil
	}

	switch msg.String() {
	case "esc":
		m.edit = nil
		m.viewMode = ViewList
		return m, nil

	case "enter":
		return m.saveEdit()

	case "tab", "ctrl+j":
		m.edit.focusIndex++
		if m.edit.focusIndex >= m.edit.visibleFieldCount() {
			m.edit.focusIndex = 0
		}
		m.edit.focusField()
		return m, nil

	case "shift+tab", "ctrl+k":
		m.edit.focusIndex--
		if m.edit.focusIndex < 0 {
			m.edit.focusIndex = m.edit.visibleFieldCount() - 1
		}
		m.edit.focusField()
		return m, nil
	}

	// Dispatch based on actual field
	actual := m.edit.actualField()

	switch actual {
	case editFieldInterval:
		switch msg.String() {
		case "l", "right":
			m.edit.cycleInterval(1)
		case "h", "left":
			m.edit.cycleInterval(-1)
		case "space":
			m.edit.cycleInterval(1)
		}
		return m, nil

	case editFieldNotifyMacOS:
		if msg.String() == "space" || msg.String() == "h" || msg.String() == "l" || msg.String() == "left" || msg.String() == "right" {
			m.edit.notifyMacOS = !m.edit.notifyMacOS
		}
		return m, nil

	case editFieldNotifyEmail:
		if msg.String() == "space" || msg.String() == "h" || msg.String() == "l" || msg.String() == "left" || msg.String() == "right" {
			m.edit.notifyEmail = !m.edit.notifyEmail
		}
		return m, nil

	case editFieldName:
		var cmd tea.Cmd
		m.edit.nameInput, cmd = m.edit.nameInput.Update(msg)
		return m, cmd

	case editFieldURL:
		var cmd tea.Cmd
		m.edit.urlInput, cmd = m.edit.urlInput.Update(msg)
		return m, cmd

	case editFieldTags:
		var cmd tea.Cmd
		m.edit.tagsInput, cmd = m.edit.tagsInput.Update(msg)
		return m, cmd

	case editFieldCustomDays:
		var cmd tea.Cmd
		m.edit.customInput, cmd = m.edit.customInput.Update(msg)
		return m, cmd

	case editFieldNextDue:
		var cmd tea.Cmd
		m.edit.nextDueInput, cmd = m.edit.nextDueInput.Update(msg)
		return m, cmd

	case editFieldRemindDays:
		var cmd tea.Cmd
		m.edit.remindInput, cmd = m.edit.remindInput.Update(msg)
		return m, cmd

	case editFieldEmailAddr:
		var cmd tea.Cmd
		m.edit.emailInput, cmd = m.edit.emailInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) saveEdit() (Model, tea.Cmd) {
	if m.edit == nil {
		m.viewMode = ViewList
		return m, nil
	}

	idx := -1
	for i := range m.reminders {
		if m.reminders[i].ID == m.edit.reminderID {
			idx = i
			break
		}
	}

	if idx == -1 {
		m.edit = nil
		m.viewMode = ViewList
		m.err = fmt.Errorf("reminder not found for editing")
		return m, nil
	}

	// Save snapshot for undo
	m.lastAction = &UndoAction{
		Type:     UndoEdit,
		Reminder: m.reminders[idx],
	}

	// Apply changes
	m.reminders[idx].Name = m.edit.nameInput.Value()
	m.reminders[idx].URL = m.edit.urlInput.Value()
	m.reminders[idx].Tags = parseEditTags(m.edit.tagsInput.Value())
	m.reminders[idx].Interval = m.edit.interval

	if m.edit.interval == domain.IntervalCustom {
		if days, err := strconv.Atoi(m.edit.customInput.Value()); err == nil && days > 0 {
			m.reminders[idx].CustomDays = days
		}
	}

	// Parse next due date
	dateStr := strings.TrimSpace(m.edit.nextDueInput.Value())
	if parsed, err := time.Parse("02.01.2006", dateStr); err == nil {
		m.reminders[idx].NextDue = parsed
	}

	if days, err := strconv.Atoi(m.edit.remindInput.Value()); err == nil && days >= 0 {
		m.reminders[idx].RemindDaysBefore = days
	}

	m.reminders[idx].Notifications.MacOS = m.edit.notifyMacOS
	m.reminders[idx].Notifications.Email = m.edit.notifyEmail

	name := m.reminders[idx].Name

	m.sortReminders()
	m.updateListItems()

	if err := m.storage.Save(context.Background(), m.reminders); err != nil {
		m.err = err
	}

	m.edit = nil
	m.viewMode = ViewList
	return m.setStatusMsg(fmt.Sprintf(i18n.T("status.updated"), name))
}
