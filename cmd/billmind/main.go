package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
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
	fmt.Println("daemon: not implemented yet")
}

func installDaemon() {
	fmt.Println("install: not implemented yet")
}

func uninstallDaemon() {
	fmt.Println("uninstall: not implemented yet")
}
