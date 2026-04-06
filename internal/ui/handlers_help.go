package ui

import tea "charm.land/bubbletea/v2"

func (m Model) handleHelpKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc", "q":
		m.viewMode = ViewList
		return m, nil
	}

	// Forward scroll keys to viewport
	var cmd tea.Cmd
	m.helpViewport, cmd = m.helpViewport.Update(msg)
	return m, cmd
}
