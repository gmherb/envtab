package envtab

import (
	"fmt"
	"os"
	"path/filepath"

	tagz "github.com/gmherb/envtab/pkg/tags"
	"github.com/gmherb/envtab/pkg/utils"
	yaml "gopkg.in/yaml.v2"
)

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
