package ui

import tea "charm.land/bubbletea/v2"

func (m Model) handlePaidKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		return m.confirmPaid()
	case "n", "esc":
		m.viewMode = ViewList
		return m, nil
	}
	return m, nil
}
