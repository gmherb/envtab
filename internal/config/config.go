package config

import (
	"log/slog"
	"os"
	"path/filepath"
)

const (
	envtabDir = "envtab"
)

// getHomeDir returns the user's home directory, exiting on error.
func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failure getting user's home directory", "error", err)
		os.Exit(1)
	}
	return home
}

// getXDGDir returns an XDG directory path, using defaults if the env var is not set.
func getXDGDir(envVar, defaultSubdir string) string {
	if dir := os.Getenv(envVar); dir != "" {
		return dir
	}
	return filepath.Join(getHomeDir(), defaultSubdir)
}

// getXDGDataHome returns the XDG data home directory, using defaults if not set.
func getXDGDataHome() string {
	return getXDGDir("XDG_DATA_HOME", ".local/share")
}

// GetEnvtabPath returns the path to the envtab data directory
// Priority: 1. ENVTAB_DIR env var, 2. XDG_DATA_HOME/envtab (with defaults)
func GetEnvtabPath() string {
	// Check ENVTAB_DIR environment variable first (overrides mode)
	if envDir := os.Getenv("ENVTAB_DIR"); envDir != "" {
		return envDir
	}

	// Use XDG with defaults
	return filepath.Join(getXDGDataHome(), envtabDir)
}

// GetUserConfigPath returns the path to the user config file
// Returns $XDG_CONFIG_HOME/envtab/envtab.yaml (with defaults)
func GetUserConfigPath() string {
	xdgConfigHome := getXDGDir("XDG_CONFIG_HOME", ".config")
	return filepath.Join(xdgConfigHome, envtabDir, "envtab.yaml")
}

// FindProjectConfig walks up the directory tree from the current working directory
// to find .envtab.yaml. Returns the path if found, empty string otherwise.
func FindProjectConfig() string {
	cwd, err := os.Getwd()
	if err != nil {
		slog.Debug("failure getting current working directory", "error", err)
		return ""
	}

	dir := cwd
	for {
		configPath := filepath.Join(dir, ".envtab.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return ""
}

// createDir creates a directory if it doesn't exist, returning the path.
func createDir(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0700); err != nil {
			slog.Error("failure creating directory", "path", path, "error", err)
			os.Exit(1)
		}
	}
	return path
}

// InitEnvtab creates the envtab directory if it doesn't exist and returns the path.
// If path is empty, uses the default envtab directory from GetEnvtabPath().
func InitEnvtab(path string) string {
	if path != "" {
		return createDir(path)
	}

	envtabPath := GetEnvtabPath()
	return createDir(envtabPath)
}

// GetTmpPath returns the path to the tmp directory and ensures it exists.
func GetTmpPath() string {
	xdgCacheHome := getXDGDir("XDG_CACHE_HOME", ".cache")
	tmpPath := filepath.Join(xdgCacheHome, envtabDir, "tmp")

	// Create tmp directory
	if err := os.MkdirAll(tmpPath, 0700); err != nil {
		slog.Error("failure creating tmp directory", "path", tmpPath, "error", err)
		os.Exit(1)
	}

	return tmpPath
}
