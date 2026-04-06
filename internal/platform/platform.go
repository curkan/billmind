package platform

import (
	"context"
	"runtime"
	"time"
)

type Platform interface {
	OpenURL(ctx context.Context, url string) error
	SendNotification(ctx context.Context, title, body string) error
	Scheduler() Scheduler
	SecretStore() SecretStore
}

type ScheduleConfig struct {
	BinaryPath string
	Args       []string
	Interval   time.Duration
	Label      string
}

type Scheduler interface {
	Available() bool
	Install(cfg ScheduleConfig) error
	Uninstall() error
	IsInstalled() (bool, error)
}

type SecretStore interface {
	Set(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

func New() Platform {
	switch runtime.GOOS {
	case "darwin":
		return newDarwinPlatform()
	case "linux":
		return newLinuxPlatform()
	case "windows":
		return newWindowsPlatform()
	default:
		return newFallbackPlatform()
	}
}
