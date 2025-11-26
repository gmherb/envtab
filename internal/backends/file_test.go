package backends

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gmherb/envtab/internal/config"
)

func TestListLoadouts(t *testing.T) {
	envtabPath := config.InitEnvtab("")

	// Create test files with unique names to avoid conflicts
	testFiles := []string{
		"test_list_loadouts_1.yaml",
		"test_list_loadouts_2.yaml",
		"test_list_loadouts_3.yaml",
	}

	for _, testFile := range testFiles {
		file, err := os.Create(filepath.Join(envtabPath, testFile))
		if err != nil {
			t.Errorf("Error creating %s: %s", testFile, err)
		}
		file.Close()
	}

	// Run function
	output, err := ListLoadouts()
	if err != nil {
		t.Errorf("Error listing loadouts: %s", err)
	}

	// Test that our test files are in the list
	expected := []string{
		"test_list_loadouts_1",
		"test_list_loadouts_2",
		"test_list_loadouts_3",
	}

	// Create a map for quick lookup
	outputMap := make(map[string]bool)
	for _, name := range output {
		outputMap[name] = true
	}

	// Verify all test files are in the output
	for _, expectedName := range expected {
		if !outputMap[expectedName] {
			t.Errorf("Expected loadout %s not found in output", expectedName)
		}
	}

	// Cleanup (remove test files)
	for _, testFile := range testFiles {
		err := os.Remove(filepath.Join(envtabPath, testFile))
		if err != nil {
			t.Errorf("Error removing %s: %s", testFile, err)
		}
	}
}
