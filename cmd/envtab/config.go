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

type EnvMetadata struct {
	CreatedAt string   `json:"createdAt" yaml:"createdAt"`
	LoadedAt  string   `json:"loadedAt" yaml:"loadedAt"`
	UpdatedAt string   `json:"updatedAt" yaml:"updatedAt"`
	Login     bool     `json:"login" yaml:"login"`
	Tags      []string `json:"tags" yaml:"tags"`
}

// EnvTable represents the structure of an envtab loadout
type EnvTable struct {
	Metadata EnvMetadata       `json:"metadata" yaml:"metadata"`
	Entries  map[string]string `json:"entries" yaml:"entries"`
}

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
func InitEnvtab() string {
	envtabPath := getEnvtabPath()
	if _, err := os.Stat(envtabPath); os.IsNotExist(err) {
		os.Mkdir(envtabPath, 0700)
	}
	return envtabPath
}

// Find all YAML files in the envtab directory, remove the extension, and return them as a slice
func getEnvtabSlice() []string {
	envtabPath := InitEnvtab()

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
