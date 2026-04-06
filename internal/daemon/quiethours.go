package daemon

import "time"

const (
	defaultQuietStart = 22
	defaultQuietEnd   = 8
)

// IsQuietHours returns true if the given time falls within quiet hours.
// quietStart/quietEnd are hours (0-23). If both are 0, defaults (22-08) are used.
// Handles overnight wrap (e.g. 22:00 -> 08:00).
func IsQuietHours(now time.Time, quietStart, quietEnd int) bool {
	if quietStart == 0 && quietEnd == 0 {
		quietStart = defaultQuietStart
		quietEnd = defaultQuietEnd
	}

	hour := now.Hour()

	if quietStart > quietEnd {
		// Overnight: e.g. 22-08 means quiet from 22:00 to 07:59
		return hour >= quietStart || hour < quietEnd
	}

	// Same-day: e.g. 13-15 means quiet from 13:00 to 14:59
	return hour >= quietStart && hour < quietEnd
}
