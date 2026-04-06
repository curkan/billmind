package daemon_test

import (
	"context"
	"testing"
	"time"

	"github.com/curkan/billmind/internal/daemon"
	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/storage"
	"github.com/curkan/billmind/test/mocks"
)

func setupStore(t *testing.T, reminders []domain.Reminder) *storage.Storage {
	t.Helper()
	dir := t.TempDir()
	fs := mocks.NewMockFileSystem(dir)
	store := storage.New(fs, dir)
	if err := store.Save(context.Background(), reminders); err != nil {
		t.Fatalf("setup save: %v", err)
	}
	// Disable quiet hours so tests work at any time of day (start == end = no quiet window)
	if err := store.SaveSettings(context.Background(), storage.Settings{
		QuietHoursStart: 1,
		QuietHoursEnd:   1,
	}); err != nil {
		t.Fatalf("setup settings: %v", err)
	}
	return store
}

func TestRun_SendsGroupedNotifications(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(24 * time.Hour)

	reminders := []domain.Reminder{
		{
			ID:               "1",
			Name:             "Due today",
			NextDue:          now,
			RemindDaysBefore: 3,
			NotifyStage:      domain.NotifySoft,
			Notifications:    domain.Notifications{MacOS: true},
		},
		{
			ID:               "2",
			Name:             "Overdue",
			NextDue:          now.Add(-2 * 24 * time.Hour),
			RemindDaysBefore: 3,
			NotifyStage:      domain.NotifyUrgent,
			Notifications:    domain.Notifications{MacOS: true},
		},
	}

	store := setupStore(t, reminders)
	plat := &mocks.MockPlatform{}

	err := daemon.Run(context.Background(), store, plat)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(plat.Notifications) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(plat.Notifications))
	}

	// Verify stages were persisted
	updated, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	for _, r := range updated {
		switch r.ID {
		case "1":
			if r.NotifyStage != domain.NotifyUrgent {
				t.Errorf("reminder 1 stage = %v, want %v", r.NotifyStage, domain.NotifyUrgent)
			}
		case "2":
			if r.NotifyStage != domain.NotifyCritical {
				t.Errorf("reminder 2 stage = %v, want %v", r.NotifyStage, domain.NotifyCritical)
			}
		}
	}
}

func TestRun_QuietHours_NoNotifications(t *testing.T) {
	t.Parallel()

	// This test can't easily control time.Now() inside daemon.Run(),
	// so we verify quiet hours logic separately in quiethours_test.go.
	// Here we just verify no crash with empty data.
	store := setupStore(t, []domain.Reminder{})
	plat := &mocks.MockPlatform{}

	err := daemon.Run(context.Background(), store, plat)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(plat.Notifications) != 0 {
		t.Errorf("expected 0 notifications, got %d", len(plat.Notifications))
	}
}

func TestRun_AllPaid_NoNotifications(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(24 * time.Hour)
	paid := now.Add(-time.Hour)

	reminders := []domain.Reminder{
		{
			ID:               "1",
			Name:             "Paid one",
			NextDue:          now,
			RemindDaysBefore: 3,
			PaidAt:           &paid,
			Notifications:    domain.Notifications{MacOS: true},
		},
	}

	store := setupStore(t, reminders)
	plat := &mocks.MockPlatform{}

	err := daemon.Run(context.Background(), store, plat)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(plat.Notifications) != 0 {
		t.Errorf("expected 0 notifications for paid reminders, got %d", len(plat.Notifications))
	}
}

func TestRun_AlreadyNotified_NoDuplicate(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(24 * time.Hour)

	reminders := []domain.Reminder{
		{
			ID:               "1",
			Name:             "Already notified",
			NextDue:          now.Add(2 * 24 * time.Hour),
			RemindDaysBefore: 3,
			NotifyStage:      domain.NotifySoft,
			Notifications:    domain.Notifications{MacOS: true},
		},
	}

	store := setupStore(t, reminders)
	plat := &mocks.MockPlatform{}

	err := daemon.Run(context.Background(), store, plat)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(plat.Notifications) != 0 {
		t.Errorf("expected 0 notifications (already sent), got %d", len(plat.Notifications))
	}
}
