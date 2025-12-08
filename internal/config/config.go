package config

import (
	"log/slog"
	"os"
	"path/filepath"
)

const (
	ENVTAB_DIR = ".envtab"
)

// getXDGDataHome returns the XDG data home directory, using defaults if not set.
func getXDGDataHome() (string, error) {
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		xdgDataHome = filepath.Join(home, ".local", "share")
	}
	return xdgDataHome, nil
}

// getPOSIXDataHome returns the POSIX fallback data directory.
func getPOSIXDataHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ENVTAB_DIR), nil
}

// GetEnvtabPath returns the path to the envtab data directory
// Priority: 1. ENVTAB_DIR env var, 2. XDG_DATA_HOME/envtab (with defaults), 3. ~/.envtab (POSIX fallback if XDG fails)
func GetEnvtabPath() string {
	// Check ENVTAB_DIR environment variable first (overrides mode)
	if envDir := os.Getenv("ENVTAB_DIR"); envDir != "" {
		return envDir
	}

	// Always try XDG first (with defaults)
	xdgDataHome, err := getXDGDataHome()
	if err != nil {
		slog.Error("failure getting user's home directory", "error", err)
		os.Exit(1)
	}
	return filepath.Join(xdgDataHome, "envtab")
}

// GetUserConfigPath returns the path to the user config file
// Returns $XDG_CONFIG_HOME/envtab/envtab.yaml (with defaults) or ~/.envtab.yaml (POSIX fallback if XDG fails)
func GetUserConfigPath() string {
	// Always try XDG first (with defaults)
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Error("failure getting user's home directory", "error", err)
			os.Exit(1)
		}
		xdgConfigHome = filepath.Join(home, ".config")
	}
	return filepath.Join(xdgConfigHome, "envtab", "envtab.yaml")
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
// If path is empty, uses the default envtab directory from GetEnvtabPath().
// Tries XDG paths first, falls back to POSIX if XDG directory creation fails.
func InitEnvtab(path string) string {
	var envtabPath string

	if path != "" {
		envtabPath = path
	} else {
		envtabPath = GetEnvtabPath()
	}

	// Try to create/ensure the directory exists
	if _, err := os.Stat(envtabPath); os.IsNotExist(err) {
		if err := os.Mkdir(envtabPath, 0700); err != nil {
			// If XDG path creation fails and we're not already using POSIX, try POSIX fallback
			if path == "" {
				posixPath, posixErr := getPOSIXDataHome()
				if posixErr != nil {
					slog.Error("failure getting user's home directory", "error", posixErr)
					os.Exit(1)
				}
				slog.Debug("XDG directory creation failed, falling back to POSIX", "xdg_path", envtabPath, "posix_path", posixPath, "error", err)
				envtabPath = posixPath
				// Check if POSIX directory already exists before creating
				if _, err := os.Stat(envtabPath); os.IsNotExist(err) {
					if err := os.Mkdir(envtabPath, 0700); err != nil {
						slog.Error("failure creating envtab directory", "path", envtabPath, "error", err)
						os.Exit(1)
					}
				}
			} else {
				slog.Error("failure creating envtab directory", "path", envtabPath, "error", err)
				os.Exit(1)
			}
		}
	}

	return envtabPath
}

// GetTmpPath returns the path to the tmp directory and ensures it exists.
// Tries XDG paths first, falls back to POSIX if XDG directory creation fails.
func GetTmpPath() string {
	// Always try XDG first (with defaults)
	xdgCacheHome := os.Getenv("XDG_CACHE_HOME")
	if xdgCacheHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Error("failure getting user's home directory", "error", err)
			os.Exit(1)
		}
		xdgCacheHome = filepath.Join(home, ".cache")
	}
	tmpPath := filepath.Join(xdgCacheHome, "envtab", "tmp")

	// Try to create XDG tmp directory
	if err := os.MkdirAll(tmpPath, 0700); err != nil {
		// Fallback to POSIX if XDG creation fails
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			slog.Error("failure getting user's home directory", "error", homeErr)
			os.Exit(1)
		}
		posixPath := filepath.Join(home, ENVTAB_DIR, "tmp")
		slog.Debug("XDG tmp directory creation failed, falling back to POSIX", "xdg_path", tmpPath, "posix_path", posixPath, "error", err)
		tmpPath = posixPath
		if err := os.MkdirAll(tmpPath, 0700); err != nil {
			slog.Error("failure creating tmp directory", "path", tmpPath, "error", err)
			os.Exit(1)
		}
	}

	return tmpPath
}
