package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/curkan/billmind/internal/i18n"
)

func (m Model) handleDeleteKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		selected := m.selectedReminder()
		if selected == nil {
			m.viewMode = ViewList
			return m, nil
		}

		// Save for undo
		m.lastAction = &UndoAction{
			Type:     UndoDelete,
			Reminder: *selected,
		}

		// Remove from reminders
		for i := range m.reminders {
			if m.reminders[i].ID == selected.ID {
				m.reminders = append(m.reminders[:i], m.reminders[i+1:]...)
				break
			}
		}

		m.updateListItems()

		if err := m.storage.Save(context.Background(), m.reminders); err != nil {
			m.err = err
		}

		m.viewMode = ViewList
		return m.setStatusMsg(fmt.Sprintf(i18n.T("status.deleted"), selected.Name))

	case "n", "esc":
		m.viewMode = ViewList
		return m, nil
	}

	return m, nil
}
