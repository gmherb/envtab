package backends

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gmherb/envtab/internal/config"
	"github.com/gmherb/envtab/internal/loadout"
	"github.com/gmherb/envtab/internal/sops"
	"github.com/gmherb/envtab/internal/utils"
	yaml "gopkg.in/yaml.v2"
)

// Write a key-value pair to a loadout (and optionally any tags)
func AddEntryToLoadout(name string, key string, value string, tags []string) error {

	// Read the existing entries if file exists
	lo, err := ReadLoadout(name)
	if err != nil && !os.IsNotExist(err) {
		return err

	} else if os.IsNotExist(err) {
		lo = loadout.InitLoadout()
	}

	lo.UpdateEntry(key, value)
	lo.UpdateTags(tags)

	// Check if file is SOPS-encrypted to preserve encryption
	filePath := filepath.Join(config.InitEnvtab(""), name+".yaml")
	isSOPSEncrypted := false
	if _, err := os.Stat(filePath); err == nil {
		isSOPSEncrypted = sops.IsSOPSEncrypted(filePath)
	}

	// Preserve SOPS encryption if the file was originally encrypted
	if isSOPSEncrypted {
		return WriteLoadoutWithEncryption(name, lo, true)
	}
	return WriteLoadout(name, lo)
}

// AddEntryToLoadoutWithSOPS writes a key-value pair to a loadout
// If useSOPS is true, encrypts the entire file with SOPS
func AddEntryToLoadoutWithSOPS(name string, key string, value string, tags []string, useSOPS bool) error {

	// Read the existing entries if file exists
	lo, err := ReadLoadout(name)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if os.IsNotExist(err) {
		lo = loadout.InitLoadout()
	}

	lo.UpdateEntry(key, value)
	lo.UpdateTags(tags)

	return WriteLoadoutWithEncryption(name, lo, useSOPS)
}

// Remove a loadout file
func RemoveLoadout(name string) error {

	filePath := filepath.Join(config.InitEnvtab(""), name+".yaml")

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}

// Read a loadout from file and return a Loadout struct
// Automatically handles SOPS-encrypted files
func ReadLoadout(name string) (*loadout.Loadout, error) {

	filePath := filepath.Join(config.InitEnvtab(""), name+".yaml")

	var content []byte
	var err error

	// Check if file is SOPS encrypted
	if sops.IsSOPSEncrypted(filePath) {
		content, err = sops.SOPSDecryptFile(filePath)
		if err != nil {
			// Provide helpful error messages
			errStr := err.Error()
			if strings.Contains(strings.ToLower(errStr), "sops command not found") {
				// Return a special error that can be handled gracefully
				return nil, fmt.Errorf("SOPS_NOT_INSTALLED: SOPS is not installed. Install SOPS to read encrypted loadouts: https://github.com/getsops/sops")
			}
			if strings.Contains(strings.ToLower(errStr), "keys may have been rotated") {
				return nil, fmt.Errorf("cannot decrypt loadout: encryption keys may have been rotated. Use 'envtab reencrypt %s' to re-encrypt with current keys: %w", name, err)
			}
			if strings.Contains(strings.ToLower(errStr), "not a valid sops file") ||
				strings.Contains(strings.ToLower(errStr), "no sops metadata found") {
				// False positive - file contains "sops:" but isn't actually encrypted
				// Fall back to reading as plain text
				content, err = os.ReadFile(filePath)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("failed to decrypt SOPS-encrypted loadout: %w", err)
			}
		}
	} else {
		content, err = os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
	}

	// Handle case where SOPS wrapped content in "data:" key (binary/blob encryption mode)
	// First try to parse and check if there's a "data" key at top level
	var dataWrapper map[string]interface{}
	if err := yaml.Unmarshal(content, &dataWrapper); err == nil {
		if dataValue, exists := dataWrapper["data"]; exists {
			// Content is wrapped in "data:" key, extract it
			if dataStr, ok := dataValue.(string); ok {
				// data: is a string (encrypted blob), use it as content
				content = []byte(dataStr)
			} else {
				// data: is an object, marshal it back to YAML/JSON
				var marshalErr error
				content, marshalErr = yaml.Marshal(dataValue)
				if marshalErr != nil {
					content, marshalErr = json.Marshal(dataValue)
					if marshalErr != nil {
						return nil, fmt.Errorf("failed to extract data from wrapper: %w", marshalErr)
					}
				}
			}
		}
	} else {
		// Try JSON format
		if err := json.Unmarshal(content, &dataWrapper); err == nil {
			if dataValue, exists := dataWrapper["data"]; exists {
				// Content is wrapped in "data:" key, extract it
				if dataStr, ok := dataValue.(string); ok {
					content = []byte(dataStr)
				} else {
					var marshalErr error
					content, marshalErr = json.Marshal(dataValue)
					if marshalErr != nil {
						return nil, fmt.Errorf("failed to extract data from JSON wrapper: %w", marshalErr)
					}
				}
			}
		}
	}

	var lo loadout.Loadout
	// Try YAML first (most common for envtab loadouts)
	err = yaml.Unmarshal(content, &lo)
	if err != nil {
		// If YAML parsing fails, try JSON (SOPS might decrypt to JSON format)
		err = json.Unmarshal(content, &lo)
		if err != nil {
			return nil, fmt.Errorf("failed to parse loadout (tried YAML and JSON): %w", err)
		}
	}

	return &lo, nil
}

// Rename a loadout file
func RenameLoadout(oldName, newName string) error {

	envtabPath := config.InitEnvtab("")
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
func WriteLoadout(name string, lo *loadout.Loadout) error {
	return WriteLoadoutWithEncryption(name, lo, false)
}

// WriteLoadoutWithEncryption writes a Loadout struct to file
// If useSOPS is true, encrypts the entire file with SOPS
func WriteLoadoutWithEncryption(name string, lo *loadout.Loadout, useSOPS bool) error {

	filePath := filepath.Join(config.InitEnvtab(""), name+".yaml")

	data, err := yaml.Marshal(lo)
	if err != nil {
		return err
	}

	if useSOPS {
		encrypted, err := sops.SOPSEncryptFile(filePath)
		if err != nil {
			return err
		}

		err = os.WriteFile(filePath, encrypted, 0600)
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
// Automatically handles SOPS-encrypted files and preserves encryption on save
func EditLoadout(name string) error {

	filePath := filepath.Join(config.InitEnvtab(""), name+".yaml")
	tempFilePath := filePath + ".tmp"

	isSOPSEncrypted := sops.IsSOPSEncrypted(filePath)

	lo, err := ReadLoadout(name)
	if err != nil {
		return err
	}

	encryptedKeys, err := lo.DecryptSOPSValues()
	if err != nil {
		slog.Warn("some SOPS values could not be decrypted", "error", err)
	}

	data, err := yaml.Marshal(lo)
	if err != nil {
		return err
	}

	// Save the original timestamps
	createdAt := lo.Metadata.CreatedAt
	loadedAt := lo.Metadata.LoadedAt

	err = os.WriteFile(tempFilePath, data, 0600)
	if err != nil {
		return err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	var editedLoadout *loadout.Loadout

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
		defer os.Remove(tempFilePath)

		// Validate YAML for duplicate keys before unmarshaling
		err = loadout.ValidateLoadoutYAML(data)
		if err != nil {
			slog.Error("invalid loadout YAML", "error", err)
			usersChoice := utils.PromptForAnswer("The file contains duplicate keys. Do you want to continue editing to fix the errors? Enter 'yes' to continue to edit or 'no' to abort and discard changes?")
			if !usersChoice {
				return nil
			}
			continue // Continue editing
		}

		// Load yaml file into a Loadout struct
		editedLoadout = &loadout.Loadout{}
		err = yaml.Unmarshal(data, editedLoadout)

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
	if loadout.CompareLoadouts(*lo, *editedLoadout) {
		editedLoadout.UpdateUpdatedAt()

		// Re-encrypt values that were originally SOPS-encrypted
		if len(encryptedKeys) > 0 {
			err := editedLoadout.ReencryptSOPSValues(encryptedKeys)
			if err != nil {
				return fmt.Errorf("failed to re-encrypt SOPS values: %w", err)
			}
		}

		// Preserve SOPS encryption if the file was originally encrypted
		if isSOPSEncrypted {
			return WriteLoadoutWithEncryption(name, editedLoadout, true)
		}
		return WriteLoadout(name, editedLoadout)
	}
	return nil
}

// ListLoadouts returns a list of all loadout names
// For file backend, this scans the envtab directory for YAML files
func ListLoadouts() ([]string, error) {
	envtabPath := config.InitEnvtab("")

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
		return nil, fmt.Errorf("error reading envtab directory %s: %w", envtabPath, err)
	}

	return loadouts, nil
}

// IsLoadoutFileEncrypted checks if a loadout file is encrypted at the file level
func IsLoadoutFileEncrypted(name string) bool {
	filePath := filepath.Join(config.InitEnvtab(""), name+".yaml")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return sops.IsSOPSEncrypted(filePath)
}

// HasValueEncryptedEntries checks if a loadout has any value-encrypted entries (SOPS: prefix)
func HasValueEncryptedEntries(lo *loadout.Loadout) bool {
	if lo == nil {
		return false
	}
	for _, value := range lo.Entries {
		if strings.HasPrefix(value, "SOPS:") {
			return true
		}
	}
	return false
}

// ParseDotenvContent parses .env file content and returns a map of key-value pairs
// It skips comments (lines starting with #) and empty lines
// Returns an error if the content cannot be parsed
func ParseDotenvContent(content []byte) (map[string]string, error) {
	entries := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		// Trim whitespace from the line
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first = only (values may contain =)
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Skip if key is empty
		if key == "" {
			continue
		}

		entries[key] = value
	}

	return entries, nil
}

// ImportFromDotenv reads a .env file and imports its entries into a loadout
func ImportFromDotenv(loadout *loadout.Loadout, dotenvFile string) error {
	dotenv, err := os.ReadFile(dotenvFile)
	if err != nil {
		return err
	}

	entries, err := ParseDotenvContent(dotenv)
	if err != nil {
		return err
	}

	for key, value := range entries {
		loadout.UpdateEntry(key, value)
	}

	return nil
}
