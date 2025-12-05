package backends

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gmherb/envtab/internal/config"
	"github.com/gmherb/envtab/internal/loadout"
)

func TestImportFromDotenv(t *testing.T) {
	envtabPath := config.InitEnvtab("")
	testLoadoutName := "test_import_dotenv"
	testDotenvFile := filepath.Join(envtabPath, "test_import.env")

	// Cleanup
	defer os.Remove(GetLoadoutFilePath(testLoadoutName))
	defer os.Remove(testDotenvFile)

	// Create a test .env file
	dotenvContent := `# Test .env file
KEY1=value1
KEY2=value2
KEY3=value with spaces
# Comment line
KEY4=value4
`
	err := os.WriteFile(testDotenvFile, []byte(dotenvContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Create a loadout with existing entries
	lo := loadout.InitLoadout()
	lo.Entries["EXISTING_KEY"] = "existing_value"
	lo.Entries["KEY1"] = "old_value" // This should be overwritten

	// Import from .env file
	err = ImportFromDotenv(lo, testDotenvFile)
	if err != nil {
		t.Fatalf("ImportFromDotenv() error = %v", err)
	}

	// Verify imported entries
	if lo.Entries["KEY1"] != "value1" {
		t.Errorf("ImportFromDotenv() failed to import KEY1, got %v, want value1", lo.Entries["KEY1"])
	}
	if lo.Entries["KEY2"] != "value2" {
		t.Errorf("ImportFromDotenv() failed to import KEY2, got %v, want value2", lo.Entries["KEY2"])
	}
	if lo.Entries["KEY3"] != "value with spaces" {
		t.Errorf("ImportFromDotenv() failed to import KEY3, got %v, want 'value with spaces'", lo.Entries["KEY3"])
	}
	if lo.Entries["KEY4"] != "value4" {
		t.Errorf("ImportFromDotenv() failed to import KEY4, got %v, want value4", lo.Entries["KEY4"])
	}

	// Verify existing entry is preserved
	if lo.Entries["EXISTING_KEY"] != "existing_value" {
		t.Errorf("ImportFromDotenv() overwrote existing key, got %v, want existing_value", lo.Entries["EXISTING_KEY"])
	}

	// Verify KEY1 was overwritten
	if lo.Entries["KEY1"] != "value1" {
		t.Errorf("ImportFromDotenv() failed to overwrite existing KEY1, got %v, want value1", lo.Entries["KEY1"])
	}

	// Test importing from non-existent file
	err = ImportFromDotenv(lo, "non_existent.env")
	if err == nil {
		t.Error("ImportFromDotenv() should return error for non-existent file")
	}
}

func TestParseDotenvContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name:     "simple key-value pairs",
			content:  "KEY1=value1\nKEY2=value2\nKEY3=value3",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2", "KEY3": "value3"},
		},
		{
			name:     "with comments",
			content:  "# This is a comment\nKEY1=value1\n# Another comment\nKEY2=value2",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "with empty lines",
			content:  "KEY1=value1\n\nKEY2=value2\n\nKEY3=value3",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2", "KEY3": "value3"},
		},
		{
			name:     "with whitespace",
			content:  "  KEY1  =  value1  \n  KEY2  =  value2  ",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "values with equals sign",
			content:  "KEY1=value=with=equals\nKEY2=normal_value",
			expected: map[string]string{"KEY1": "value=with=equals", "KEY2": "normal_value"},
		},
		{
			name:     "empty content",
			content:  "",
			expected: map[string]string{},
		},
		{
			name:     "only comments and empty lines",
			content:  "# Comment 1\n\n# Comment 2\n\n",
			expected: map[string]string{},
		},
		{
			name:     "invalid lines without equals",
			content:  "KEY1=value1\nINVALID_LINE\nKEY2=value2",
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "line with empty key",
			content:  "KEY1=value1\n=value2\nKEY3=value3",
			expected: map[string]string{"KEY1": "value1", "KEY3": "value3"},
		},
		{
			name:     "values with special characters",
			content:  "KEY1=value with spaces\nKEY2=value/with/slashes\nKEY3=value@with#special",
			expected: map[string]string{"KEY1": "value with spaces", "KEY2": "value/with/slashes", "KEY3": "value@with#special"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := ParseDotenvContent([]byte(tt.content))
			if err != nil {
				t.Fatalf("ParseDotenvContent() error = %v", err)
			}

			if len(entries) != len(tt.expected) {
				t.Errorf("ParseDotenvContent() returned %d entries, want %d", len(entries), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if entries[key] != expectedValue {
					t.Errorf("ParseDotenvContent() entries[%s] = %v, want %v", key, entries[key], expectedValue)
				}
			}
		})
	}
}
