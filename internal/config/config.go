package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	ENVTAB_DIR = ".envtab"
)

// GetEnvtabPath returns the path to the envtab directory
// Checks viper config first (supports ENVTAB_DIR env var and config file), then defaults to ~/.envtab
func GetEnvtabPath() string {
	if viper.IsSet("dir") {
		return viper.GetString("dir")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failure getting user's home directory", "error", err)
		os.Exit(1)
	}

	return filepath.Join(home, ENVTAB_DIR)
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
