package platform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/pkg/browser"
)

type linuxPlatform struct{}

func newLinuxPlatform() Platform {
	return &linuxPlatform{}
}

func (p *linuxPlatform) OpenURL(_ context.Context, url string) error {
	return browser.OpenURL(url)
}

func (p *linuxPlatform) SendNotification(_ context.Context, title, body string) error {
	return beeep.Notify(title, body, "")
}

func (p *linuxPlatform) Scheduler() Scheduler {
	return &systemdScheduler{}
}

func (p *linuxPlatform) SecretStore() SecretStore {
	return &noopSecretStore{}
}

// --- systemd user Scheduler ---

type systemdScheduler struct{}

func (s *systemdScheduler) Available() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}

func (s *systemdScheduler) configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user")
}

func (s *systemdScheduler) Install(cfg ScheduleConfig) error {
	dir := s.configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating systemd user dir: %w", err)
	}

	args := strings.Join(append([]string{cfg.BinaryPath}, cfg.Args...), " ")
	home, _ := os.UserHomeDir()
	logPath := filepath.Join(home, ".config", "billmind", "daemon.log")

	service := fmt.Sprintf(`[Unit]
Description=BillMind reminder daemon

[Service]
Type=oneshot
ExecStart=%s
StandardOutput=append:%s
StandardError=append:%s
`, args, logPath, logPath)

	if err := os.WriteFile(filepath.Join(dir, cfg.Label+".service"), []byte(service), 0o644); err != nil {
		return fmt.Errorf("writing service unit: %w", err)
	}

	intervalSec := int(cfg.Interval / time.Second)
	if intervalSec < 60 {
		intervalSec = 60
	}

	timer := fmt.Sprintf(`[Unit]
Description=BillMind reminder timer

[Timer]
OnBootSec=60
OnUnitActiveSec=%ds
Persistent=true

[Install]
WantedBy=timers.target
`, intervalSec)

	if err := os.WriteFile(filepath.Join(dir, cfg.Label+".timer"), []byte(timer), 0o644); err != nil {
		return fmt.Errorf("writing timer unit: %w", err)
	}

	if out, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("daemon-reload: %s: %w", strings.TrimSpace(string(out)), err)
	}
	if out, err := exec.Command("systemctl", "--user", "enable", "--now", cfg.Label+".timer").CombinedOutput(); err != nil {
		return fmt.Errorf("enable timer: %s: %w", strings.TrimSpace(string(out)), err)
	}

	return nil
}

func (s *systemdScheduler) Uninstall() error {
	label := "com.billmind.daemon"
	_ = exec.Command("systemctl", "--user", "disable", "--now", label+".timer").Run()
	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	dir := s.configDir()
	_ = os.Remove(filepath.Join(dir, label+".service"))
	_ = os.Remove(filepath.Join(dir, label+".timer"))
	return nil
}

func (s *systemdScheduler) IsInstalled() (bool, error) {
	out, err := exec.Command("systemctl", "--user", "is-enabled", "com.billmind.daemon.timer").CombinedOutput()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) == "enabled", nil
}
