package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/i18n"
)

// handleWizardUpdate handles all messages when viewMode == ViewWizard.
func (m Model) handleWizardUpdate(msg tea.Msg) (Model, tea.Cmd) {
	if m.wizard == nil {
		m.viewMode = ViewList
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleWizardKeys(msg)
	default:
		cmd := m.updateWizardInputs(msg)
		return m, cmd
	}
}

// updateWizardInputs forwards non-key messages (cursor blink etc.) to focused input.
func (m Model) updateWizardInputs(msg tea.Msg) tea.Cmd {
	w := m.wizard
	var cmd tea.Cmd

	switch w.step {
	case WizardStepInfo:
		switch w.focusIndex {
		case 0:
			w.nameInput, cmd = w.nameInput.Update(msg)
		case 1:
			w.urlInput, cmd = w.urlInput.Update(msg)
		case 2:
			w.tagsInput, cmd = w.tagsInput.Update(msg)
		}
	case WizardStepSchedule:
		if idx := w.customFieldIndex(); idx >= 0 && w.focusIndex == idx {
			w.customInput, cmd = w.customInput.Update(msg)
		} else if w.focusIndex == w.stepFieldCount()-2 {
			w.nextDueInput, cmd = w.nextDueInput.Update(msg)
		} else if w.focusIndex == w.stepFieldCount()-1 {
			w.remindInput, cmd = w.remindInput.Update(msg)
		}
	case WizardStepNotifications:
		// toggles only, no text input
	}
	return cmd
}

// handleWizardKeys is the top-level key dispatcher for the wizard.
func (m Model) handleWizardKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	w := m.wizard
	w.validationErr = ""

	// Global: cancel wizard
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.wizard = nil
		m.viewMode = ViewList
		return m, nil
	}

	// Determine if current field is a text input
	isTextInput := w.isCurrentFieldTextInput()

	// Universal navigation (works on ALL field types)
	switch msg.String() {
	case "tab", "enter", "ctrl+j":
		return m.wizardAdvance()

	case "shift+tab", "ctrl+k":
		return m.wizardGoBack()
	}

	// Non-text-input fields: toggles and selectors
	if !isTextInput {
		return m.handleWizardNonTextKeys(msg)
	}

	// Text input: forward everything else to the input
	return m.forwardToCurrentInput(msg)
}

// handleWizardNonTextKeys handles keys for toggles and selectors.
func (m Model) handleWizardNonTextKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	w := m.wizard

	switch msg.String() {
	// Horizontal: toggle or cycle selector
	case "h", "left", "l", "right", "space":
		return m.handleWizardHorizontal(msg.String())
	// j/k on non-text = next/prev field (convenience)
	case "j", "down":
		return m.wizardAdvance()
	case "k", "up":
		return m.wizardGoBack()
	}

	_ = w
	return m, nil
}

// handleWizardHorizontal handles h/l/←/→/space on toggles and selectors.
func (m Model) handleWizardHorizontal(key string) (Model, tea.Cmd) {
	w := m.wizard

	switch w.step {
	case WizardStepSchedule:
		if w.focusIndex == 0 {
			// Type toggle: recurring ↔ once
			w.isRecurring = !w.isRecurring
			return m, nil
		}
		if w.focusIndex == 1 && w.isRecurring {
			// Interval selector: cycle left/right
			switch key {
			case "h", "left":
				if w.intervalCursor > 0 {
					w.intervalCursor--
					w.interval = intervalOptions[w.intervalCursor]
				}
			case "l", "right":
				if w.intervalCursor < len(intervalOptions)-1 {
					w.intervalCursor++
					w.interval = intervalOptions[w.intervalCursor]
				}
			case "space":
				// space cycles forward
				w.intervalCursor = (w.intervalCursor + 1) % len(intervalOptions)
				w.interval = intervalOptions[w.intervalCursor]
			}
			return m, nil
		}

	case WizardStepNotifications:
		switch w.focusIndex {
		case 0:
			w.notifyMacOS = !w.notifyMacOS
		case 1:
			w.notifyNtfy = !w.notifyNtfy
		}
		return m, nil
	}

	return m, nil
}

// wizardAdvance moves to the next field. If on last field, goes to next step.
func (m Model) wizardAdvance() (Model, tea.Cmd) {
	w := m.wizard

	// Sync numeric fields before leaving schedule step
	if w.step == WizardStepSchedule {
		w.syncNumericFields()
	}

	max := w.stepFieldCount() - 1
	if w.focusIndex < max {
		// Next field within step
		w.focusIndex++
		return m, m.wizardFocusField()
	}

	// Last field → validate and go to next step
	if errMsg := w.validateCurrentStep(); errMsg != "" {
		w.validationErr = errMsg
		return m, nil
	}

	if w.step == WizardStepConfirm {
		return m.wizardSave()
	}

	return m.wizardNextStep()
}

// wizardGoBack moves to the previous field. If on first field, goes to previous step.
func (m Model) wizardGoBack() (Model, tea.Cmd) {
	w := m.wizard
	if w.focusIndex > 0 {
		w.focusIndex--
		return m, m.wizardFocusField()
	}
	// First field → previous step
	return m.wizardPrevStep()
}

// forwardToCurrentInput sends the key to whichever text input is focused.
func (m Model) forwardToCurrentInput(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	w := m.wizard
	var cmd tea.Cmd

	switch w.step {
	case WizardStepInfo:
		switch w.focusIndex {
		case 0:
			w.nameInput, cmd = w.nameInput.Update(msg)
		case 1:
			w.urlInput, cmd = w.urlInput.Update(msg)
		case 2:
			w.tagsInput, cmd = w.tagsInput.Update(msg)
		}
	case WizardStepSchedule:
		if idx := w.customFieldIndex(); idx >= 0 && w.focusIndex == idx {
			w.customInput, cmd = w.customInput.Update(msg)
		} else if w.focusIndex == w.stepFieldCount()-2 {
			w.nextDueInput, cmd = w.nextDueInput.Update(msg)
		} else if w.focusIndex == w.stepFieldCount()-1 {
			w.remindInput, cmd = w.remindInput.Update(msg)
		}
	case WizardStepNotifications:
		// toggles only, no text input
	}

	return m, cmd
}

// ---------------------------------------------------------------------------
// Navigation helpers
// ---------------------------------------------------------------------------

func (m Model) wizardNextStep() (Model, tea.Cmd) {
	w := m.wizard
	w.blurAll()
	w.step++
	w.focusIndex = 0
	return m, m.wizardFocusField()
}

func (m Model) wizardPrevStep() (Model, tea.Cmd) {
	w := m.wizard
	if w.step > WizardStepInfo {
		w.blurAll()
		w.step--
		// Go to last field of previous step
		w.focusIndex = w.stepFieldCount() - 1
		if w.focusIndex < 0 {
			w.focusIndex = 0
		}
		return m, m.wizardFocusField()
	}
	return m, nil
}

func (m Model) wizardFocusField() tea.Cmd {
	w := m.wizard
	w.blurAll()

	switch w.step {
	case WizardStepInfo:
		switch w.focusIndex {
		case 0:
			return w.nameInput.Focus()
		case 1:
			return w.urlInput.Focus()
		case 2:
			return w.tagsInput.Focus()
		}
	case WizardStepSchedule:
		// 0 = type toggle, 1 = interval selector — no text input
		if idx := w.customFieldIndex(); idx >= 0 && w.focusIndex == idx {
			return w.customInput.Focus()
		}
		if w.focusIndex == w.stepFieldCount()-2 {
			return w.nextDueInput.Focus()
		}
		if w.focusIndex == w.stepFieldCount()-1 {
			return w.remindInput.Focus()
		}
	case WizardStepNotifications:
		// toggles only, no text input to focus
	}
	return nil
}

func (w *WizardState) blurAll() {
	w.nameInput.Blur()
	w.urlInput.Blur()
	w.tagsInput.Blur()
	w.customInput.Blur()
	w.nextDueInput.Blur()
	w.remindInput.Blur()
}

// isCurrentFieldTextInput returns true if the focused field is a text input.
func (w *WizardState) isCurrentFieldTextInput() bool {
	switch w.step {
	case WizardStepInfo:
		return true // all fields are text inputs
	case WizardStepSchedule:
		// 0 = type toggle, 1 = interval selector
		if w.focusIndex == 0 || (w.focusIndex == 1 && w.isRecurring) {
			return false
		}
		return true // custom days or remind days
	case WizardStepNotifications:
		return false // toggles only
	case WizardStepConfirm:
		return false
	}
	return false
}

// customFieldIndex returns the index of the custom days field, or -1 if not visible.
func (w *WizardState) customFieldIndex() int {
	if w.isRecurring && w.interval == domain.IntervalCustom {
		return 2
	}
	return -1
}

// ---------------------------------------------------------------------------
// Save
// ---------------------------------------------------------------------------

func (m Model) wizardSave() (Model, tea.Cmd) {
	w := m.wizard

	for _, step := range []WizardStep{WizardStepInfo, WizardStepSchedule, WizardStepNotifications} {
		origStep := w.step
		w.step = step
		if errMsg := w.validateCurrentStep(); errMsg != "" {
			w.validationErr = errMsg
			w.focusIndex = 0
			return m, m.wizardFocusField()
		}
		w.step = origStep
	}

	reminder := w.buildReminder()
	name := strings.TrimSpace(w.nameInput.Value())

	m.reminders = append(m.reminders, reminder)
	m.sortReminders()
	m.updateListItems()

	if err := m.storage.Save(context.Background(), m.reminders); err != nil {
		m.err = err
		m.wizard = nil
		m.viewMode = ViewList
		return m, nil
	}

	m.wizard = nil
	m.viewMode = ViewList
	return m.setStatusMsg(fmt.Sprintf(i18n.T("status.saved"), name))
}

func (m Model) initWizard() (Model, tea.Cmd) {
	existingTags := m.collectAllTags()
	m.wizard = newWizardState(existingTags)
	m.viewMode = ViewWizard
	cmd := m.wizard.nameInput.Focus()
	return m, cmd
}

func parseIntOr(s string, fallback int) int {
	s = strings.TrimSpace(s)
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return fallback
}
