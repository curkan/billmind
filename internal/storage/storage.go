package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/curkan/billmind/internal/domain"
)

const (
	dataFileName  = "data.json"
	backupDir     = "backups"
	maxBackups    = 10
	dirPerm       = 0o755
	filePerm      = 0o644
	backupTimeFmt = "2006-01-02_150405"
	backupPrefix  = "data_"
	backupSuffix  = ".json"
)

// dataFile is the top-level JSON structure persisted to disk.
type dataFile struct {
	Reminders []domain.Reminder `json:"reminders"`
	Settings  Settings          `json:"settings"`
}

// Settings holds user-level configuration.
type Settings struct {
	Email    string `json:"email"`
	Language string `json:"language"`
}

// Storage manages persistence of reminders and settings to the file system.
type Storage struct {
	fs       FileSystem
	basePath string
}

// New creates a new Storage instance with the given file system and base directory path.
func New(fs FileSystem, basePath string) *Storage {
	return &Storage{
		fs:       fs,
		basePath: basePath,
	}
}

// DefaultPath returns the default storage directory path (~/.config/billmind).
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "billmind")
}

// dataFilePath returns the full path to data.json.
func (s *Storage) dataFilePath() string {
	return filepath.Join(s.basePath, dataFileName)
}

// backupDirPath returns the full path to the backups directory.
func (s *Storage) backupDirPath() string {
	return filepath.Join(s.basePath, backupDir)
}

// Load reads and parses the data file, returning the stored reminders.
// Returns an empty slice if the file does not exist.
// Returns domain.ErrStorageCorrupted if the file contains invalid JSON.
func (s *Storage) Load(_ context.Context) ([]domain.Reminder, error) {
	data, err := s.fs.ReadFile(s.dataFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []domain.Reminder{}, nil
		}
		return nil, fmt.Errorf("reading data file: %w", err)
	}

	var df dataFile
	if err := json.Unmarshal(data, &df); err != nil {
		return nil, fmt.Errorf("parsing data file: %w", domain.ErrStorageCorrupted)
	}

	if df.Reminders == nil {
		return []domain.Reminder{}, nil
	}

	return df.Reminders, nil
}

// Save persists the given reminders to disk. It creates a backup of the
// current data file before writing, and uses atomic write (temp file + rename)
// to prevent data corruption.
func (s *Storage) Save(ctx context.Context, reminders []domain.Reminder) error {
	if err := s.fs.MkdirAll(s.basePath, dirPerm); err != nil {
		return fmt.Errorf("creating storage directory: %w", err)
	}

	if err := s.Backup(ctx); err != nil {
		// Backup errors are non-blocking warnings; we proceed.
		_ = err
	}

	df := dataFile{
		Reminders: reminders,
		Settings:  Settings{},
	}

	// Try to preserve existing settings.
	existing, readErr := s.fs.ReadFile(s.dataFilePath())
	if readErr == nil {
		var existingDF dataFile
		if json.Unmarshal(existing, &existingDF) == nil {
			df.Settings = existingDF.Settings
		}
	}

	return s.writeDataFile(df)
}

// LoadSettings reads and returns the settings from the data file.
// Returns an empty Settings if the file does not exist.
func (s *Storage) LoadSettings(_ context.Context) (Settings, error) {
	data, err := s.fs.ReadFile(s.dataFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Settings{}, nil
		}
		return Settings{}, fmt.Errorf("reading data file: %w", err)
	}

	var df dataFile
	if err := json.Unmarshal(data, &df); err != nil {
		return Settings{}, fmt.Errorf("parsing data file: %w", domain.ErrStorageCorrupted)
	}

	return df.Settings, nil
}

// SaveSettings loads the full data file, updates settings, and saves back.
func (s *Storage) SaveSettings(_ context.Context, settings Settings) error {
	if err := s.fs.MkdirAll(s.basePath, dirPerm); err != nil {
		return fmt.Errorf("creating storage directory: %w", err)
	}

	df := dataFile{
		Settings: settings,
	}

	// Preserve existing data.
	existing, readErr := s.fs.ReadFile(s.dataFilePath())
	if readErr == nil {
		var existingDF dataFile
		if json.Unmarshal(existing, &existingDF) == nil {
			df.Reminders = existingDF.Reminders
			// Merge: keep existing email if new settings don't have one
			if settings.Email == "" && existingDF.Settings.Email != "" {
				df.Settings.Email = existingDF.Settings.Email
			}
		}
	}

	return s.writeDataFile(df)
}

// writeDataFile atomically writes the data file to disk.
func (s *Storage) writeDataFile(df dataFile) error {
	encoded, err := json.MarshalIndent(df, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding data: %w", err)
	}

	// Atomic write: create temp file in the same directory, write, then rename.
	tmpFile, err := s.fs.CreateTemp(s.basePath, "data-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(encoded); err != nil {
		tmpFile.Close()
		s.fs.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		s.fs.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := s.fs.Rename(tmpPath, s.dataFilePath()); err != nil {
		s.fs.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// Backup creates a timestamped copy of the current data file in the backups
// directory. It keeps at most 10 backups, deleting the oldest ones.
// If the data file does not exist, Backup is a no-op.
func (s *Storage) Backup(_ context.Context) error {
	_, err := s.fs.Stat(s.dataFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("checking data file: %w", err)
	}

	data, err := s.fs.ReadFile(s.dataFilePath())
	if err != nil {
		return fmt.Errorf("reading data file for backup: %w", err)
	}

	backupPath := s.backupDirPath()
	if err := s.fs.MkdirAll(backupPath, dirPerm); err != nil {
		return fmt.Errorf("creating backup directory: %w", err)
	}

	timestamp := time.Now().Format(backupTimeFmt)
	backupFile := filepath.Join(backupPath, backupPrefix+timestamp+backupSuffix)

	if err := s.fs.WriteFile(backupFile, data, filePerm); err != nil {
		return fmt.Errorf("writing backup file: %w", err)
	}

	if err := s.rotateBackups(backupPath); err != nil {
		return fmt.Errorf("rotating backups: %w", err)
	}

	return nil
}

// rotateBackups keeps only the most recent maxBackups backup files,
// deleting any older ones.
func (s *Storage) rotateBackups(backupPath string) error {
	entries, err := s.fs.ReadDir(backupPath)
	if err != nil {
		return fmt.Errorf("reading backup directory: %w", err)
	}

	var backups []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), backupPrefix) && strings.HasSuffix(e.Name(), backupSuffix) {
			backups = append(backups, e)
		}
	}

	if len(backups) <= maxBackups {
		return nil
	}

	// Sort by name ascending (timestamps in name ensure chronological order).
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Name() < backups[j].Name()
	})

	toDelete := backups[:len(backups)-maxBackups]
	for _, b := range toDelete {
		path := filepath.Join(backupPath, b.Name())
		if err := s.fs.Remove(path); err != nil {
			return fmt.Errorf("removing old backup %s: %w", b.Name(), err)
		}
	}

	return nil
}
