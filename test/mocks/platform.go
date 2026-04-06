package mocks

import (
	"context"
	"sync"

	"github.com/curkan/billmind/internal/platform"
)

type Notification struct {
	Title string
	Body  string
}

type MockPlatform struct {
	mu            sync.Mutex
	Notifications []Notification
	OpenedURLs    []string
}

func (m *MockPlatform) SendNotification(_ context.Context, title, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Notifications = append(m.Notifications, Notification{Title: title, Body: body})
	return nil
}

func (m *MockPlatform) OpenURL(_ context.Context, url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OpenedURLs = append(m.OpenedURLs, url)
	return nil
}

func (m *MockPlatform) Scheduler() platform.Scheduler {
	return &MockScheduler{}
}

func (m *MockPlatform) SecretStore() platform.SecretStore {
	return &MockSecretStore{}
}

type MockScheduler struct {
	installed bool
}

func (s *MockScheduler) Available() bool                          { return true }
func (s *MockScheduler) Install(_ platform.ScheduleConfig) error  { s.installed = true; return nil }
func (s *MockScheduler) Uninstall() error                         { s.installed = false; return nil }
func (s *MockScheduler) IsInstalled() (bool, error)               { return s.installed, nil }

type MockSecretStore struct {
	mu   sync.Mutex
	data map[string]string
}

func (s *MockSecretStore) Set(_ context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data == nil {
		s.data = make(map[string]string)
	}
	s.data[key] = value
	return nil
}

func (s *MockSecretStore) Get(_ context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[key], nil
}

func (s *MockSecretStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}
