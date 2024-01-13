package envtab

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	tagz "github.com/gmherb/envtab/pkg/tags"
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

type EnvTable struct {
	Metadata EnvMetadata       `json:"metadata" yaml:"metadata"`
	Entries  map[string]string `json:"entries" yaml:"entries"`
}

func (c *EnvTable) Save() error {
	return nil
}

func getCurrentTime() string {
	return fmt.Sprintf("%s", time.Now().Format(time.RFC3339))
}

func getEnvtabPath() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting user's home directory: %s\n", err)
		os.Exit(1)
	}

	return filepath.Join(usr.HomeDir, ENVTAB_DIR)
}

func InitEnvtab() string {
	envtabPath := getEnvtabPath()
	if _, err := os.Stat(envtabPath); os.IsNotExist(err) {
		os.Mkdir(envtabPath, 0700)
	}
	return envtabPath
}

func ListEnvtabEntries() []string {
	envtabPath := InitEnvtab()

	err := filepath.Walk(envtabPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fmt.Println(info.Name())
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error listing entries in the cache directory: %s\n", err)
		os.Exit(1)
	}
	return nil
}

func readEntry(name string) (*EnvTable, error) {

	filePath := filepath.Join(InitEnvtab(), name)

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

func writeEntryToFile(name, key, value string, tags []string) error {

	filePath := filepath.Join(InitEnvtab(), name)

	// Read the existing entries if file exists
	content, err := readEntry(name)
	if err != nil && !os.IsNotExist(err) {
		return err

		// Create a new file if it doesn't exist
	} else if os.IsNotExist(err) {
		content = &EnvTable{
			Metadata: EnvMetadata{
				CreatedAt: getCurrentTime(),
				LoadedAt:  getCurrentTime(),
				UpdatedAt: getCurrentTime(),
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
	content.Metadata.UpdatedAt = getCurrentTime()

	// Write the updated entries to the file
	data, err := yaml.Marshal(content)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0577)
	if err != nil {
		return err
	}

	return nil
}
