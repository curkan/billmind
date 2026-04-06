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

type darwinPlatform struct{}

func newDarwinPlatform() Platform {
	return &darwinPlatform{}
}

func (p *darwinPlatform) OpenURL(_ context.Context, url string) error {
	return browser.OpenURL(url)
}

func (p *darwinPlatform) SendNotification(ctx context.Context, title, body string) error {
	if path, err := exec.LookPath("terminal-notifier"); err == nil {
		return p.sendViaTerminalNotifier(ctx, path, title, body)
	}
	return beeep.Notify(title, body, "")
}

func (p *darwinPlatform) sendViaTerminalNotifier(ctx context.Context, binPath, title, body string) error {
	args := []string{
		"-title", title,
		"-message", body,
		"-sound", "default",
		"-group", "billmind",
	}

	// Add "Show" action button that opens billmind TUI
	if selfPath, err := os.Executable(); err == nil {
		termApp := detectTerminal()
		script := fmt.Sprintf(`tell application %q to do script "%s"`, termApp, selfPath)
		args = append(args, "-execute", fmt.Sprintf(`osascript -e '%s'`, script))
	}

	cmd := exec.CommandContext(ctx, binPath, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("terminal-notifier: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// detectTerminal returns the name of the user's terminal app.
func detectTerminal() string {
	termProgram := os.Getenv("TERM_PROGRAM")
	switch termProgram {
	case "iTerm.app":
		return "iTerm"
	case "WezTerm":
		return "WezTerm"
	case "Alacritty":
		return "Alacritty"
	default:
		return "Terminal"
	}
}

func (p *darwinPlatform) Scheduler() Scheduler {
	return &launchdScheduler{}
}

func (p *darwinPlatform) SecretStore() SecretStore {
	return &noopSecretStore{} // TODO: implement with 99designs/keyring
}

// --- launchd Scheduler ---

type launchdScheduler struct{}

func (s *launchdScheduler) Available() bool {
	_, err := exec.LookPath("launchctl")
	return err == nil
}

func (s *launchdScheduler) plistPath(label string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist")
}

func (s *launchdScheduler) Install(cfg ScheduleConfig) error {
	intervalSec := int(cfg.Interval / time.Second)
	if intervalSec < 60 {
		intervalSec = 60
	}

	args := ""
	for _, a := range cfg.Args {
		args += fmt.Sprintf("\n\t\t<string>%s</string>", a)
	}

	home, _ := os.UserHomeDir()
	logPath := filepath.Join(home, ".config", "billmind", "daemon.log")

	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>%s
	</array>
	<key>StartInterval</key>
	<integer>%d</integer>
	<key>RunAtLoad</key>
	<true/>
	<key>StandardOutPath</key>
	<string>%s</string>
	<key>StandardErrorPath</key>
	<string>%s</string>
</dict>
</plist>`, cfg.Label, cfg.BinaryPath, args, intervalSec, logPath, logPath)

	path := s.plistPath(cfg.Label)

	// Unload existing if present
	if _, err := os.Stat(path); err == nil {
		_ = exec.Command("launchctl", "bootout", fmt.Sprintf("gui/%d", os.Getuid()), path).Run()
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating LaunchAgents dir: %w", err)
	}

	if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
		return fmt.Errorf("writing plist: %w", err)
	}

	cmd := exec.Command("launchctl", "bootstrap", fmt.Sprintf("gui/%d", os.Getuid()), path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl bootstrap: %s: %w", strings.TrimSpace(string(out)), err)
	}

	return nil
}

func (s *launchdScheduler) Uninstall() error {
	label := "com.billmind.daemon"
	path := s.plistPath(label)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	_ = exec.Command("launchctl", "bootout", fmt.Sprintf("gui/%d", os.Getuid()), path).Run()
	return os.Remove(path)
}

func (s *launchdScheduler) IsInstalled() (bool, error) {
	path := s.plistPath("com.billmind.daemon")
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
