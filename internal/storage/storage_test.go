package storage_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/storage"
	"github.com/curkan/billmind/test/mocks"
)

func newTestStorage(t *testing.T) (*storage.Storage, string) {
	t.Helper()
	dir := t.TempDir()
	fs := mocks.NewMockFileSystem(dir)
	return storage.New(fs, dir), dir
}

func sampleReminders() []domain.Reminder {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	return []domain.Reminder{
		{
			ID:               "id-001",
			Name:             "Electricity",
			URL:              "https://electric.example.com",
			Tags:             []string{"utilities", "home"},
			Interval:         domain.IntervalMonthly,
			NextDue:          now.Add(7 * 24 * time.Hour),
			RemindDaysBefore: 3,
			Notifications:    domain.Notifications{MacOS: true, Email: false},
			PaidAt:           nil,
		},
		{
			ID:               "id-002",
			Name:             "Internet",
			URL:              "https://isp.example.com",
			Tags:             []string{"utilities"},
			Interval:         domain.IntervalMonthly,
			NextDue:          now.Add(14 * 24 * time.Hour),
			RemindDaysBefore: 5,
			Notifications:    domain.Notifications{MacOS: true, Email: true},
			PaidAt:           nil,
		},
	}
}

func TestLoadSave(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		reminders []domain.Reminder
	}{
		{
			name:      "single reminder",
			reminders: sampleReminders()[:1],
		},
		{
			name:      "multiple reminders",
			reminders: sampleReminders(),
		},
		{
			name:      "empty list",
			reminders: []domain.Reminder{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store, _ := newTestStorage(t)
			ctx := context.Background()

			if err := store.Save(ctx, tt.reminders); err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			loaded, err := store.Load(ctx)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if len(loaded) != len(tt.reminders) {
				t.Fatalf("Load() returned %d reminders, want %d", len(loaded), len(tt.reminders))
			}

			for i, got := range loaded {
				want := tt.reminders[i]
				if got.ID != want.ID {
					t.Errorf("reminder[%d].ID = %q, want %q", i, got.ID, want.ID)
				}
				if got.Name != want.Name {
					t.Errorf("reminder[%d].Name = %q, want %q", i, got.Name, want.Name)
				}
				if got.URL != want.URL {
					t.Errorf("reminder[%d].URL = %q, want %q", i, got.URL, want.URL)
				}
				if got.Interval != want.Interval {
					t.Errorf("reminder[%d].Interval = %q, want %q", i, got.Interval, want.Interval)
				}
				if !got.NextDue.Equal(want.NextDue) {
					t.Errorf("reminder[%d].NextDue = %v, want %v", i, got.NextDue, want.NextDue)
				}
				if got.RemindDaysBefore != want.RemindDaysBefore {
					t.Errorf("reminder[%d].RemindDaysBefore = %d, want %d", i, got.RemindDaysBefore, want.RemindDaysBefore)
				}
				if len(got.Tags) != len(want.Tags) {
					t.Errorf("reminder[%d].Tags length = %d, want %d", i, len(got.Tags), len(want.Tags))
				}
			}
		})
	}
}

func TestLoadEmpty(t *testing.T) {
	t.Parallel()
	store, _ := newTestStorage(t)
	ctx := context.Background()

	reminders, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(reminders) != 0 {
		t.Errorf("Load() returned %d reminders, want 0", len(reminders))
	}
}

func TestAtomicWrite(t *testing.T) {
	t.Parallel()
	store, dir := newTestStorage(t)
	ctx := context.Background()

	reminders := sampleReminders()
	if err := store.Save(ctx, reminders); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify the data file exists after save.
	dataPath := filepath.Join(dir, "data.json")
	info, err := os.Stat(dataPath)
	if err != nil {
		t.Fatalf("data.json should exist after Save(): %v", err)
	}
	if info.Size() == 0 {
		t.Error("data.json should not be empty after Save()")
	}

	// Verify no leftover temp files.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("leftover temp file found: %s", e.Name())
		}
	}
}

func TestBackupRotation(t *testing.T) {
	t.Parallel()
	store, dir := newTestStorage(t)
	ctx := context.Background()

	reminders := sampleReminders()

	// Save 12 times; each save creates a backup (except the first, since
	// no data.json exists yet).
	for i := 0; i < 12; i++ {
		if err := store.Save(ctx, reminders); err != nil {
			t.Fatalf("Save() iteration %d error = %v", i, err)
		}
		// Small delay so backup timestamps differ.
		time.Sleep(10 * time.Millisecond)
	}

	backupPath := filepath.Join(dir, "backups")
	entries, err := os.ReadDir(backupPath)
	if err != nil {
		t.Fatalf("reading backups dir: %v", err)
	}

	backupCount := 0
	for _, e := range entries {
		if !e.IsDir() {
			backupCount++
		}
	}

	if backupCount > 10 {
		t.Errorf("expected at most 10 backups, got %d", backupCount)
	}
}

func TestBackupKeepsLast10(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	fs := mocks.NewMockFileSystem(dir)
	store := storage.New(fs, dir)
	ctx := context.Background()

	// Manually create 15 backup files with known timestamps.
	backupPath := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupPath, 0o755); err != nil {
		t.Fatalf("creating backup dir: %v", err)
	}

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 15; i++ {
		ts := baseTime.Add(time.Duration(i) * time.Hour).Format("2006-01-02_150405")
		name := fmt.Sprintf("data_%s.json", ts)
		path := filepath.Join(backupPath, name)
		if err := os.WriteFile(path, []byte(`{}`), 0o644); err != nil {
			t.Fatalf("creating backup %d: %v", i, err)
		}
	}

	// Write a data.json so the next Save triggers a backup + rotation.
	dataPath := filepath.Join(dir, "data.json")
	if err := os.WriteFile(dataPath, []byte(`{"reminders":[],"settings":{"email":""}}`), 0o644); err != nil {
		t.Fatalf("writing data.json: %v", err)
	}

	reminders := sampleReminders()
	if err := store.Save(ctx, reminders); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	entries, err := os.ReadDir(backupPath)
	if err != nil {
		t.Fatalf("reading backups dir: %v", err)
	}

	backupCount := 0
	for _, e := range entries {
		if !e.IsDir() {
			backupCount++
		}
	}

	if backupCount > 10 {
		t.Errorf("expected at most 10 backups after rotation, got %d", backupCount)
	}

	// Verify that the oldest backups were deleted.
	for i := 0; i < 6; i++ {
		ts := baseTime.Add(time.Duration(i) * time.Hour).Format("2006-01-02_150405")
		name := fmt.Sprintf("data_%s.json", ts)
		path := filepath.Join(backupPath, name)
		if _, err := os.Stat(path); err == nil {
			t.Errorf("old backup %s should have been deleted", name)
		}
	}
}

func TestCorruptedFile(t *testing.T) {
	t.Parallel()
	store, dir := newTestStorage(t)
	ctx := context.Background()

	// Write invalid JSON to data.json.
	dataPath := filepath.Join(dir, "data.json")
	if err := os.WriteFile(dataPath, []byte(`{invalid json!!!`), 0o644); err != nil {
		t.Fatalf("writing corrupted file: %v", err)
	}

	_, err := store.Load(ctx)
	if err == nil {
		t.Fatal("Load() should return error for corrupted file")
	}

	if !errors.Is(err, domain.ErrStorageCorrupted) {
		t.Errorf("Load() error = %v, want domain.ErrStorageCorrupted", err)
	}
}
