package envtab

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gmherb/envtab/internal/crypto"
	"github.com/gmherb/envtab/internal/tags"
	"github.com/gmherb/envtab/internal/utils"
	yaml "gopkg.in/yaml.v2"
)

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

type LoadoutMetadata struct {
	CreatedAt   string   `json:"createdAt" yaml:"createdAt"`
	LoadedAt    string   `json:"loadedAt" yaml:"loadedAt"`
	UpdatedAt   string   `json:"updatedAt" yaml:"updatedAt"`
	Login       bool     `json:"login" yaml:"login"`
	Tags        []string `json:"tags" yaml:"tags"`
	Description string   `json:"description" yaml:"description"`
}

type Loadout struct {
	Metadata LoadoutMetadata   `json:"metadata" yaml:"metadata"`
	Entries  map[string]string `json:"entries" yaml:"entries"`
}

func (l Loadout) Export() {

	pathMap := make(map[string]bool)
	order := []string{}

	for _, p := range strings.Split(os.Getenv("PATH"), string(os.PathListSeparator)) {
		if _, exists := pathMap[p]; !exists {
			order = append(order, p)
		}
		pathMap[p] = true
	}

	re := regexp.MustCompile(`\$PATH`)
	reSOPS := regexp.MustCompile(`^SOPS:`)

		for key, value := range l.Entries {
		if value != "" {
			match := re.MatchString(value)
			sopsEncrypted := reSOPS.MatchString(value)
			if key == "PATH" && match {

				newPath := re.ReplaceAllString(value, "")
				newPath = strings.Trim(newPath, ":")
				for strings.Contains(newPath, "::") {
					newPath = strings.ReplaceAll(newPath, "::", ":")
				}

				println("DEBUG: Found potential new PATH(s) [" + newPath + "].")

				for _, np := range strings.Split(newPath, string(os.PathListSeparator)) {
					if _, exists := pathMap[np]; !exists {

						println("DEBUG: Adding new path [" + np + "] to PATH map.")
						order = append(order, np)
						pathMap[np] = true
					}
				}

				paths := make([]string, 0, len(pathMap)) // preallocate for efficiency
				paths = append(paths, order...)

				os.Setenv("PATH", strings.Join(paths, string(os.PathListSeparator)))
				fmt.Printf("export PATH=%s\n", os.Getenv("PATH"))
			} else if sopsEncrypted {
				// Decrypt SOPS-encrypted value
				// SOPS metadata is preserved in the encrypted value string
				decrypted, err := crypto.SOPSDecryptValue(value)
				if err != nil {
					// Check if it's a key rotation issue
					if contains(err.Error(), "keys may have been rotated") {
						fmt.Fprintf(os.Stderr, "WARNING: Cannot decrypt %s - encryption keys may have been rotated. Skipping.\n", key)
						fmt.Fprintf(os.Stderr, "         To fix: re-encrypt the loadout with current keys using 'envtab reencrypt'\n")
					} else {
						fmt.Fprintf(os.Stderr, "ERROR: Failed to decrypt SOPS value for %s: %s\n", key, err)
					}
					continue
				}
				fmt.Printf("export %s=%s\n", key, decrypted)
			} else {
				fmt.Printf("export %s=%s\n", key, value)
			}
		}
	}
	l.UpdateLoadedAt()
}

func (l *Loadout) UpdateEntry(key string, value string) error {
	println("DEBUG: UpdateEntry called")
	l.Entries[key] = value
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateTags(newTags []string) error {
	println("DEBUG: UpdateTags called")
	l.Metadata.Tags = tags.MergeTags(l.Metadata.Tags, newTags)
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) ReplaceTags(tags []string) error {
	println("DEBUG: ReplaceTags called")
	l.Metadata.Tags = tags
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateDescription(description string) error {
	println("DEBUG: UpdateDescription called")
	l.Metadata.Description = description
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateLogin(login bool) error {
	println("DEBUG: UpdateLogin called")
	l.Metadata.Login = login
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateUpdatedAt() error {
	println("DEBUG: UpdateUpdatedAt called")
	l.Metadata.UpdatedAt = utils.GetCurrentTime()
	return nil
}

func (l *Loadout) UpdateLoadedAt() error {
	println("DEBUG: UpdateLoadedAt called")
	l.Metadata.LoadedAt = utils.GetCurrentTime()
	return nil
}

// DecryptSOPSValues decrypts all SOPS-encrypted values in the loadout entries
// Returns a map of keys that were encrypted (for re-encryption on save)
func (l *Loadout) DecryptSOPSValues() (map[string]bool, error) {
	encryptedKeys := make(map[string]bool)
	for key, value := range l.Entries {
		if strings.HasPrefix(value, "SOPS:") {
			decrypted, err := crypto.SOPSDecryptValue(value)
			if err != nil {
				// If decryption fails, keep the encrypted value and mark it
				// This allows editing other values even if some can't be decrypted
				fmt.Fprintf(os.Stderr, "WARNING: Cannot decrypt %s - keeping encrypted value: %s\n", key, err)
				encryptedKeys[key] = true
				continue
			}
			l.Entries[key] = decrypted
			encryptedKeys[key] = true
		}
	}
	return encryptedKeys, nil
}

// ReencryptSOPSValues re-encrypts values for keys that were originally encrypted
func (l *Loadout) ReencryptSOPSValues(encryptedKeys map[string]bool) error {
	for key := range encryptedKeys {
		value, exists := l.Entries[key]
		if !exists {
			continue
		}
		// Only re-encrypt if the value doesn't already start with SOPS:
		// (user might have manually edited it to be encrypted)
		if !strings.HasPrefix(value, "SOPS:") {
			encrypted, err := crypto.SOPSEncryptValue(value)
			if err != nil {
				return fmt.Errorf("failed to re-encrypt %s: %w", key, err)
			}
			l.Entries[key] = encrypted
		}
	}
	return nil
}

func (l *Loadout) PrintLoadout() error {

	data, err := yaml.Marshal(l)
	if err != nil {
		return err
	}

	fmt.Printf("%s", string(data))

	return nil
}

// Initialize a new Loadout struct
func InitLoadout() *Loadout {

	loadout := &Loadout{
		Metadata: LoadoutMetadata{
			CreatedAt:   utils.GetCurrentTime(),
			LoadedAt:    utils.GetCurrentTime(),
			UpdatedAt:   utils.GetCurrentTime(),
			Login:       false,
			Tags:        []string{},
			Description: "",
		},
		Entries: map[string]string{},
	}

	return loadout
}

func CompareLoadouts(old Loadout, new Loadout) bool {
	if old.Metadata.CreatedAt != new.Metadata.CreatedAt {
		return true
	}
	if old.Metadata.LoadedAt != new.Metadata.LoadedAt {
		return true
	}
	if old.Metadata.UpdatedAt != new.Metadata.UpdatedAt {
		return true
	}
	if old.Metadata.Login != new.Metadata.Login {
		return true
	}
	if len(old.Metadata.Tags) != len(new.Metadata.Tags) {
		return true
	}
	for i, tag := range old.Metadata.Tags {
		if tag != new.Metadata.Tags[i] {
			return true
		}
	}
	if old.Metadata.Description != new.Metadata.Description {
		return true
	}
	if len(old.Entries) != len(new.Entries) {
		return true
	}
	for key, value := range old.Entries {
		if value != new.Entries[key] {
			return true
		}
	}
	return false
}
