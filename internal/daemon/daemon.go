package daemon

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/curkan/billmind/internal/platform"
	"github.com/curkan/billmind/internal/storage"
)

// Run is the daemon entry point. Called by `billmind daemon`.
// It loads reminders, checks which need notification, sends them,
// and persists updated NotifyStage values.
func Run(ctx context.Context, store *storage.Storage, plat platform.Platform) error {
	now := time.Now()

	settings, err := store.LoadSettings(ctx)
	if err != nil {
		return fmt.Errorf("loading settings: %w", err)
	}

	if IsQuietHours(now, settings.QuietHoursStart, settings.QuietHoursEnd) {
		log.Println("quiet hours — skipping")
		return nil
	}

	reminders, err := store.Load(ctx)
	if err != nil {
		return fmt.Errorf("loading reminders: %w", err)
	}

	groups := GroupByStage(reminders, now)

	if groups.Empty() {
		log.Println("no notifications to send")
		return nil
	}

	if err := SendGrouped(ctx, plat, settings.NtfyTopic, groups); err != nil {
		return fmt.Errorf("sending notifications: %w", err)
	}

	ApplyStages(reminders, groups)

	if err := store.Save(ctx, reminders); err != nil {
		return fmt.Errorf("saving updated stages: %w", err)
	}

	log.Printf("sent: %d soft, %d urgent, %d critical",
		len(groups.Soft), len(groups.Urgent), len(groups.Critical))

	return nil
}
