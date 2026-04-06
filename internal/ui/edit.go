package ui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/i18n"
)

const (
	editFieldName = iota
	editFieldURL
	editFieldTags
	editFieldInterval
	editFieldCustomDays
	editFieldNextDue
	editFieldRemindDays
	editFieldNotifyMacOS
	editFieldNotifyNtfy
)

// EditState holds the state of the edit overlay form.
type EditState struct {
	reminderID    string
	nameInput     textinput.Model
	urlInput      textinput.Model
	tagsInput     textinput.Model
	interval      domain.Interval
	customDays    int
	customInput   textinput.Model
	nextDueInput  textinput.Model
	remindDays    int
	remindInput   textinput.Model
	notifyMacOS bool
	notifyNtfy  bool
	focusIndex  int
}

var allIntervals = []domain.Interval{
	domain.IntervalWeekly,
	domain.IntervalMonthly,
	domain.IntervalYearly,
	domain.IntervalOnce,
	domain.IntervalCustom,
}

// newEditState creates an EditState pre-filled with values from the given reminder.
func newEditState(r *domain.Reminder) *EditState {
	nameIn := textinput.New()
	nameIn.Placeholder = i18n.T("wizard.name_placeholder")
	nameIn.SetValue(r.Name)
	nameIn.CharLimit = 100
	nameIn.SetWidth(40)
	nameIn.Focus()

	urlIn := textinput.New()
	urlIn.Placeholder = "https://..."
	urlIn.SetValue(r.URL)
	urlIn.CharLimit = 200
	urlIn.SetWidth(40)

	tagsIn := textinput.New()
	tagsIn.Placeholder = "tag1, tag2"
	tagsIn.SetValue(strings.Join(r.Tags, ", "))
	tagsIn.CharLimit = 200
	tagsIn.SetWidth(40)

	customIn := textinput.New()
	customIn.Placeholder = "14"
	customIn.CharLimit = 5
	customIn.SetWidth(10)
	if r.CustomDays > 0 {
		customIn.SetValue(strconv.Itoa(r.CustomDays))
	}

	nextDueIn := textinput.New()
	nextDueIn.Placeholder = i18n.T("wizard.next_due_hint")
	nextDueIn.SetValue(r.NextDue.Format("02.01.2006"))
	nextDueIn.CharLimit = 10
	nextDueIn.SetWidth(12)

	remindIn := textinput.New()
	remindIn.Placeholder = "3"
	remindIn.SetValue(strconv.Itoa(r.RemindDaysBefore))
	remindIn.CharLimit = 5
	remindIn.SetWidth(10)

	return &EditState{
		reminderID:   r.ID,
		nameInput:    nameIn,
		urlInput:     urlIn,
		tagsInput:    tagsIn,
		interval:     r.Interval,
		customDays:   r.CustomDays,
		customInput:  customIn,
		nextDueInput: nextDueIn,
		remindDays:   r.RemindDaysBefore,
		remindInput:  remindIn,
		notifyMacOS:  r.Notifications.MacOS,
		notifyNtfy:   r.Notifications.Ntfy,
		focusIndex:   editFieldName,
	}
}

// cycleInterval moves to the next or previous interval in the list.
func (e *EditState) cycleInterval(delta int) {
	idx := 0
	for i, iv := range allIntervals {
		if iv == e.interval {
			idx = i
			break
		}
	}
	idx += delta
	if idx < 0 {
		idx = len(allIntervals) - 1
	}
	if idx >= len(allIntervals) {
		idx = 0
	}
	e.interval = allIntervals[idx]
}

// focusField blurs all inputs then focuses the one at focusIndex.
// visibleFieldCount returns how many fields are currently visible.
func (e *EditState) visibleFieldCount() int {
	n := 8 // name, url, tags, interval, nextDue, remind, macOS, ntfy toggle
	if e.interval == domain.IntervalCustom {
		n++ // custom days
	}
	return n
}

func (e *EditState) focusField() {
	e.nameInput.Blur()
	e.urlInput.Blur()
	e.tagsInput.Blur()
	e.customInput.Blur()
	e.nextDueInput.Blur()
	e.remindInput.Blur()

	actual := e.actualField()
	switch actual {
	case editFieldName:
		e.nameInput.Focus()
	case editFieldURL:
		e.urlInput.Focus()
	case editFieldTags:
		e.tagsInput.Focus()
	case editFieldCustomDays:
		e.customInput.Focus()
	case editFieldNextDue:
		e.nextDueInput.Focus()
	case editFieldRemindDays:
		e.remindInput.Focus()
	}
}

// actualField maps focusIndex to the real field, skipping hidden fields.
func (e *EditState) actualField() int {
	idx := e.focusIndex
	// Fields: name(0), url(1), tags(2), interval(3), [customDays(4)], nextDue, remindDays, macOS, ntfy
	if e.interval != domain.IntervalCustom && idx >= editFieldCustomDays {
		idx++ // skip customDays
	}
	return idx
}

// renderEditForm builds the edit form string.
func (e *EditState) renderEditForm(width int) string {
	formWidth := 56
	if width < formWidth+6 {
		formWidth = width - 6
	}
	if formWidth < 30 {
		formWidth = 30
	}

	labelStyle := lipgloss.NewStyle().Width(22).Foreground(colorHint)
	activeLabel := lipgloss.NewStyle().Width(22).Foreground(colorAccent).Bold(true)
	toggleOn := lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	toggleOff := lipgloss.NewStyle().Foreground(colorHint)

	actual := e.actualField()

	label := func(field int, text string) string {
		if actual == field {
			return activeLabel.Render(text)
		}
		return labelStyle.Render(text)
	}

	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Render(i18n.T("edit.title"))
	b.WriteString(title + "\n\n")

	// Name
	b.WriteString(label(editFieldName, i18n.T("wizard.name")+":") + " " + e.nameInput.View() + "\n")
	// URL
	b.WriteString(label(editFieldURL, i18n.T("wizard.url")+":") + " " + e.urlInput.View() + "\n")
	// Tags
	b.WriteString(label(editFieldTags, i18n.T("wizard.tags")+":") + " " + e.tagsInput.View() + "\n")
	// Interval
	intervalDisplay := fmt.Sprintf("< %s >", e.interval.String())
	if actual == editFieldInterval {
		intervalDisplay = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(intervalDisplay)
	}
	b.WriteString(label(editFieldInterval, i18n.T("wizard.interval")+":") + " " + intervalDisplay + "\n")
	// Custom days (shown only for custom interval)
	if e.interval == domain.IntervalCustom {
		b.WriteString(label(editFieldCustomDays, i18n.T("wizard.custom_days")+":") + " " + e.customInput.View() + "\n")
	}
	// Next due date
	b.WriteString(label(editFieldNextDue, i18n.T("wizard.next_due")+":") + " " + e.nextDueInput.View() + "\n")
	// Remind days
	b.WriteString(label(editFieldRemindDays, i18n.T("wizard.remind_before")+" ("+i18n.T("wizard.days")+"):") + " " + e.remindInput.View() + "\n")
	// macOS notifications
	var macToggle string
	if e.notifyMacOS {
		macToggle = toggleOn.Render("[x]")
	} else {
		macToggle = toggleOff.Render("[ ]")
	}
	b.WriteString(label(editFieldNotifyMacOS, i18n.T("wizard.macos_notify")+":") + " " + macToggle + "\n")
	// ntfy notifications
	var ntfyToggle string
	if e.notifyNtfy {
		ntfyToggle = toggleOn.Render("[x]")
	} else {
		ntfyToggle = toggleOff.Render("[ ]")
	}
	b.WriteString(label(editFieldNotifyNtfy, i18n.T("wizard.ntfy_notify")+":") + " " + ntfyToggle + "\n")

	b.WriteString("\n")
	b.WriteString(formatKeyHelp(
		"tab", i18n.T("edit.next_field"),
		"h/l", "←/→",
		"enter", i18n.T("edit.save"),
		"esc", i18n.T("edit.cancel"),
	))

	return b.String()
}

// parseEditTags splits a comma-separated string into trimmed tags, filtering empty.
func parseEditTags(s string) []string {
	parts := strings.Split(s, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
