package config

import (
	"log/slog"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	ENVTAB_DIR = ".envtab"
)

// Get the path to the envtab directory
func getEnvtabPath() string {
	// Try to get from Viper config first
	if viper.IsSet("envtab_dir") {
		return viper.GetString("envtab_dir")
	}

	// Fall back to environment variable
	if envPath := os.Getenv("ENVTAB_DIR"); envPath != "" {
		return envPath
	}

	// Default to home directory
	usr, err := user.Current()
	if err != nil {
		slog.Error("failure getting user's home directory", "error", err)
		os.Exit(1)
	}

	return filepath.Join(usr.HomeDir, ENVTAB_DIR)
}

// Create the envtab directory if it doesn't exist and return the path
func InitEnvtab(path string) string {
	var envtabPath string

	if path == "" {
		envtabPath = getEnvtabPath()
	} else {
		envtabPath = path
	}

	if _, err := os.Stat(envtabPath); os.IsNotExist(err) {
		os.Mkdir(envtabPath, 0700)
	}
	return envtabPath
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	if viper.ConfigFileUsed() != "" {
		return viper.ConfigFileUsed()
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ENVTAB_DIR, ".envtab.yaml")
}
