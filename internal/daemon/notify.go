package daemon

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/ntfy"
	"github.com/curkan/billmind/internal/platform"
)

// NotifyGroups holds reminders grouped by the stage that should fire now.
type NotifyGroups struct {
	Soft     []indexedReminder
	Urgent   []indexedReminder
	Critical []indexedReminder
}

// indexedReminder tracks original slice index for applying stage updates back.
type indexedReminder struct {
	Index    int
	Reminder domain.Reminder
	Stage    domain.NotifyStage
}

func (g NotifyGroups) Empty() bool {
	return len(g.Soft) == 0 && len(g.Urgent) == 0 && len(g.Critical) == 0
}

// GroupByStage evaluates ShouldNotify for each reminder and groups results.
func GroupByStage(reminders []domain.Reminder, now time.Time) NotifyGroups {
	var g NotifyGroups

	for i, r := range reminders {
		if !r.Notifications.MacOS && !r.Notifications.Ntfy {
			continue
		}

		stage := r.ShouldNotify(now)
		if stage == domain.NotifyNone {
			continue
		}

		ir := indexedReminder{Index: i, Reminder: r, Stage: stage}

		switch stage {
		case domain.NotifySoft:
			g.Soft = append(g.Soft, ir)
		case domain.NotifyUrgent:
			g.Urgent = append(g.Urgent, ir)
		case domain.NotifyCritical:
			g.Critical = append(g.Critical, ir)
		}
	}

	return g
}

// ApplyStages writes back the new NotifyStage to the original reminders slice.
func ApplyStages(reminders []domain.Reminder, g NotifyGroups) {
	for _, ir := range g.Soft {
		reminders[ir.Index].NotifyStage = ir.Stage
	}
	for _, ir := range g.Urgent {
		reminders[ir.Index].NotifyStage = ir.Stage
	}
	for _, ir := range g.Critical {
		reminders[ir.Index].NotifyStage = ir.Stage
	}
}

// SendGrouped sends one OS notification and one ntfy message per non-empty stage group.
func SendGrouped(ctx context.Context, plat platform.Platform, ntfyTopic string, g NotifyGroups) error {
	var errs []error

	type batch struct {
		label    string
		items    []indexedReminder
		priority ntfy.Priority
	}

	batches := []batch{
		{"OVERDUE", g.Critical, ntfy.PriorityUrgent},
		{"Payment today", g.Urgent, ntfy.PriorityHigh},
		{"Payment soon", g.Soft, ntfy.PriorityDefault},
	}

	for _, b := range batches {
		if len(b.items) == 0 {
			continue
		}
		msg := FormatBatch(b.label, b.items)

		if hasChannel(b.items, func(n domain.Notifications) bool { return n.MacOS }) {
			if err := plat.SendNotification(ctx, "billmind", msg); err != nil {
				errs = append(errs, fmt.Errorf("%s os notification: %w", b.label, err))
			}
		}

		if ntfyTopic != "" && hasChannel(b.items, func(n domain.Notifications) bool { return n.Ntfy }) {
			if err := ntfy.Send(ctx, ntfyTopic, "billmind — "+b.label, msg, b.priority); err != nil {
				errs = append(errs, fmt.Errorf("%s ntfy: %w", b.label, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("notification errors: %v", errs)
	}
	return nil
}

func hasChannel(items []indexedReminder, check func(domain.Notifications) bool) bool {
	for _, ir := range items {
		if check(ir.Reminder.Notifications) {
			return true
		}
	}
	return false
}

// FormatBatch creates notification body text.
// Single reminder: "Hetzner VPS — payment in 3 days (May 01)"
// Multiple: "Payment soon (3): Hetzner VPS, GitHub Pro, Netflix"
func FormatBatch(prefix string, items []indexedReminder) string {
	if len(items) == 1 {
		r := items[0].Reminder
		days := r.DaysUntilDue()
		switch {
		case days < 0:
			return fmt.Sprintf("%s — overdue by %d day(s)", r.Name, -days)
		case days == 0:
			return fmt.Sprintf("%s — payment today!", r.Name)
		default:
			return fmt.Sprintf("%s — payment in %d day(s) (%s)",
				r.Name, days, r.NextDue.Format("Jan 02"))
		}
	}

	names := make([]string, 0, len(items))
	for _, ir := range items {
		names = append(names, ir.Reminder.Name)
	}
	return fmt.Sprintf("%s (%d): %s", prefix, len(items), strings.Join(names, ", "))
}
