package envtab

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestGetEnvtabPath(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Errorf("Error getting user's home directory: %s", err)
	}

	expected := filepath.Join(usr.HomeDir, ENVTAB_DIR)
	actual := getEnvtabPath()

	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestInitEnvtab(t *testing.T) {
	envtabPath := getEnvtabPath()

	var err error

	// Prep (if envtab exists, rename it)
	if _, err = os.Stat(envtabPath); err == nil {
		err = os.Rename(envtabPath, envtabPath+".bak")
		if err != nil {
			t.Errorf("Error renaming %s to %s: %s", envtabPath, envtabPath+".bak", err)
		}
	}

	// Run function
	output := InitEnvtab()

	// Test
	if _, err = os.Stat(envtabPath); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", envtabPath)
	}

	if output != envtabPath {
		t.Errorf("Expected %s, got %s", envtabPath, output)
	}

	// Cleanup (rename envtab back to original name)
	err = os.Remove(envtabPath)
	if err != nil {
		t.Errorf("Error removing %s: %s", envtabPath, err)
	}

	err = os.Rename(envtabPath+".bak", envtabPath)
	if err != nil {
		t.Errorf("Error renaming %s to %s: %s", envtabPath, envtabPath+".bak", err)
	}

}

func TestPrintEnvtabLoadouts(t *testing.T) {
	PrintEnvtabLoadouts()
}

func TestReadLoadout(t *testing.T) {
	name := "TestReadLoadout"
	filePath := filepath.Join(InitEnvtab(), name+".yaml")

	// Create test file
	f, err := os.Create(filePath)
	if err != nil {
		t.Errorf("Error creating test file %s: %s", filePath, err)
	}
	defer f.Close()

	// Write test data to file
	_, err = f.WriteString("metadata:\n  createdAt: 2021-07-04T15:04:05Z\n  loadedAt: 2021-07-04T15:04:05Z\n  updatedAt: 2021-07-04T15:04:05Z\n  login: false\n  tags: []\nentries:\n  test: test\n")
	if err != nil {
		t.Errorf("Error writing test data to %s: %s", filePath, err)
	}

	// Read test file
	entry, err := ReadLoadout(name)
	if err != nil {
		t.Errorf("Error reading test file %s: %s", filePath, err)
	}

	// Test
	if entry.Metadata.CreatedAt != "2021-07-04T15:04:05Z" {
		t.Errorf("Expected 2021-07-04T15:04:05Z, got %s", entry.Metadata.CreatedAt)
	}

	// Cleanup
	err = os.Remove(filePath)
	if err != nil {
		t.Errorf("Error removing %s: %s", filePath, err)
	}
}

func TestWriteEntryToLoadout(t *testing.T) {
	name := "TestWriteEntryToLoadout"
	filePath := filepath.Join(InitEnvtab(), name+".yaml")

	err := WriteEntryToLoadout(name, "test2", "test2", []string{"test"})
	if err != nil {
		t.Errorf("Error writing test data to %s: %s", filePath, err)
	}

	entry, err := ReadLoadout(name)
	if err != nil {
		t.Errorf("Error reading test file %s: %s", filePath, err)
	}

	if entry.Entries["test2"] != "test2" {
		t.Errorf("Expected test2, got %s", entry.Entries["test2"])
	}

	err = os.Remove(filePath)
	if err != nil {
		t.Errorf("Error removing %s: %s", filePath, err)
	}
}
