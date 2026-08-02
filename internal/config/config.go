// Package config persists the small amount of application state that should
// survive between viewer sessions.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

const (
	configDirectory = ".config/yaml-viewer"
	configFileName  = "config.yaml"
)

// Config is intentionally small and extensible. The path is only a recent
// file hint; startup does not automatically open it when no CLI path is given.
type Config struct {
	LastOpenedFile string `yaml:"last_opened_file,omitempty"`
}

// Path returns the platform user's requested configuration path:
// ~/.config/yaml-viewer/config.yaml.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(home, configDirectory, configFileName), nil
}

// Load reads the configuration. A missing file is treated as a first launch.
func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return Config{}, nil
	}
	var value Config
	if err := yaml.Unmarshal(data, &value); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return value, nil
}

// Save writes the configuration atomically with user-only file permissions.
// The temporary file lives beside the final file so rename remains atomic on
// the same filesystem.
func Save(value Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create config directory %s: %w", directory, err)
	}
	data, err := yaml.Marshal(&value)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".config.yaml-*")
	if err != nil {
		return fmt.Errorf("create temporary config: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set config permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write config: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync config: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary config: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}
