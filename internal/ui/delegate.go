package ui

import (
	"strings"
)

// highlightMatch wraps the first case-insensitive match of query in text with styleAccent.
func highlightMatch(text, query string) string {
	if query == "" {
		return text
	}
	lower := strings.ToLower(text)
	q := strings.ToLower(query)
	idx := strings.Index(lower, q)
	if idx == -1 {
		return text
	}
	before := text[:idx]
	match := text[idx : idx+len(query)]
	after := text[idx+len(query):]
	return before + styleAccent.Render(match) + after
}
