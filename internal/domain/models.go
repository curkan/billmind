package domain

import (
	"errors"
	"time"

	"github.com/curkan/billmind/internal/i18n"
	"github.com/google/uuid"
)

// Sentinel errors
var (
	ErrNotFound         = errors.New("reminder not found")
	ErrInvalidData      = errors.New("invalid data format")
	ErrStorageCorrupted = errors.New("storage file corrupted")
)

type Interval string

const (
	IntervalWeekly  Interval = "weekly"
	IntervalMonthly Interval = "monthly"
	IntervalYearly  Interval = "yearly"
	IntervalOnce    Interval = "once"
	IntervalCustom  Interval = "custom"
)

func (i Interval) Valid() bool {
	switch i {
	case IntervalWeekly, IntervalMonthly, IntervalYearly, IntervalOnce, IntervalCustom:
		return true
	}
	return false
}

// String returns human-readable label for the interval using the current i18n language.
func (i Interval) String() string {
	switch i {
	case IntervalWeekly:
		return i18n.T("interval.weekly")
	case IntervalMonthly:
		return i18n.T("interval.monthly")
	case IntervalYearly:
		return i18n.T("interval.yearly")
	case IntervalOnce:
		return i18n.T("interval.once")
	case IntervalCustom:
		return i18n.T("interval.custom")
	}
	return string(i)
}

type NotifyStage int

const (
	NotifyNone     NotifyStage = 0
	NotifySoft     NotifyStage = 1
	NotifyUrgent   NotifyStage = 2
	NotifyCritical NotifyStage = 3
)

func (s NotifyStage) String() string {
	switch s {
	case NotifySoft:
		return "soft"
	case NotifyUrgent:
		return "urgent"
	case NotifyCritical:
		return "critical"
	default:
		return "none"
	}
}

type Notifications struct {
	MacOS bool `json:"macos"`
	Ntfy  bool `json:"ntfy"`
}

type Reminder struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	URL              string        `json:"url"`
	Tags             []string      `json:"tags"`
	Interval         Interval      `json:"interval"`
	CustomDays       int           `json:"custom_days,omitempty"`
	NextDue          time.Time     `json:"next_due"`
	RemindDaysBefore int           `json:"remind_days_before"`
	Notifications    Notifications `json:"notifications"`
	NotifyStage      NotifyStage   `json:"notify_stage"`
	PaidAt           *time.Time    `json:"paid_at"`
}

func NewReminder(name string) Reminder {
	return Reminder{
		ID:               uuid.New().String(),
		Name:             name,
		Tags:             []string{},
		Interval:         IntervalMonthly,
		RemindDaysBefore: 3,
		NextDue:          time.Now().AddDate(0, 1, 0),
	}
}

func (r Reminder) IsOverdue() bool {
	if r.IsPaid() {
		return false
	}
	return time.Now().After(r.NextDue)
}

func (r Reminder) IsDueSoon(days int) bool {
	if r.IsPaid() {
		return false
	}
	return !r.IsOverdue() && r.DaysUntilDue() <= days
}

func (r Reminder) IsPaid() bool {
	return r.PaidAt != nil
}

func (r Reminder) DaysUntilDue() int {
	return r.daysUntilDueFrom(time.Now())
}

func (r Reminder) daysUntilDueFrom(now time.Time) int {
	today := now.Truncate(24 * time.Hour)
	due := r.NextDue.Truncate(24 * time.Hour)
	return int(due.Sub(today).Hours() / 24)
}

// ShouldNotify returns the notification stage that should fire,
// or NotifyNone if no notification is needed.
func (r Reminder) ShouldNotify(now time.Time) NotifyStage {
	if r.IsPaid() {
		return NotifyNone
	}

	days := r.daysUntilDueFrom(now)

	switch {
	case days < 0 && r.NotifyStage < NotifyCritical:
		return NotifyCritical
	case days == 0 && r.NotifyStage < NotifyUrgent:
		return NotifyUrgent
	case days >= 0 && days <= r.RemindDaysBefore && r.NotifyStage < NotifySoft:
		return NotifySoft
	default:
		return NotifyNone
	}
}

func (r *Reminder) ResetNotifyStage() {
	r.NotifyStage = NotifyNone
}
