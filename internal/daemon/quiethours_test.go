package daemon

import (
	"testing"
	"time"
)

func TestIsQuietHours(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		hour  int
		start int
		end   int
		want  bool
	}{
		{"23:00 in default 22-08", 23, 0, 0, true},
		{"03:00 in default 22-08", 3, 0, 0, true},
		{"08:00 in default 22-08", 8, 0, 0, false},
		{"21:00 in default 22-08", 21, 0, 0, false},
		{"12:00 in default 22-08", 12, 0, 0, false},
		{"22:00 in default 22-08", 22, 0, 0, true},
		{"custom 01-06 at 03:00", 3, 1, 6, true},
		{"custom 01-06 at 07:00", 7, 1, 6, false},
		{"custom 20-06 at 23:00", 23, 20, 6, true},
		{"custom 20-06 at 04:00", 4, 20, 6, true},
		{"custom 20-06 at 10:00", 10, 20, 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			now := time.Date(2026, 4, 6, tt.hour, 30, 0, 0, time.UTC)
			got := IsQuietHours(now, tt.start, tt.end)
			if got != tt.want {
				t.Errorf("IsQuietHours(hour=%d, %d-%d) = %v, want %v",
					tt.hour, tt.start, tt.end, got, tt.want)
			}
		})
	}
}
