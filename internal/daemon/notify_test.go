package daemon

import (
	"testing"
	"time"

	"github.com/curkan/billmind/internal/domain"
)

func TestGroupByStage(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)

	reminders := []domain.Reminder{
		{
			Name:             "Soft one",
			NextDue:          time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC),
			RemindDaysBefore: 3,
			NotifyStage:      domain.NotifyNone,
			Notifications:    domain.Notifications{MacOS: true},
		},
		{
			Name:             "Urgent one",
			NextDue:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
			RemindDaysBefore: 3,
			NotifyStage:      domain.NotifySoft,
			Notifications:    domain.Notifications{MacOS: true},
		},
		{
			Name:             "Critical one",
			NextDue:          time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC),
			RemindDaysBefore: 3,
			NotifyStage:      domain.NotifyUrgent,
			Notifications:    domain.Notifications{MacOS: true},
		},
		{
			Name:             "No notifications enabled",
			NextDue:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
			RemindDaysBefore: 3,
			NotifyStage:      domain.NotifyNone,
			Notifications:    domain.Notifications{MacOS: false, Ntfy: false},
		},
	}

	g := GroupByStage(reminders, now)

	if len(g.Soft) != 1 {
		t.Errorf("expected 1 soft, got %d", len(g.Soft))
	}
	if len(g.Urgent) != 1 {
		t.Errorf("expected 1 urgent, got %d", len(g.Urgent))
	}
	if len(g.Critical) != 1 {
		t.Errorf("expected 1 critical, got %d", len(g.Critical))
	}
}

func TestGroupByStage_AllPaid(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	paid := time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC)

	reminders := []domain.Reminder{
		{
			Name:             "Paid",
			NextDue:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
			RemindDaysBefore: 3,
			PaidAt:           &paid,
			Notifications:    domain.Notifications{MacOS: true},
		},
	}

	g := GroupByStage(reminders, now)
	if !g.Empty() {
		t.Error("expected empty groups for paid reminders")
	}
}

func TestApplyStages(t *testing.T) {
	t.Parallel()

	reminders := []domain.Reminder{
		{Name: "A", NotifyStage: domain.NotifyNone},
		{Name: "B", NotifyStage: domain.NotifySoft},
	}

	g := NotifyGroups{
		Soft:   []indexedReminder{{Index: 0, Stage: domain.NotifySoft}},
		Urgent: []indexedReminder{{Index: 1, Stage: domain.NotifyUrgent}},
	}

	ApplyStages(reminders, g)

	if reminders[0].NotifyStage != domain.NotifySoft {
		t.Errorf("reminders[0].NotifyStage = %v, want %v", reminders[0].NotifyStage, domain.NotifySoft)
	}
	if reminders[1].NotifyStage != domain.NotifyUrgent {
		t.Errorf("reminders[1].NotifyStage = %v, want %v", reminders[1].NotifyStage, domain.NotifyUrgent)
	}
}

func TestFormatBatch_SingleItem(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(24 * time.Hour)
	items := []indexedReminder{
		{Reminder: domain.Reminder{
			Name:    "Hetzner VPS",
			NextDue: now.Add(3 * 24 * time.Hour),
		}},
	}

	msg := FormatBatch("Payment soon", items)
	if msg == "" {
		t.Error("expected non-empty message")
	}
	// Should contain the name
	if !contains(msg, "Hetzner VPS") {
		t.Errorf("message should contain reminder name, got: %s", msg)
	}
}

func TestFormatBatch_MultipleItems(t *testing.T) {
	t.Parallel()

	items := []indexedReminder{
		{Reminder: domain.Reminder{Name: "Hetzner VPS", NextDue: time.Now().Add(24 * time.Hour)}},
		{Reminder: domain.Reminder{Name: "GitHub Pro", NextDue: time.Now().Add(24 * time.Hour)}},
		{Reminder: domain.Reminder{Name: "Netflix", NextDue: time.Now().Add(24 * time.Hour)}},
	}

	msg := FormatBatch("Payment soon", items)
	if !contains(msg, "(3)") {
		t.Errorf("message should contain count, got: %s", msg)
	}
	if !contains(msg, "Hetzner VPS") || !contains(msg, "Netflix") {
		t.Errorf("message should contain all names, got: %s", msg)
	}
}

func TestFormatBatch_Overdue(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(24 * time.Hour)
	items := []indexedReminder{
		{Reminder: domain.Reminder{
			Name:    "Hetzner VPS",
			NextDue: now.Add(-2 * 24 * time.Hour),
		}},
	}

	msg := FormatBatch("OVERDUE", items)
	if !contains(msg, "overdue") {
		t.Errorf("message should contain 'overdue', got: %s", msg)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
