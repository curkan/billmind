package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/curkan/billmind/internal/daemon"
	"github.com/curkan/billmind/internal/platform"
	"github.com/curkan/billmind/internal/storage"
	"github.com/curkan/billmind/internal/ui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "daemon":
			runDaemon()
			return
		case "install":
			installDaemon()
			return
		case "uninstall":
			uninstallDaemon()
			return
		}
	}

	store := storage.New(storage.RealFileSystem{}, storage.DefaultPath())
	plat := platform.New()
	model := ui.New(store, plat)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func runDaemon() {
	setupDaemonLog()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	store := storage.New(storage.RealFileSystem{}, storage.DefaultPath())
	plat := platform.New()

	if err := daemon.Run(ctx, store, plat); err != nil {
		log.Printf("daemon error: %v", err)
		os.Exit(1)
	}
}

func installDaemon() {
	plat := platform.New()
	sched := plat.Scheduler()

	if !sched.Available() {
		fmt.Fprintln(os.Stderr, "Scheduler not available on this platform.")
		fmt.Fprintln(os.Stderr, "Add 'billmind daemon' to your crontab manually.")
		os.Exit(1)
	}

	binPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine binary path: %v\n", err)
		os.Exit(1)
	}

	cfg := platform.ScheduleConfig{
		BinaryPath: binPath,
		Args:       []string{"daemon"},
		Interval:   1 * time.Hour,
		Label:      "com.billmind.daemon",
	}

	if err := sched.Install(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Install failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Daemon installed. It will run every hour.")
}

func uninstallDaemon() {
	plat := platform.New()
	sched := plat.Scheduler()

	if err := sched.Uninstall(); err != nil {
		fmt.Fprintf(os.Stderr, "Uninstall failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Daemon uninstalled.")
}

func setupDaemonLog() {
	logDir := storage.DefaultPath()
	_ = os.MkdirAll(logDir, 0o755)

	logPath := filepath.Join(logDir, "daemon.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
