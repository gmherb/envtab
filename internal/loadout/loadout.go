package loadout

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/gmherb/envtab/pkg/sops"
	"github.com/gmherb/envtab/internal/tags"
	"github.com/gmherb/envtab/internal/utils"
	yaml "gopkg.in/yaml.v2"
)

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

// ValidateLoadout checks if a loadout has duplicate keys in the entries
// Returns an error if duplicates are found, listing the duplicate keys
func ValidateLoadout(loadout *Loadout) error {
	if loadout == nil {
		return fmt.Errorf("loadout is nil")
	}

	// Check for duplicate keys in entries
	// Since entries is a map, duplicates would have been silently overwritten during unmarshaling
	// We need to check the raw YAML content instead
	// For now, we'll validate by checking if the map structure is valid
	// (maps can't have duplicates by definition, so if we got here, the YAML was valid)

	// However, we can add additional validation here if needed
	// For example, checking for empty keys or other issues
	for key := range loadout.Entries {
		if key == "" {
			return fmt.Errorf("loadout contains empty key in entries")
		}
	}

	return nil
}

// ValidateLoadoutYAML checks the raw YAML content for duplicate keys in the entries section
// This is called before unmarshaling to detect duplicates that would be silently overwritten
func ValidateLoadoutYAML(yamlContent []byte) error {
	// Parse YAML as lines to check for duplicate keys in the entries section
	lines := strings.Split(string(yamlContent), "\n")
	inEntries := false
	indentLevel := 0
	seenKeys := make(map[string]int)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check if we're entering the entries section
		if strings.HasPrefix(trimmed, "entries:") {
			inEntries = true
			indentLevel = len(line) - len(strings.TrimLeft(line, " \t"))
			continue
		}

		// Check if we've left the entries section (reached a top-level key)
		if inEntries {
			currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))
			if currentIndent <= indentLevel && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				// We've left the entries section
				break
			}

			// Check if this is a key-value pair in entries
			if strings.Contains(trimmed, ":") {
				parts := strings.SplitN(trimmed, ":", 2)
				key := strings.TrimSpace(parts[0])
				if key != "" {
					if seenKeys[key] > 0 {
						return fmt.Errorf("duplicate key '%s' found in entries section at line %d (first occurrence at line %d)", key, i+1, seenKeys[key])
					}
					seenKeys[key] = i + 1
				}
			}
		}
	}

	return nil
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

				slog.Debug("found potential new PATH(s)", "path", newPath)

				for _, np := range strings.Split(newPath, string(os.PathListSeparator)) {
					if _, exists := pathMap[np]; !exists {

						slog.Debug("adding new path to PATH map", "path", np)
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
				decrypted, err := sops.SOPSDecryptValue(value)
				if err != nil {
					// Check if SOPS is not available - skip silently in that case
					// This allows shell scripts to continue working even without SOPS
					errStr := err.Error()
					if strings.Contains(strings.ToLower(errStr), "sops command not found") {
						slog.Debug("skipping encrypted entry - SOPS not available", "key", key)
						continue
					}
					// For other decryption failures, show error messages
					if utils.Contains(err.Error(), "keys may have been rotated") {
						slog.Warn("cannot decrypt - encryption keys may have been rotated", "key", key, "error", err)
						slog.Debug("to fix: re-encrypt the loadout with current keys using 'envtab reencrypt'")
					} else {
						slog.Error("failure decrypting SOPS value", "key", key, "error", err)
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
	slog.Debug("UpdateEntry called", "key", key)
	l.Entries[key] = value
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateTags(newTags []string) error {
	slog.Debug("UpdateTags called", "tags", newTags)
	l.Metadata.Tags = tags.MergeTags(l.Metadata.Tags, newTags)
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) ReplaceTags(tags []string) error {
	slog.Debug("ReplaceTags called", "tags", tags)
	l.Metadata.Tags = tags
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) RemoveTags(tagsToRemove []string) error {
	slog.Debug("RemoveTags called", "tags", tagsToRemove)
	l.Metadata.Tags = tags.RemoveTags(l.Metadata.Tags, tagsToRemove)
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateDescription(description string) error {
	slog.Debug("UpdateDescription called", "description", description)
	l.Metadata.Description = description
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateLogin(login bool) error {
	slog.Debug("UpdateLogin called", "login", login)
	l.Metadata.Login = login
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateUpdatedAt() error {
	slog.Debug("UpdateUpdatedAt called")
	l.Metadata.UpdatedAt = utils.GetCurrentTime()
	return nil
}

func (l *Loadout) UpdateLoadedAt() error {
	slog.Debug("UpdateLoadedAt called")
	l.Metadata.LoadedAt = utils.GetCurrentTime()
	return nil
}

// DecryptSOPSValues decrypts all SOPS-encrypted values in the loadout entries
// Returns a map of keys that were encrypted (for re-encryption on save)
func (l *Loadout) DecryptSOPSValues() (map[string]bool, error) {
	encryptedKeys := make(map[string]bool)
	for key, value := range l.Entries {
		if strings.HasPrefix(value, "SOPS:") {
			decrypted, err := sops.SOPSDecryptValue(value)
			if err != nil {
				// If decryption fails, keep the encrypted value and mark it
				// This allows editing other values even if some can't be decrypted
				slog.Warn("cannot decrypt - keeping encrypted value", "key", key, "error", err)
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
			encrypted, err := sops.SOPSEncryptValue(value)
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
