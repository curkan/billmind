package ui

import (
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/platform"
	"github.com/curkan/billmind/internal/storage"
	"github.com/curkan/billmind/test/mocks"
)

// newTestModel builds a Model with the given reminders pre-loaded for testing.
func newTestModel(t *testing.T, reminders []domain.Reminder) Model {
	t.Helper()

	tmpDir := t.TempDir()
	fs := mocks.NewMockFileSystem(tmpDir)
	store := storage.New(fs, tmpDir)

	if err := store.Save(context.Background(), reminders); err != nil {
		t.Fatalf("saving test reminders: %v", err)
	}

	plat := platform.New()
	m := New(store, plat)

	// Copy reminders so mutations in tests don't affect the original slice.
	m.reminders = make([]domain.Reminder, len(reminders))
	copy(m.reminders, reminders)
	m.sortReminders()
	m.updateListItems()

	m.width = 120
	m.height = 40
	m.table.SetHeight(36)
	m.table.SetWidth(120)

	return m
}

// sendKey simulates a single key press through Update, returning the new Model.
// The key string should match what tea.KeyPressMsg.String() returns in bubbletea v2,
// e.g. "q", "d", "space", "esc", "enter", "ctrl+c".
func sendKey(t *testing.T, m Model, key string) Model {
	t.Helper()
	msg := keyMsg(key)
	result, _ := m.Update(msg)
	return result.(Model)
}

// keyMsg constructs a tea.KeyPressMsg that produces the given string from String().
// For printable single-rune keys, we set Text so String() returns it directly.
// For special keys like "space", "esc", "enter", we set the appropriate Code.
func keyMsg(key string) tea.KeyPressMsg {
	switch key {
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	default:
		// For printable characters, set Text so String() returns it.
		if len(key) == 1 {
			return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
		}
		return tea.KeyPressMsg{Text: key}
	}
}

// makeReminder creates a test reminder with the given name and days until due.
func makeReminder(name string, daysUntilDue int) domain.Reminder {
	r := domain.NewReminder(name)
	r.NextDue = time.Now().AddDate(0, 0, daysUntilDue)
	return r
}

// makePaidReminder creates a test reminder that is already marked as paid.
func makePaidReminder(name string, daysUntilDue int) domain.Reminder {
	r := makeReminder(name, daysUntilDue)
	now := time.Now()
	r.PaidAt = &now
	return r
}

func TestListPaidConfirm(t *testing.T) {
	t.Parallel()

	reminders := []domain.Reminder{
		makeReminder("Netflix", 5),
		makeReminder("Spotify", 10),
	}

	t.Run("space opens confirm, y advances next due", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t, reminders)

		oldNextDue := m.reminders[0].NextDue

		// Press space to open confirmation modal
		m = sendKey(t, m, "space")
		if m.viewMode != ViewConfirmPaid {
			t.Fatalf("expected ViewConfirmPaid, got %d", m.viewMode)
		}

		// Press y to confirm payment
		m = sendKey(t, m, "y")
		if m.viewMode != ViewList {
			t.Fatalf("expected ViewList after confirm, got %d", m.viewMode)
		}

		// NextDue should have advanced
		for _, r := range m.reminders {
			if r.Name == "Netflix" {
				if !r.NextDue.After(oldNextDue) {
					t.Fatal("expected NextDue to advance after payment confirmation")
				}
				return
			}
		}
		t.Fatal("Netflix reminder not found")
	})
}

func TestListSorting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []domain.Reminder
		expected []string
	}{
		{
			name: "sorted by date ascending",
			input: []domain.Reminder{
				makeReminder("Later", 30),
				makeReminder("Sooner", 3),
				makeReminder("Middle", 15),
			},
			expected: []string{"Sooner", "Middle", "Later"},
		},
		{
			name: "all sorted by date",
			input: []domain.Reminder{
				makeReminder("C", 20),
				makeReminder("A", 1),
				makeReminder("B", 10),
			},
			expected: []string{"A", "B", "C"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := newTestModel(t, tc.input)

			rows := m.table.Rows()
			if len(rows) != len(tc.expected) {
				t.Fatalf("expected %d rows, got %d", len(tc.expected), len(rows))
			}

			// Compare using filteredReminders which preserves domain data
			filtered := m.filteredReminders()
			for i, r := range filtered {
				if r.Name != tc.expected[i] {
					t.Errorf("position %d: expected %q, got %q", i, tc.expected[i], r.Name)
				}
			}
		})
	}
}

func TestGGPending(t *testing.T) {
	t.Parallel()

	reminders := []domain.Reminder{
		makeReminder("First", 1),
		makeReminder("Second", 5),
		makeReminder("Third", 10),
	}

	t.Run("gg moves cursor to top", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t, reminders)

		// Move cursor down first.
		m.table.SetCursor(2)

		// First g sets gPending.
		m = sendKey(t, m, "g")
		if !m.gPending {
			t.Fatal("expected gPending to be true after first g")
		}

		// Second g should move to top.
		m = sendKey(t, m, "g")
		if m.gPending {
			t.Fatal("expected gPending to be false after second g")
		}
		if m.table.Cursor() != 0 {
			t.Fatalf("expected cursor at 0, got %d", m.table.Cursor())
		}
	})

	t.Run("g timeout resets pending", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t, reminders)
		m.table.SetCursor(2)

		m = sendKey(t, m, "g")
		if !m.gPending {
			t.Fatal("expected gPending to be true")
		}

		// Simulate the timeout message.
		result, _ := m.Update(resetGPendingMsg{})
		m = result.(Model)
		if m.gPending {
			t.Fatal("expected gPending to be false after timeout")
		}
	})
}

func TestDDPending(t *testing.T) {
	t.Parallel()

	reminders := []domain.Reminder{
		makeReminder("Target", 5),
	}

	t.Run("dd switches to ViewDelete", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t, reminders)

		// First d sets dPending.
		m = sendKey(t, m, "d")
		if !m.dPending {
			t.Fatal("expected dPending to be true after first d")
		}

		// Second d should switch to ViewDelete.
		m = sendKey(t, m, "d")
		if m.dPending {
			t.Fatal("expected dPending to be false after second d")
		}
		if m.viewMode != ViewDelete {
			t.Fatalf("expected viewMode ViewDelete, got %d", m.viewMode)
		}
	})

	t.Run("d timeout resets pending", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t, reminders)

		m = sendKey(t, m, "d")
		if !m.dPending {
			t.Fatal("expected dPending to be true")
		}

		result, _ := m.Update(resetDPendingMsg{})
		m = result.(Model)
		if m.dPending {
			t.Fatal("expected dPending to be false after timeout")
		}
	})
}

func TestDeleteConfirm(t *testing.T) {
	t.Parallel()

	reminders := []domain.Reminder{
		makeReminder("ToDelete", 5),
		makeReminder("ToKeep", 10),
	}

	t.Run("y confirms deletion", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t, reminders)

		// Enter delete mode via dd.
		m = sendKey(t, m, "d")
		m = sendKey(t, m, "d")
		if m.viewMode != ViewDelete {
			t.Fatalf("expected ViewDelete, got %d", m.viewMode)
		}

		// Confirm with y.
		m = sendKey(t, m, "y")

		if m.viewMode != ViewList {
			t.Fatalf("expected ViewList after confirm, got %d", m.viewMode)
		}

		if len(m.reminders) != 1 {
			t.Fatalf("expected 1 reminder, got %d", len(m.reminders))
		}

		if m.reminders[0].Name != "ToKeep" {
			t.Fatalf("expected remaining reminder to be ToKeep, got %q", m.reminders[0].Name)
		}
	})
}

func TestDeleteCancel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{"cancel with n", "n"},
		{"cancel with esc", "esc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reminders := []domain.Reminder{
				makeReminder("Keep", 5),
			}
			m := newTestModel(t, reminders)

			// Enter delete mode.
			m = sendKey(t, m, "d")
			m = sendKey(t, m, "d")

			// Cancel.
			m = sendKey(t, m, tc.key)

			if m.viewMode != ViewList {
				t.Fatalf("expected ViewList after cancel, got %d", m.viewMode)
			}
			if len(m.reminders) != 1 {
				t.Fatalf("expected 1 reminder unchanged, got %d", len(m.reminders))
			}
		})
	}
}

func TestUndo(t *testing.T) {
	t.Parallel()

	t.Run("undo paid confirmation", func(t *testing.T) {
		t.Parallel()

		reminders := []domain.Reminder{
			makeReminder("Service", 5),
		}
		m := newTestModel(t, reminders)

		oldNextDue := m.reminders[0].NextDue

		// Confirm payment: space → y
		m = sendKey(t, m, "space")
		m = sendKey(t, m, "y")

		// NextDue should have advanced
		for _, r := range m.reminders {
			if r.Name == "Service" {
				if !r.NextDue.After(oldNextDue) {
					t.Fatal("expected NextDue to advance after payment")
				}
				break
			}
		}

		// Undo
		m = sendKey(t, m, "u")

		for _, r := range m.reminders {
			if r.Name == "Service" {
				if r.NextDue != oldNextDue {
					t.Fatal("expected NextDue to revert after undo")
				}
				return
			}
		}
		t.Fatal("Service reminder not found after undo")
	})

	t.Run("undo delete", func(t *testing.T) {
		t.Parallel()

		reminders := []domain.Reminder{
			makeReminder("Deleted", 5),
		}
		m := newTestModel(t, reminders)

		// Delete the reminder.
		m = sendKey(t, m, "d")
		m = sendKey(t, m, "d")
		m = sendKey(t, m, "y")

		if len(m.reminders) != 0 {
			t.Fatal("expected 0 reminders after delete")
		}

		// Undo.
		m = sendKey(t, m, "u")

		if len(m.reminders) != 1 {
			t.Fatalf("expected 1 reminder after undo, got %d", len(m.reminders))
		}
		if m.reminders[0].Name != "Deleted" {
			t.Fatalf("expected restored reminder to be Deleted, got %q", m.reminders[0].Name)
		}
	})

	t.Run("undo with no action is noop", func(t *testing.T) {
		t.Parallel()

		reminders := []domain.Reminder{
			makeReminder("Stable", 5),
		}
		m := newTestModel(t, reminders)

		before := len(m.reminders)
		m = sendKey(t, m, "u")
		if len(m.reminders) != before {
			t.Fatal("undo without action should not change reminders")
		}
	})
}

func TestHelpViewMode(t *testing.T) {
	t.Parallel()

	t.Run("question mark toggles help", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t, nil)

		// Open help.
		m = sendKey(t, m, "?")
		if m.viewMode != ViewHelp {
			t.Fatalf("expected ViewHelp, got %d", m.viewMode)
		}

		// Close help with ?.
		m = sendKey(t, m, "?")
		if m.viewMode != ViewList {
			t.Fatalf("expected ViewList, got %d", m.viewMode)
		}
	})

	t.Run("esc closes help", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t, nil)

		m = sendKey(t, m, "?")
		m = sendKey(t, m, "esc")
		if m.viewMode != ViewList {
			t.Fatalf("expected ViewList after esc, got %d", m.viewMode)
		}
	})

	t.Run("q closes help", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t, nil)

		m = sendKey(t, m, "?")
		m = sendKey(t, m, "q")
		if m.viewMode != ViewList {
			t.Fatalf("expected ViewList after q, got %d", m.viewMode)
		}
	})
}

func TestWindowSizeMsg(t *testing.T) {
	t.Parallel()

	m := newTestModel(t, nil)

	result, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
	m = result.(Model)

	if m.width != 200 {
		t.Fatalf("expected width 200, got %d", m.width)
	}
	if m.height != 60 {
		t.Fatalf("expected height 60, got %d", m.height)
	}
}

func TestFilteredReminders(t *testing.T) {
	t.Parallel()

	r1 := makeReminder("Netflix", 5)
	r1.Tags = []string{"streaming"}
	r2 := makeReminder("Spotify", 10)
	r2.Tags = []string{"music", "streaming"}
	r3 := makeReminder("VPN", 15)
	r3.Tags = []string{"security"}

	reminders := []domain.Reminder{r1, r2, r3}

	t.Run("filter by single tag", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t, reminders)
		m.activeFilters = []string{"streaming"}
		filtered := m.filteredReminders()
		if len(filtered) != 2 {
			t.Fatalf("expected 2 reminders with streaming tag, got %d", len(filtered))
		}
	})

	t.Run("filter by tag not present", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t, reminders)
		m.activeFilters = []string{"nonexistent"}
		filtered := m.filteredReminders()
		if len(filtered) != 0 {
			t.Fatalf("expected 0 reminders with nonexistent tag, got %d", len(filtered))
		}
	})
}

func TestTableRowCount(t *testing.T) {
	t.Parallel()

	reminders := []domain.Reminder{
		makeReminder("A", 1),
		makeReminder("B", 2),
		makeReminder("C", 3),
	}

	m := newTestModel(t, reminders)
	rows := m.table.Rows()

	if len(rows) != 3 {
		t.Fatalf("expected 3 table rows, got %d", len(rows))
	}
}

func TestReminderToRow(t *testing.T) {
	t.Parallel()

	r := makeReminder("Test", 5)
	r.Tags = []string{"streaming"}
	r.URL = "https://example.com"

	row := reminderToRow(r)

	if len(row.Cells) != 6 {
		t.Fatalf("expected row with 6 cells, got %d", len(row.Cells))
	}

	// Last cell should have link icon since URL is set
	if row.Cells[5] == "" {
		t.Fatal("expected link icon in last cell for reminder with URL")
	}
}
