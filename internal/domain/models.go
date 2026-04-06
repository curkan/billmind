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

type Notifications struct {
	MacOS bool `json:"macos"`
	Email bool `json:"email"`
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
	now := time.Now().Truncate(24 * time.Hour)
	due := r.NextDue.Truncate(24 * time.Hour)
	d := int(due.Sub(now).Hours() / 24)
	return d
}
