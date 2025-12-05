package backends

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gmherb/envtab/internal/config"
	"github.com/gmherb/envtab/internal/loadout"
	yaml "gopkg.in/yaml.v2"
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

func TestAddEntryToLoadout(t *testing.T) {
	testLoadoutName := "test_add_entry"

	defer os.Remove(GetLoadoutFilePath(testLoadoutName))

	err := AddEntryToLoadout(testLoadoutName, "TEST_KEY", "test_value", []string{"tag1"})
	if err != nil {
		t.Fatalf("AddEntryToLoadout() error = %v", err)
	}

	// Read back and verify
	lo, err := ReadLoadout(testLoadoutName)
	if err != nil {
		t.Fatalf("ReadLoadout() error = %v", err)
	}

	if lo.Entries["TEST_KEY"] != "test_value" {
		t.Errorf("AddEntryToLoadout() failed to add entry, got %v, want test_value", lo.Entries["TEST_KEY"])
	}

	// Test adding to existing loadout
	err = AddEntryToLoadout(testLoadoutName, "TEST_KEY2", "test_value2", []string{"tag2"})
	if err != nil {
		t.Fatalf("AddEntryToLoadout() error on second add = %v", err)
	}

	lo, err = ReadLoadout(testLoadoutName)
	if err != nil {
		t.Fatalf("ReadLoadout() error = %v", err)
	}

	if lo.Entries["TEST_KEY2"] != "test_value2" {
		t.Errorf("AddEntryToLoadout() failed to add second entry")
	}
}

func TestReadLoadout(t *testing.T) {
	testLoadoutName := "test_read_loadout"

	defer os.Remove(GetLoadoutFilePath(testLoadoutName))

	lo := loadout.InitLoadout()
	lo.Entries["TEST_KEY"] = "test_value"
	lo.Metadata.Description = "test description"
	lo.Metadata.Tags = []string{"tag1", "tag2"}

	err := WriteLoadout(testLoadoutName, lo)
	if err != nil {
		t.Fatalf("WriteLoadout() error = %v", err)
	}

	// Read it back
	readLo, err := ReadLoadout(testLoadoutName)
	if err != nil {
		t.Fatalf("ReadLoadout() error = %v", err)
	}

	if readLo.Entries["TEST_KEY"] != "test_value" {
		t.Errorf("ReadLoadout() failed to read entry, got %v, want test_value", readLo.Entries["TEST_KEY"])
	}

	if readLo.Metadata.Description != "test description" {
		t.Errorf("ReadLoadout() failed to read description")
	}

	// Test reading non-existent loadout
	_, err = ReadLoadout("non_existent_loadout")
	if err == nil {
		t.Error("ReadLoadout() should return error for non-existent loadout")
	}
}

func TestWriteLoadout(t *testing.T) {
	testLoadoutName := "test_write_loadout"
	filePath := GetLoadoutFilePath(testLoadoutName)

	defer os.Remove(filePath)

	lo := loadout.InitLoadout()
	lo.Entries["TEST_KEY"] = "test_value"

	err := WriteLoadout(testLoadoutName, lo)
	if err != nil {
		t.Fatalf("WriteLoadout() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("WriteLoadout() failed to create file")
	}

	// Read back and verify
	readLo, err := ReadLoadout(testLoadoutName)
	if err != nil {
		t.Fatalf("ReadLoadout() error = %v", err)
	}

	if readLo.Entries["TEST_KEY"] != "test_value" {
		t.Errorf("WriteLoadout() failed to write entry")
	}
}

func TestRemoveLoadout(t *testing.T) {
	testLoadoutName := "test_remove_loadout"
	filePath := GetLoadoutFilePath(testLoadoutName)

	lo := loadout.InitLoadout()
	err := WriteLoadout(testLoadoutName, lo)
	if err != nil {
		t.Fatalf("WriteLoadout() error = %v", err)
	}

	// Verify it exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Test file should exist before removal")
	}

	// Remove it
	err = RemoveLoadout(testLoadoutName)
	if err != nil {
		t.Fatalf("RemoveLoadout() error = %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("RemoveLoadout() failed to remove file")
	}
}

func TestRenameLoadout(t *testing.T) {
	oldName := "test_rename_old"
	newName := "test_rename_new"

	// Cleanup
	defer os.Remove(GetLoadoutFilePath(oldName))
	defer os.Remove(GetLoadoutFilePath(newName))

	// Create a test loadout
	lo := loadout.InitLoadout()
	lo.Entries["TEST_KEY"] = "test_value"
	err := WriteLoadout(oldName, lo)
	if err != nil {
		t.Fatalf("WriteLoadout() error = %v", err)
	}

	// Rename it
	err = RenameLoadout(oldName, newName)
	if err != nil {
		t.Fatalf("RenameLoadout() error = %v", err)
	}

	// Verify old file doesn't exist
	oldPath := GetLoadoutFilePath(oldName)
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("RenameLoadout() failed to remove old file")
	}

	// Verify new file exists and has correct content
	readLo, err := ReadLoadout(newName)
	if err != nil {
		t.Fatalf("ReadLoadout() error after rename = %v", err)
	}

	if readLo.Entries["TEST_KEY"] != "test_value" {
		t.Errorf("RenameLoadout() failed to preserve content")
	}
}

func TestWriteLoadoutWithEncryption(t *testing.T) {
	testLoadoutName := "test_write_encrypted"

	defer os.Remove(GetLoadoutFilePath(testLoadoutName))

	lo := loadout.InitLoadout()
	lo.Entries["TEST_KEY"] = "test_value"

	// Write without encryption
	err := WriteLoadoutWithEncryption(testLoadoutName, lo, false)
	if err != nil {
		t.Fatalf("WriteLoadoutWithEncryption() error = %v", err)
	}

	// Verify file exists and is not encrypted (can be read directly)
	filePath := GetLoadoutFilePath(testLoadoutName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Should be readable YAML
	var testLo loadout.Loadout
	err = yaml.Unmarshal(content, &testLo)
	if err != nil {
		t.Errorf("WriteLoadoutWithEncryption() with useSOPS=false should write readable YAML: %v", err)
	}
}

func TestAddEntryToLoadoutWithSOPS(t *testing.T) {
	testLoadoutName := "test_add_entry_sops"

	defer os.Remove(GetLoadoutFilePath(testLoadoutName))

	err := AddEntryToLoadoutWithSOPS(testLoadoutName, "TEST_KEY", "test_value", []string{"tag1"}, false)
	if err != nil {
		t.Fatalf("AddEntryToLoadoutWithSOPS() error = %v", err)
	}

	// Verify entry was added
	lo, err := ReadLoadout(testLoadoutName)
	if err != nil {
		t.Fatalf("ReadLoadout() error = %v", err)
	}

	if lo.Entries["TEST_KEY"] != "test_value" {
		t.Errorf("AddEntryToLoadoutWithSOPS() failed to add entry")
	}
}

func TestIsLoadoutFileEncrypted(t *testing.T) {
	testLoadoutName := "test_is_encrypted"

	defer os.Remove(GetLoadoutFilePath(testLoadoutName))

	// Test with non-existent file
	isEncrypted := IsLoadoutFileEncrypted("non_existent_loadout")
	if isEncrypted {
		t.Error("IsLoadoutFileEncrypted() should return false for non-existent file")
	}

	// Test with unencrypted file
	lo := loadout.InitLoadout()
	lo.Entries["TEST_KEY"] = "test_value"
	err := WriteLoadout(testLoadoutName, lo)
	if err != nil {
		t.Fatalf("WriteLoadout() error = %v", err)
	}

	isEncrypted = IsLoadoutFileEncrypted(testLoadoutName)
	if isEncrypted {
		t.Error("IsLoadoutFileEncrypted() should return false for unencrypted file")
	}
}

func TestHasValueEncryptedEntries(t *testing.T) {
	tests := []struct {
		name     string
		loadout  *loadout.Loadout
		expected bool
	}{
		{
			name:     "nil loadout",
			loadout:  nil,
			expected: false,
		},
		{
			name: "no encrypted entries",
			loadout: &loadout.Loadout{
				Entries: map[string]string{
					"KEY1": "value1",
					"KEY2": "value2",
				},
			},
			expected: false,
		},
		{
			name: "with SOPS: prefix",
			loadout: &loadout.Loadout{
				Entries: map[string]string{
					"KEY1": "value1",
					"KEY2": "SOPS:encrypted_value",
				},
			},
			expected: true,
		},
		{
			name: "multiple encrypted entries",
			loadout: &loadout.Loadout{
				Entries: map[string]string{
					"KEY1": "SOPS:encrypted_value1",
					"KEY2": "SOPS:encrypted_value2",
				},
			},
			expected: true,
		},
		{
			name: "value contains SOPS: but not as prefix",
			loadout: &loadout.Loadout{
				Entries: map[string]string{
					"KEY1": "value with SOPS: in the middle",
				},
			},
			expected: false,
		},
		{
			name: "empty loadout",
			loadout: &loadout.Loadout{
				Entries: map[string]string{},
			},
			expected: false,
		},
		{
			name: "lowercase sops: prefix",
			loadout: &loadout.Loadout{
				Entries: map[string]string{
					"KEY1": "sops:encrypted_value",
				},
			},
			expected: false, // Should only match uppercase "SOPS:"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasValueEncryptedEntries(tt.loadout)
			if result != tt.expected {
				t.Errorf("HasValueEncryptedEntries() = %v, want %v", result, tt.expected)
			}
		})
	}
}
