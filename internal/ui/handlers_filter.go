package ui

import tea "charm.land/bubbletea/v2"

func (m Model) handleFilterKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down", "ctrl+j":
		if m.filterCursor < len(m.allTags)-1 {
			m.filterCursor++
		}
		return m, nil

	case "k", "up", "ctrl+k":
		if m.filterCursor > 0 {
			m.filterCursor--
		}
		return m, nil

	case "space":
		if m.filterCursor < len(m.allTags) {
			tag := m.allTags[m.filterCursor]
			found := false
			for i, f := range m.pendingFilters {
				if f == tag {
					m.pendingFilters = append(m.pendingFilters[:i], m.pendingFilters[i+1:]...)
					found = true
					break
				}
			}
			if !found {
				m.pendingFilters = append(m.pendingFilters, tag)
			}
		}
		return m, nil

	case "enter":
		m.activeFilters = m.pendingFilters
		m.viewMode = ViewList
		m.updateListItems()
		return m, nil

	case "esc":
		m.viewMode = ViewList
		return m, nil
	}

	return m, nil
}
