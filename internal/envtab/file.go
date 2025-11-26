package envtab

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gmherb/envtab/internal/crypto"
	"github.com/gmherb/envtab/internal/utils"
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

// AddEntryToLoadoutWithSOPS writes a key-value pair to a loadout
// If useSOPS is true, encrypts the entire file with SOPS
func AddEntryToLoadoutWithSOPS(name string, key string, value string, tags []string, useSOPS bool) error {

	// Read the existing entries if file exists
	loadout, err := ReadLoadout(name)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if os.IsNotExist(err) {
		loadout = InitLoadout()
	}

	loadout.UpdateEntry(key, value)
	loadout.UpdateTags(tags)

	return WriteLoadoutWithEncryption(name, loadout, useSOPS)
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
// Automatically handles SOPS-encrypted files
func ReadLoadout(name string) (*Loadout, error) {

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	var content []byte
	var err error

	// Check if file is SOPS encrypted
	if crypto.IsSOPSEncrypted(filePath) {
		content, err = crypto.SOPSDecryptFile(filePath)
		if err != nil {
			// Provide helpful error for key rotation
			if strings.Contains(strings.ToLower(err.Error()), "keys may have been rotated") {
				return nil, fmt.Errorf("cannot decrypt loadout: encryption keys may have been rotated. Use 'envtab reencrypt %s' to re-encrypt with current keys: %w", name, err)
			}
			return nil, fmt.Errorf("failed to decrypt SOPS-encrypted loadout: %w", err)
		}
	} else {
		content, err = os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
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
// If useSOPS is true, encrypts the entire file with SOPS
func WriteLoadout(name string, loadout *Loadout) error {
	return WriteLoadoutWithEncryption(name, loadout, false)
}

// WriteLoadoutWithEncryption writes a Loadout struct to file
// If useSOPS is true, encrypts the entire file with SOPS
func WriteLoadoutWithEncryption(name string, loadout *Loadout, useSOPS bool) error {

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	data, err := yaml.Marshal(loadout)
	if err != nil {
		return err
	}

	if useSOPS {
		// Write to temp file first, then encrypt
		tmpFile := filePath + ".tmp"
		err = os.WriteFile(tmpFile, data, 0600)
		if err != nil {
			return err
		}

		encrypted, err := crypto.SOPSEncryptFile(tmpFile)
		if err != nil {
			os.Remove(tmpFile)
			return err
		}

		err = os.WriteFile(filePath, encrypted, 0600)
		os.Remove(tmpFile)
		if err != nil {
			return err
		}
	} else {
		err = os.WriteFile(filePath, data, 0600)
		if err != nil {
			return err
		}
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
