package envtab

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

// Find all YAML files in the envtab directory, remove the extension, and return them as a slice
func GetEnvtabSlice(path string) []string {
	var envtabPath string

	if path == "" {
		envtabPath = getEnvtabPath()
	} else {
		envtabPath = InitEnvtab(path)
	}

	var loadouts []string
	err := filepath.Walk(envtabPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".yaml" {
			loadouts = append(loadouts, filepath.Base(path[:len(path)-5]))
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error reading envtab loadout %s: %s\n", envtabPath, err)
		os.Exit(1)
	}

	return loadouts
}
