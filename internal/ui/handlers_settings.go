package ui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/curkan/billmind/internal/i18n"
	"github.com/curkan/billmind/internal/storage"
)

// settingsLangs is the ordered list of available languages for the settings screen.
var settingsLangs = []i18n.Lang{i18n.LangRu, i18n.LangEn}

const (
	settingsSectionLang = 0
	settingsSectionNtfy = 1
	settingsSectionsMax = 2
)

func (m Model) handleSettingsKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "s":
		m.ntfyTopicInput.Blur()
		// Save ntfy topic on exit
		m.saveSettingsNtfy()
		m.viewMode = ViewList
		return m, nil

	case "tab", "ctrl+j":
		m.ntfyTopicInput.Blur()
		m.settingsSection = (m.settingsSection + 1) % settingsSectionsMax
		if m.settingsSection == settingsSectionNtfy {
			return m, m.ntfyTopicInput.Focus()
		}
		return m, nil

	case "shift+tab", "ctrl+k":
		m.ntfyTopicInput.Blur()
		m.settingsSection--
		if m.settingsSection < 0 {
			m.settingsSection = settingsSectionsMax - 1
		}
		if m.settingsSection == settingsSectionNtfy {
			return m, m.ntfyTopicInput.Focus()
		}
		return m, nil
	}

	// Section-specific keys
	switch m.settingsSection {
	case settingsSectionLang:
		return m.handleSettingsLangKeys(msg)
	case settingsSectionNtfy:
		return m.handleSettingsNtfyKeys(msg)
	}

	return m, nil
}

func (m Model) handleSettingsLangKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.settingsCursor < len(settingsLangs)-1 {
			m.settingsCursor++
		}
	case "k", "up":
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
	case "space", "enter":
		selected := settingsLangs[m.settingsCursor]
		i18n.SetLang(selected)
		m.table.SetColumns(getTableColumns())
		m.updateListItems()
		m.saveSettingsAll(string(selected), strings.TrimSpace(m.ntfyTopicInput.Value()))
	}
	return m, nil
}

func (m Model) handleSettingsNtfyKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Forward all keys to text input
	var cmd tea.Cmd
	m.ntfyTopicInput, cmd = m.ntfyTopicInput.Update(msg)
	return m, cmd
}

func (m Model) saveSettingsNtfy() {
	m.saveSettingsAll(string(i18n.CurrentLang()), strings.TrimSpace(m.ntfyTopicInput.Value()))
}

func (m Model) saveSettingsAll(lang, ntfyTopic string) {
	settings := storage.Settings{
		Language:  lang,
		NtfyTopic: ntfyTopic,
	}
	if err := m.storage.SaveSettings(context.Background(), settings); err != nil {
		m.err = err
	}
}
