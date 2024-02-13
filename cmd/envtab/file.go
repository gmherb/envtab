package envtab

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gmherb/envtab/pkg/utils"
	yaml "gopkg.in/yaml.v2"
)

// Write a key-value pair to a loadout (and optionally any tags)
func AddEntryToLoadout(name string, key string, value string, tags []string) error {

	// Read the existing entries if file exists
	loadout, err := ReadLoadout(name)
	if err != nil && !os.IsNotExist(err) {
		return err

	} else if os.IsNotExist(err) {
		loadout = InitLoadout()
	}

	loadout.UpdateEntry(key, value)
	loadout.UpdateTags(tags)

	return WriteLoadout(name, loadout)
}

// Remove a loadout file
func RemoveLoadout(name string) error {

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}

// Read a loadout from file and return a Loadout struct
func ReadLoadout(name string) (*Loadout, error) {

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var loadout Loadout
	err = yaml.Unmarshal(content, &loadout)
	if err != nil {
		return nil, err
	}

	return &loadout, nil
}

// Rename a loadout file
func RenameLoadout(oldName, newName string) error {

	envtabPath := InitEnvtab("")
	oldFilePath := filepath.Join(envtabPath, oldName+".yaml")
	newFilePath := filepath.Join(envtabPath, newName+".yaml")

	err := os.Rename(oldFilePath, newFilePath)
	if err != nil {
		return err
	}

	return nil
}

// Write a Loadout struct to file
func WriteLoadout(name string, loadout *Loadout) error {

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	data, err := yaml.Marshal(loadout)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

// Enter an interactive session to edit a loadout file
func EditLoadout(name string) error {

	var loadout Loadout
	var editedLoadout Loadout

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")
	tempFilePath := filePath + ".tmp"

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Load yaml file into a Loadout struct
	err = yaml.Unmarshal(data, &loadout)
	if err != nil {
		return err
	}

	// Save the original timestamps
	createdAt := loadout.Metadata.CreatedAt
	loadedAt := loadout.Metadata.LoadedAt

	// Write the Loadout struct to a temp file
	err = os.WriteFile(tempFilePath, data, 0600)
	if err != nil {
		return err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// Loop until valid answer is given or user aborts
	for {

		// Open the temp file in the editor
		cmd := exec.Command(editor, tempFilePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Run()

		// Read the temp file back into a Loadout struct
		data, err = os.ReadFile(tempFilePath)
		if err != nil {
			return err
		}

		// Load yaml file into a Loadout struct
		err = yaml.Unmarshal(data, &editedLoadout)

		// If the contents of the file could not be parsed
		// Ask the user to continue editing the file or abort
		if err != nil {

			usersChoice := utils.PromptForAnswer("The file could not be parsed. Do you want to continue editing to try to fix the errors? Enter 'yes' to continue to edit or 'no' to abort and discard changes?")
			if !usersChoice {
				return nil
			}
		}

		// If the contents of the file could be parsed
		// Break the loop
		if err == nil {
			break
		}
	}

	// Restore the original timestamps
	editedLoadout.Metadata.CreatedAt = createdAt
	editedLoadout.Metadata.LoadedAt = loadedAt

	// Only overwrite the loadout when modified
	if CompareLoadouts(loadout, editedLoadout) {
		editedLoadout.UpdateUpdatedAt()

		return WriteLoadout(name, &editedLoadout)
	}

	// Remove the temp file
	os.Remove(tempFilePath)
	return nil
}
