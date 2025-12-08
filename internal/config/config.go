package config

import (
	"log/slog"
	"os"
	"path/filepath"
)

const (
	ENVTAB_DIR = ".envtab"
)

// GetEnvtabPath returns the path to the envtab data directory
// Priority: 1. ENVTAB_DIR env var, 2. XDG_DATA_HOME/envtab, 3. ~/.envtab
func GetEnvtabPath() string {
	// Check ENVTAB_DIR environment variable first
	if envDir := os.Getenv("ENVTAB_DIR"); envDir != "" {
		return envDir
	}

	// Check XDG_DATA_HOME
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "envtab")
	}

	// Default to ~/.envtab
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failure getting user's home directory", "error", err)
		os.Exit(1)
	}

	return filepath.Join(home, ENVTAB_DIR)
}

// GetUserConfigPath returns the path to the user config file
// Returns ~/.envtab.yaml or $XDG_CONFIG_HOME/envtab/.envtab.yaml if XDG_CONFIG_HOME is set
func GetUserConfigPath() string {
	// Check XDG_CONFIG_HOME
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "envtab", ".envtab.yaml")
	}

	// Default to ~/.envtab.yaml
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failure getting user's home directory", "error", err)
		os.Exit(1)
	}

	return filepath.Join(home, ".envtab.yaml")
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

// InitEnvtab creates the envtab directory if it doesn't exist and returns the path.
// If path is empty, uses the default envtab directory from config or ~/.envtab.
func InitEnvtab(path string) string {
	var envtabPath string

	if path != "" {
		envtabPath = path
	} else {
		envtabPath = GetEnvtabPath()
	}

	if _, err := os.Stat(envtabPath); os.IsNotExist(err) {
		if err := os.Mkdir(envtabPath, 0700); err != nil {
			slog.Error("failure creating envtab directory", "path", envtabPath, "error", err)
			os.Exit(1)
		}
	}

	return envtabPath
}

// GetTmpPath returns the path to the tmp directory and ensures it exists.
// If envtabPath is provided (non-empty), uses it; otherwise determines it.
// Returns ENVTAB_DIR/tmp/
func GetTmpPath(envtabPath ...string) string {
	var path string
	if len(envtabPath) > 0 && envtabPath[0] != "" {
		path = envtabPath[0]
	} else {
		path = InitEnvtab("")
	}

	tmpPath := filepath.Join(path, "tmp")

	// Use MkdirAll which is idempotent - creates directory if it doesn't exist,
	// or does nothing if it already exists. This avoids race conditions.
	if err := os.MkdirAll(tmpPath, 0700); err != nil {
		slog.Error("failure creating tmp directory", "path", tmpPath, "error", err)
		os.Exit(1)
	}

	return tmpPath
}
