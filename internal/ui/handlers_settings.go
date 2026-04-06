package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/curkan/billmind/internal/i18n"
	"github.com/curkan/billmind/internal/storage"
)

// settingsLangs is the ordered list of available languages for the settings screen.
var settingsLangs = []i18n.Lang{i18n.LangRu, i18n.LangEn}

func (m Model) handleSettingsKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "s":
		m.viewMode = ViewList
		return m, nil

	case "j", "down", "ctrl+j":
		if m.settingsCursor < len(settingsLangs)-1 {
			m.settingsCursor++
		}
		return m, nil

	case "k", "up", "ctrl+k":
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
		return m, nil

	case "space", "enter":
		selected := settingsLangs[m.settingsCursor]
		i18n.SetLang(selected)
		// Rebuild table with translated headers and cell content
		m.table.SetColumns(getTableColumns())
		m.updateListItems()
		settings := storage.Settings{
			Language: string(selected),
		}
		if err := m.storage.SaveSettings(context.Background(), settings); err != nil {
			m.err = err
		}
		return m, nil
	}

	return m, nil
}
