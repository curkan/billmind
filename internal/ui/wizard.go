package ui

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/i18n"
)

// WizardStep identifies the current page of the add-reminder wizard.
type WizardStep int

const (
	WizardStepInfo          WizardStep = iota // Name, URL, Tags
	WizardStepSchedule                        // Recurring/Once, Interval, Remind days
	WizardStepNotifications                   // macOS, Email toggles
	WizardStepConfirm                         // Read-only summary
)

const wizardTotalSteps = 4

// intervalOptions lists the selectable intervals in the schedule step.
var intervalOptions = []domain.Interval{
	domain.IntervalWeekly,
	domain.IntervalMonthly,
	domain.IntervalYearly,
	domain.IntervalCustom,
}

// WizardState holds all mutable state for the add-reminder wizard.
type WizardState struct {
	step WizardStep

	// Step 1 -- Info
	nameInput textinput.Model
	urlInput  textinput.Model
	tagsInput textinput.Model

	// Step 2 -- Schedule
	isRecurring    bool
	interval       domain.Interval
	intervalCursor int
	customDays     int
	customInput    textinput.Model
	nextDueInput   textinput.Model
	remindDays     int
	remindInput    textinput.Model

	// Step 3 -- Notifications
	notifyMacOS bool
	notifyEmail bool
	emailInput  textinput.Model

	// Per-step focused field index
	focusIndex int

	// Validation error shown below the current field
	validationErr string
}

// newWizardState creates a fresh wizard with all text inputs configured.
func newWizardState(existingTags []string) *WizardState {
	nameIn := textinput.New()
	nameIn.Placeholder = i18n.T("wizard.name_placeholder")
	nameIn.CharLimit = 80
	nameIn.SetWidth(40)

	urlIn := textinput.New()
	urlIn.Placeholder = i18n.T("wizard.url_placeholder")
	urlIn.CharLimit = 256
	urlIn.SetWidth(40)

	tagsIn := textinput.New()
	tagsIn.Placeholder = i18n.T("wizard.tags_placeholder")
	tagsIn.CharLimit = 120
	tagsIn.SetWidth(40)

	customIn := textinput.New()
	customIn.Placeholder = "14"
	customIn.CharLimit = 5
	customIn.SetWidth(10)

	nextDueIn := textinput.New()
	nextDueIn.Placeholder = i18n.T("wizard.next_due_hint")
	nextDueIn.CharLimit = 10
	nextDueIn.SetWidth(12)
	// Default: +1 month from now
	nextDueIn.SetValue(time.Now().AddDate(0, 1, 0).Format("02.01.2006"))

	remindIn := textinput.New()
	remindIn.SetValue("3")
	remindIn.CharLimit = 3
	remindIn.SetWidth(10)

	emailIn := textinput.New()
	emailIn.Placeholder = "user@example.com"
	emailIn.CharLimit = 120
	emailIn.SetWidth(40)

	_ = existingTags // reserved for future autocomplete

	return &WizardState{
		step:           WizardStepInfo,
		nameInput:      nameIn,
		urlInput:       urlIn,
		tagsInput:      tagsIn,
		isRecurring:    true,
		interval:       domain.IntervalMonthly,
		intervalCursor: 1, // monthly
		customDays:     14,
		customInput:    customIn,
		nextDueInput:   nextDueIn,
		remindDays:     3,
		remindInput:    remindIn,
		notifyMacOS:    true,
		notifyEmail:    false,
		emailInput:     emailIn,
		focusIndex:     0,
	}
}

// stepFieldCount returns how many focusable fields a given step has.
func (w *WizardState) stepFieldCount() int {
	switch w.step {
	case WizardStepInfo:
		return 3 // name, url, tags
	case WizardStepSchedule:
		n := 3 // type toggle + next due date + remind days
		if w.isRecurring {
			n++ // interval selector
			if w.interval == domain.IntervalCustom {
				n++ // custom days input
			}
		}
		return n
	case WizardStepNotifications:
		n := 2 // macOS toggle, email toggle
		if w.notifyEmail {
			n++ // email input
		}
		return n
	case WizardStepConfirm:
		return 0
	}
	return 0
}

// validateCurrentStep returns an error string if the current step has invalid data.
func (w *WizardState) validateCurrentStep() string {
	switch w.step {
	case WizardStepInfo:
		name := strings.TrimSpace(w.nameInput.Value())
		if name == "" {
			return i18n.T("validation.name_required")
		}
		rawURL := strings.TrimSpace(w.urlInput.Value())
		if rawURL != "" {
			if _, err := url.ParseRequestURI(rawURL); err != nil || (!strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://")) {
				return i18n.T("validation.invalid_url")
			}
		}
	case WizardStepSchedule:
		if w.isRecurring && w.interval == domain.IntervalCustom {
			if w.customDays <= 0 {
				return i18n.T("validation.custom_days")
			}
		}
		// Validate date
		dateStr := strings.TrimSpace(w.nextDueInput.Value())
		if dateStr != "" {
			if _, err := time.Parse("02.01.2006", dateStr); err != nil {
				return i18n.T("validation.invalid_date")
			}
		}
		if w.remindDays < 0 {
			return i18n.T("validation.remind_days")
		}
	case WizardStepNotifications:
		if w.notifyEmail {
			email := strings.TrimSpace(w.emailInput.Value())
			if email == "" || !strings.Contains(email, "@") {
				return i18n.T("validation.email_required")
			}
		}
	}
	return ""
}

// parseTags splits the comma-separated tags input into a trimmed slice.
func (w *WizardState) parseTags() []string {
	raw := strings.TrimSpace(w.tagsInput.Value())
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

// syncNumericFields parses the text inputs for custom days and remind days
// into their integer counterparts.
func (w *WizardState) syncNumericFields() {
	if v, err := strconv.Atoi(strings.TrimSpace(w.customInput.Value())); err == nil && v > 0 {
		w.customDays = v
	}
	if v, err := strconv.Atoi(strings.TrimSpace(w.remindInput.Value())); err == nil && v >= 0 {
		w.remindDays = v
	}
}

// buildReminder constructs a domain.Reminder from the wizard state.
func (w *WizardState) buildReminder() domain.Reminder {
	w.syncNumericFields()

	r := domain.NewReminder(strings.TrimSpace(w.nameInput.Value()))
	r.URL = strings.TrimSpace(w.urlInput.Value())
	r.Tags = w.parseTags()

	// Parse due date from input
	dateStr := strings.TrimSpace(w.nextDueInput.Value())
	if parsed, err := time.Parse("02.01.2006", dateStr); err == nil {
		r.NextDue = parsed
	} else {
		r.NextDue = time.Now().AddDate(0, 1, 0)
	}

	if w.isRecurring {
		r.Interval = w.interval
		if w.interval == domain.IntervalCustom {
			r.CustomDays = w.customDays
		}
	} else {
		r.Interval = domain.IntervalOnce
	}

	r.RemindDaysBefore = w.remindDays
	r.Notifications.MacOS = w.notifyMacOS
	r.Notifications.Email = w.notifyEmail

	return r
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

var (
	wizardBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 3)

	wizardLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8F8F2")).
				Bold(true)

	wizardDimStyle = lipgloss.NewStyle().
			Foreground(colorHint)

	wizardSelectedStyle = lipgloss.NewStyle().
				Foreground(colorSuccess).
				Bold(true)

	wizardToggleOnStyle = lipgloss.NewStyle().
				Foreground(colorSuccess).
				Bold(true)

	wizardToggleOffStyle = lipgloss.NewStyle().
				Foreground(colorHint)
)

// getStepLabels returns translated step labels.
func getStepLabels() []string {
	return []string{
		i18n.T("wizard.step_info"),
		i18n.T("wizard.step_schedule"),
		i18n.T("wizard.step_notify"),
		i18n.T("wizard.step_confirm"),
	}
}


// renderWizard produces the full wizard view string.
func renderWizard(w *WizardState, width, height int) string {
	var b strings.Builder

	// Step content
	switch w.step {
	case WizardStepInfo:
		b.WriteString(renderField(i18n.T("wizard.name"), w.nameInput.View(), w.focusIndex == 0, true))
		b.WriteString(renderField(i18n.T("wizard.url"), w.urlInput.View(), w.focusIndex == 1, false))
		b.WriteString(renderField(i18n.T("wizard.tags"), w.tagsInput.View(), w.focusIndex == 2, false))
		b.WriteString(wizardDimStyle.Render("    " + i18n.T("wizard.tags_hint")))

	case WizardStepSchedule:
		b.WriteString(renderToggleLine(i18n.T("wizard.type"), w.focusIndex == 0, w.isRecurring, i18n.T("wizard.recurring"), i18n.T("wizard.once")))
		b.WriteString("\n")

		fieldIdx := 1
		if w.isRecurring {
			b.WriteString(renderIntervalSelector(w, w.focusIndex == fieldIdx))
			fieldIdx++

			if w.interval == domain.IntervalCustom {
				b.WriteString(renderField(i18n.T("wizard.custom_days"), w.customInput.View(), w.focusIndex == fieldIdx, false))
				fieldIdx++
			}
		}

		b.WriteString(renderField(i18n.T("wizard.next_due"), w.nextDueInput.View(), w.focusIndex == fieldIdx, true))
		fieldIdx++

		b.WriteString(renderField(i18n.T("wizard.remind_before")+" ("+i18n.T("wizard.days")+")", w.remindInput.View(), w.focusIndex == fieldIdx, false))

	case WizardStepNotifications:
		b.WriteString(renderToggleLine(i18n.T("wizard.macos_notify"), w.focusIndex == 0, w.notifyMacOS, i18n.T("wizard.on"), i18n.T("wizard.off")))
		b.WriteString("\n")
		b.WriteString(renderToggleLine(i18n.T("wizard.email_notify"), w.focusIndex == 1, w.notifyEmail, i18n.T("wizard.on"), i18n.T("wizard.off")))

		if w.notifyEmail {
			b.WriteString("\n")
			b.WriteString(renderField(i18n.T("wizard.email"), w.emailInput.View(), w.focusIndex == 2, false))
		}

	case WizardStepConfirm:
		b.WriteString(renderSummary(w))
	}

	// Validation error
	if w.validationErr != "" {
		b.WriteString("\n")
		b.WriteString(styleError.Render(w.validationErr))
	}

	// Footer help bar
	b.WriteString("\n\n")
	switch w.step {
	case WizardStepInfo:
		b.WriteString(formatKeyHelp(
			"tab/C-j", i18n.T("wizard.next"),
			"S-tab/C-k", i18n.T("wizard.back"),
			"esc", i18n.T("wizard.cancel"),
		))
	case WizardStepSchedule, WizardStepNotifications:
		b.WriteString(formatKeyHelp(
			"tab/C-j", i18n.T("wizard.next"),
			"S-tab/C-k", i18n.T("wizard.back"),
			"h/l", "←/→",
			"esc", i18n.T("wizard.cancel"),
		))
	case WizardStepConfirm:
		b.WriteString(formatKeyHelp(
			"enter", i18n.T("wizard.save"),
			"shift+tab", i18n.T("wizard.back"),
			"esc", i18n.T("wizard.cancel"),
		))
	}

	// Title with step counter
	stepNum := int(w.step) + 1
	title := styleAccent.Bold(true).Render(i18n.T("wizard.title"))
	stepLabel := getStepLabels()[w.step]
	counter := styleHint.Render(fmt.Sprintf("%d/%d", stepNum, wizardTotalSteps)) +
		" " + styleAccent.Render(stepLabel)

	box := wizardBoxStyle.Render(b.String())
	boxWidth := lipgloss.Width(box)

	// Title left, counter right
	gap := boxWidth - lipgloss.Width(title) - lipgloss.Width(counter)
	if gap < 1 {
		gap = 1
	}
	header := title + strings.Repeat(" ", gap) + counter

	result := header + "\n" + box

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, result)
}

// renderField renders a labeled text input field.
func renderField(label, inputView string, focused, required bool) string {
	marker := "  "
	if focused {
		marker = "> "
	}
	reqMark := ""
	if required {
		reqMark = " " + styleError.Render("*")
	}
	return fmt.Sprintf("%s%s%s\n  %s\n\n", marker, wizardLabelStyle.Render(label), reqMark, inputView)
}

// renderToggleLine renders a boolean toggle that can be switched with space.
func renderToggleLine(label string, focused, value bool, onText, offText string) string {
	marker := "  "
	if focused {
		marker = "> "
	}

	var valStr string
	if value {
		valStr = wizardToggleOnStyle.Render("[" + onText + "]")
	} else {
		valStr = wizardToggleOffStyle.Render("[" + offText + "]")
	}

	return fmt.Sprintf("%s%s  %s\n", marker, wizardLabelStyle.Render(label), valStr)
}

// renderIntervalSelector renders the interval options as a horizontal selector.
func renderIntervalSelector(w *WizardState, focused bool) string {
	marker := "  "
	if focused {
		marker = "> "
	}

	var opts strings.Builder
	for i, opt := range intervalOptions {
		label := opt.String()
		if i == w.intervalCursor {
			opts.WriteString(wizardSelectedStyle.Render(" [" + label + "] "))
		} else {
			opts.WriteString(wizardDimStyle.Render("  " + label + "  "))
		}
	}

	return fmt.Sprintf("%s%s\n  %s\n\n", marker, wizardLabelStyle.Render(i18n.T("wizard.interval")), opts.String())
}

// renderSummary shows the read-only confirmation of all wizard fields.
func renderSummary(w *WizardState) string {
	w.syncNumericFields()
	var b strings.Builder

	summaryLine := func(label, value string) {
		b.WriteString(fmt.Sprintf("  %-16s  %s\n", wizardLabelStyle.Render(label), value))
	}

	summaryLine(i18n.T("wizard.name"), strings.TrimSpace(w.nameInput.Value()))

	rawURL := strings.TrimSpace(w.urlInput.Value())
	if rawURL != "" {
		summaryLine(i18n.T("wizard.url"), rawURL)
	}

	tags := w.parseTags()
	if len(tags) > 0 {
		summaryLine(i18n.T("wizard.tags"), strings.Join(tags, ", "))
	}

	if w.isRecurring {
		summaryLine(i18n.T("wizard.type"), i18n.T("wizard.recurring"))
		summaryLine(i18n.T("wizard.interval"), w.interval.String())
		if w.interval == domain.IntervalCustom {
			summaryLine(i18n.T("wizard.custom_days"), fmt.Sprintf("%d %s", w.customDays, i18n.T("wizard.days")))
		}
	} else {
		summaryLine(i18n.T("wizard.type"), i18n.T("wizard.once"))
	}

	summaryLine(i18n.T("wizard.remind_before"), fmt.Sprintf("%d %s", w.remindDays, i18n.T("wizard.days")))

	macOS := i18n.T("wizard.off")
	if w.notifyMacOS {
		macOS = i18n.T("wizard.on")
	}
	summaryLine(i18n.T("wizard.macos_notify"), macOS)

	email := i18n.T("wizard.off")
	if w.notifyEmail {
		email = i18n.T("wizard.on") + " (" + strings.TrimSpace(w.emailInput.Value()) + ")"
	}
	summaryLine(i18n.T("wizard.email_notify"), email)

	return b.String()
}
