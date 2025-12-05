package backends

import (
	"os"
	"strings"

	"github.com/gmherb/envtab/internal/loadout"
)

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
