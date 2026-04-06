package platform

import (
	"context"
	"fmt"

	"github.com/pkg/browser"
)

type fallbackPlatform struct{}

func newFallbackPlatform() *fallbackPlatform {
	return &fallbackPlatform{}
}

func (p *fallbackPlatform) OpenURL(_ context.Context, url string) error {
	return browser.OpenURL(url)
}

func (p *fallbackPlatform) SendNotification(_ context.Context, _, _ string) error {
	return nil // silent noop
}

func (p *fallbackPlatform) Scheduler() Scheduler {
	return &noopScheduler{}
}

func (p *fallbackPlatform) SecretStore() SecretStore {
	return &noopSecretStore{}
}

type noopScheduler struct{}

func (s *noopScheduler) Available() bool                { return false }
func (s *noopScheduler) Install(_ ScheduleConfig) error { return fmt.Errorf("scheduler not available on this platform") }
func (s *noopScheduler) Uninstall() error               { return fmt.Errorf("scheduler not available on this platform") }
func (s *noopScheduler) IsInstalled() (bool, error)     { return false, nil }

type noopSecretStore struct{}

func (s *noopSecretStore) Set(_ context.Context, _, _ string) error        { return fmt.Errorf("secret store not available on this platform") }
func (s *noopSecretStore) Get(_ context.Context, _ string) (string, error) { return "", fmt.Errorf("secret store not available on this platform") }
func (s *noopSecretStore) Delete(_ context.Context, _ string) error        { return fmt.Errorf("secret store not available on this platform") }
