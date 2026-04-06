package platform

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/pkg/browser"
)

type windowsPlatform struct{}

func newWindowsPlatform() Platform {
	return &windowsPlatform{}
}

func (p *windowsPlatform) OpenURL(_ context.Context, url string) error {
	return browser.OpenURL(url)
}

func (p *windowsPlatform) SendNotification(_ context.Context, title, body string) error {
	return beeep.Notify(title, body, "")
}

func (p *windowsPlatform) Scheduler() Scheduler {
	return &schtasksScheduler{}
}

func (p *windowsPlatform) SecretStore() SecretStore {
	return &noopSecretStore{}
}

// --- schtasks Scheduler ---

type schtasksScheduler struct{}

func (s *schtasksScheduler) Available() bool {
	_, err := exec.LookPath("schtasks")
	return err == nil
}

func (s *schtasksScheduler) Install(cfg ScheduleConfig) error {
	minutes := int(cfg.Interval / time.Minute)
	if minutes < 1 {
		minutes = 1
	}

	args := strings.Join(append([]string{cfg.BinaryPath}, cfg.Args...), " ")

	cmd := exec.Command("schtasks", "/create", "/f",
		"/tn", cfg.Label,
		"/tr", args,
		"/sc", "minute",
		"/mo", fmt.Sprintf("%d", minutes),
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks create: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (s *schtasksScheduler) Uninstall() error {
	cmd := exec.Command("schtasks", "/delete", "/f", "/tn", "com.billmind.daemon")
	_ = cmd.Run()
	return nil
}

func (s *schtasksScheduler) IsInstalled() (bool, error) {
	cmd := exec.Command("schtasks", "/query", "/tn", "com.billmind.daemon")
	err := cmd.Run()
	return err == nil, nil
}
