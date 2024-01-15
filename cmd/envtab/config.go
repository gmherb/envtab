package envtab

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	tagz "github.com/gmherb/envtab/pkg/tags"
	"github.com/gmherb/envtab/pkg/utils"
	yaml "gopkg.in/yaml.v2"
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

	var entries []string
	err := filepath.Walk(envtabPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".yaml" {
			entries = append(entries, filepath.Base(path[:len(path)-5]))
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error reading envtab loadout %s: %s\n", envtabPath, err)
		os.Exit(1)
	}

	return entries
}

// Print all envtab loadouts
func PrintEnvtabLoadouts() {
	entries := getEnvtabSlice()
	for _, entry := range entries {
		fmt.Println(entry)
	}
}

// Print all envtab entries in a loadout and its metadata
func ReadLoadout(name string) (*EnvTable, error) {

	filePath := filepath.Join(InitEnvtab(), name+".yaml")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var entry EnvTable
	err = yaml.Unmarshal(content, &entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil

}

// Write a key-value pair to a loadout (and optionally add tags) and update the metadata
func WriteEntryToLoadout(name, key, value string, tags []string) error {

	filePath := filepath.Join(InitEnvtab(), name+".yaml")

	// Read the existing entries if file exists
	content, err := ReadLoadout(name)
	if err != nil && !os.IsNotExist(err) {
		return err

		// Create a new file if it doesn't exist
	} else if os.IsNotExist(err) {
		content = &EnvTable{
			Metadata: EnvMetadata{
				CreatedAt: utils.GetCurrentTime(),
				LoadedAt:  utils.GetCurrentTime(),
				UpdatedAt: utils.GetCurrentTime(),
				Login:     false,
				Tags:      []string{},
			},
			Entries: map[string]string{},
		}
	}

	// Update or add the new key-value pair
	content.Entries[key] = value

	// Append unique tags to the existing list
	content.Metadata.Tags = tagz.MergeTags(content.Metadata.Tags, tags)

	// Update metadata
	content.Metadata.UpdatedAt = utils.GetCurrentTime()

	// Write the updated entries to the file
	data, err := yaml.Marshal(content)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0700)
	if err != nil {
		return err
	}

	return nil
}
