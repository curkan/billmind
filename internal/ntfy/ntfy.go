package ntfy

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const defaultServer = "https://ntfy.sh"

// Send publishes a notification to a ntfy.sh topic.
func Send(ctx context.Context, topic, title, body string, priority Priority) error {
	if topic == "" {
		return fmt.Errorf("ntfy: topic is empty")
	}

	url := defaultServer + "/" + topic

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("ntfy: creating request: %w", err)
	}

	req.Header.Set("Title", title)
	req.Header.Set("Priority", string(priority))
	req.Header.Set("Tags", "money_with_wings")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ntfy: sending: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy: unexpected status %d", resp.StatusCode)
	}

	return nil
}

// Priority levels for ntfy messages.
type Priority string

const (
	PriorityMin     Priority = "min"
	PriorityLow     Priority = "low"
	PriorityDefault Priority = "default"
	PriorityHigh    Priority = "high"
	PriorityUrgent  Priority = "urgent"
)
