package domain

import (
	"testing"
	"time"
)

func TestIntervalValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		interval Interval
		want     bool
	}{
		{name: "weekly is valid", interval: IntervalWeekly, want: true},
		{name: "monthly is valid", interval: IntervalMonthly, want: true},
		{name: "yearly is valid", interval: IntervalYearly, want: true},
		{name: "once is valid", interval: IntervalOnce, want: true},
		{name: "custom is valid", interval: IntervalCustom, want: true},
		{name: "empty string is invalid", interval: Interval(""), want: false},
		{name: "arbitrary string is invalid", interval: Interval("daily"), want: false},
		{name: "uppercase is invalid", interval: Interval("Weekly"), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.interval.Valid()
			if got != tc.want {
				t.Errorf("Interval(%q).Valid() = %v, want %v", tc.interval, got, tc.want)
			}
		})
	}
}

func TestIntervalString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		interval Interval
		want     string
	}{
		{name: "weekly label", interval: IntervalWeekly, want: "еженедельно"},
		{name: "monthly label", interval: IntervalMonthly, want: "ежемесячно"},
		{name: "yearly label", interval: IntervalYearly, want: "ежегодно"},
		{name: "once label", interval: IntervalOnce, want: "единоразово"},
		{name: "custom label", interval: IntervalCustom, want: "своё"},
		{name: "unknown falls back to raw value", interval: Interval("biweekly"), want: "biweekly"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.interval.String()
			if got != tc.want {
				t.Errorf("Interval(%q).String() = %q, want %q", tc.interval, got, tc.want)
			}
		})
	}
}

func TestIsOverdue(t *testing.T) {
	t.Parallel()

	now := time.Now()
	paidAt := now.Add(-48 * time.Hour)

	tests := []struct {
		name string
		r    Reminder
		want bool
	}{
		{
			name: "past due date is overdue",
			r:    Reminder{NextDue: now.Add(-24 * time.Hour)},
			want: true,
		},
		{
			name: "future due date is not overdue",
			r:    Reminder{NextDue: now.Add(24 * time.Hour)},
			want: false,
		},
		{
			name: "paid reminder is not overdue even if past due",
			r:    Reminder{NextDue: now.Add(-24 * time.Hour), PaidAt: &paidAt},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.r.IsOverdue()
			if got != tc.want {
				t.Errorf("IsOverdue() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsDueSoon(t *testing.T) {
	t.Parallel()

	now := time.Now()
	paidAt := now

	tests := []struct {
		name      string
		r         Reminder
		threshold int
		want      bool
	}{
		{
			name:      "2 days left with threshold 3 is due soon",
			r:         Reminder{NextDue: now.Add(2 * 24 * time.Hour)},
			threshold: 3,
			want:      true,
		},
		{
			name:      "5 days left with threshold 3 is not due soon",
			r:         Reminder{NextDue: now.Add(5 * 24 * time.Hour)},
			threshold: 3,
			want:      false,
		},
		{
			name:      "paid reminder is not due soon",
			r:         Reminder{NextDue: now.Add(2 * 24 * time.Hour), PaidAt: &paidAt},
			threshold: 3,
			want:      false,
		},
		{
			name:      "overdue reminder is not due soon",
			r:         Reminder{NextDue: now.Add(-24 * time.Hour)},
			threshold: 3,
			want:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.r.IsDueSoon(tc.threshold)
			if got != tc.want {
				t.Errorf("IsDueSoon(%d) = %v, want %v", tc.threshold, got, tc.want)
			}
		})
	}
}

func TestDaysUntilDue(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(24 * time.Hour)

	tests := []struct {
		name string
		r    Reminder
		want int
	}{
		{
			name: "due in 7 days",
			r:    Reminder{NextDue: now.Add(7 * 24 * time.Hour)},
			want: 7,
		},
		{
			name: "due today",
			r:    Reminder{NextDue: now},
			want: 0,
		},
		{
			name: "overdue by 3 days",
			r:    Reminder{NextDue: now.Add(-3 * 24 * time.Hour)},
			want: -3,
		},
		{
			name: "due in 30 days",
			r:    Reminder{NextDue: now.Add(30 * 24 * time.Hour)},
			want: 30,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.r.DaysUntilDue()
			if got != tc.want {
				t.Errorf("DaysUntilDue() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestIsPaid(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name string
		r    Reminder
		want bool
	}{
		{
			name: "nil PaidAt is not paid",
			r:    Reminder{PaidAt: nil},
			want: false,
		},
		{
			name: "non-nil PaidAt is paid",
			r:    Reminder{PaidAt: &now},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.r.IsPaid()
			if got != tc.want {
				t.Errorf("IsPaid() = %v, want %v", got, tc.want)
			}
		})
	}
}

func ptr(t time.Time) *time.Time { return &t }

func TestShouldNotify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		reminder  Reminder
		now       time.Time
		wantStage NotifyStage
	}{
		{
			name: "paid reminder returns none",
			reminder: Reminder{
				NextDue:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
				PaidAt:           ptr(time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)),
				RemindDaysBefore: 3,
			},
			now:       time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC),
			wantStage: NotifyNone,
		},
		{
			name: "soft stage fires within remind window",
			reminder: Reminder{
				NextDue:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
				RemindDaysBefore: 3,
				NotifyStage:      NotifyNone,
			},
			now:       time.Date(2026, 4, 7, 12, 0, 0, 0, time.UTC),
			wantStage: NotifySoft,
		},
		{
			name: "soft already sent returns none",
			reminder: Reminder{
				NextDue:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
				RemindDaysBefore: 3,
				NotifyStage:      NotifySoft,
			},
			now:       time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC),
			wantStage: NotifyNone,
		},
		{
			name: "urgent on due day",
			reminder: Reminder{
				NextDue:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
				RemindDaysBefore: 3,
				NotifyStage:      NotifySoft,
			},
			now:       time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC),
			wantStage: NotifyUrgent,
		},
		{
			name: "critical when overdue",
			reminder: Reminder{
				NextDue:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
				RemindDaysBefore: 3,
				NotifyStage:      NotifyUrgent,
			},
			now:       time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC),
			wantStage: NotifyCritical,
		},
		{
			name: "catch-up skips to critical after long sleep",
			reminder: Reminder{
				NextDue:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
				RemindDaysBefore: 3,
				NotifyStage:      NotifyNone,
			},
			now:       time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC),
			wantStage: NotifyCritical,
		},
		{
			name: "too early returns none",
			reminder: Reminder{
				NextDue:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
				RemindDaysBefore: 3,
				NotifyStage:      NotifyNone,
			},
			now:       time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
			wantStage: NotifyNone,
		},
		{
			name: "all stages exhausted returns none",
			reminder: Reminder{
				NextDue:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
				RemindDaysBefore: 3,
				NotifyStage:      NotifyCritical,
			},
			now:       time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC),
			wantStage: NotifyNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.reminder.ShouldNotify(tt.now)
			if got != tt.wantStage {
				t.Errorf("ShouldNotify() = %v, want %v", got, tt.wantStage)
			}
		})
	}
}

func TestResetNotifyStage(t *testing.T) {
	t.Parallel()

	r := Reminder{NotifyStage: NotifyCritical}
	r.ResetNotifyStage()
	if r.NotifyStage != NotifyNone {
		t.Errorf("ResetNotifyStage() stage = %v, want %v", r.NotifyStage, NotifyNone)
	}
}

func TestNewReminder(t *testing.T) {
	t.Parallel()

	r := NewReminder("Netflix")

	if r.ID == "" {
		t.Error("NewReminder should generate a non-empty ID")
	}

	if r.Name != "Netflix" {
		t.Errorf("NewReminder name = %q, want %q", r.Name, "Netflix")
	}

	if r.Interval != IntervalMonthly {
		t.Errorf("NewReminder interval = %q, want %q", r.Interval, IntervalMonthly)
	}

	if r.RemindDaysBefore != 3 {
		t.Errorf("NewReminder RemindDaysBefore = %d, want 3", r.RemindDaysBefore)
	}

	if r.Tags == nil {
		t.Error("NewReminder Tags should be initialized, not nil")
	}

	if len(r.Tags) != 0 {
		t.Errorf("NewReminder Tags length = %d, want 0", len(r.Tags))
	}

	if r.PaidAt != nil {
		t.Error("NewReminder PaidAt should be nil")
	}

	if r.NextDue.IsZero() {
		t.Error("NewReminder NextDue should not be zero")
	}

	// NextDue should be approximately 1 month from now
	expectedDue := time.Now().AddDate(0, 1, 0)
	diff := r.NextDue.Sub(expectedDue)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("NewReminder NextDue is not ~1 month from now, got %v", r.NextDue)
	}

	// Two reminders should have different IDs
	r2 := NewReminder("Spotify")
	if r.ID == r2.ID {
		t.Error("two NewReminder calls should produce different IDs")
	}
}
