package envtab

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tagz "github.com/gmherb/envtab/pkg/tags"
	"github.com/gmherb/envtab/pkg/utils"
	yaml "gopkg.in/yaml.v2"
)

// Print all envtab loadouts
func PrintEnvtabLoadouts() {
	loadouts := getEnvtabSlice()
	for _, loadouts := range loadouts {
		fmt.Println(loadouts)
	}
}

// Print all envtab entries in a loadout and its metadata
func ReadLoadout(name string) (*EnvTable, error) {

	filePath := filepath.Join(InitEnvtab(), name+".yaml")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var loadout EnvTable
	err = yaml.Unmarshal(content, &loadout)
	if err != nil {
		return nil, err
	}

	return &loadout, nil

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

func EditLoadout(name string) error {

	filePath := filepath.Join(InitEnvtab(), name+".yaml")

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func DeleteLoadout(name string) error {

	filePath := filepath.Join(InitEnvtab(), name+".yaml")

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}

func ListEnvtabLoadouts() {
	envtabSlice := getEnvtabSlice()

	fmt.Println("UpdatedAt    LoadedAt    Login   Name                 Tags")
	for _, loadout := range envtabSlice {

		lo, err := ReadLoadout(loadout)
		if err != nil {
			fmt.Printf("Error reading loadout %s: %s\n", loadout, err)
			os.Exit(1)
		}

		updatedAt, err := time.Parse(time.RFC3339, lo.Metadata.UpdatedAt)
		if err != nil {
			fmt.Printf("Error parsing time %s: %s\n", lo.Metadata.UpdatedAt, err)
			os.Exit(1)
		}

		loadedAt, err := time.Parse(time.RFC3339, lo.Metadata.LoadedAt)
		if err != nil {
			fmt.Printf("Error parsing time %s: %s\n", lo.Metadata.UpdatedAt, err)
			os.Exit(1)
		}

		// TODO: Determine if time is under 24 hours and print time only
		// instead of date

		// Support column length of 20 characters for loadout name
		paddedLoadout := utils.PadString(loadout, 20)

		fmt.Println(updatedAt.Format(time.DateOnly), " ", loadedAt.Format(time.DateOnly), "", lo.Metadata.Login, " ", paddedLoadout, lo.Metadata.Tags)

	}
}
