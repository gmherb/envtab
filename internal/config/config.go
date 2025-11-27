package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const (
	ENVTAB_DIR = ".envtab"
)

// Get the path to the envtab directory
func getEnvtabPath() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting user's home directory: %s\n", err)
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
