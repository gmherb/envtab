package envtab

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadLoadout(t *testing.T) {
	name := "TestReadLoadout"
	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

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

func TestAddEntryToLoadout(t *testing.T) {
	name := "TestAddEntryToLoadout"
	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	err := AddEntryToLoadout(name, "test2", "test2", []string{"test"})
	if err != nil {
		t.Errorf("Error writing test data to %s: %s", filePath, err)
	}

	loadout, err := ReadLoadout(name)
	if err != nil {
		t.Errorf("Error reading test file %s: %s", filePath, err)
	}
	loadout.PrintLoadout()

	if loadout.Entries["test2"] != "test2" {
		t.Errorf("Expected test2, got %s", loadout.Entries["test2"])
	}

	println(loadout.Metadata.Tags[0])
	if loadout.Metadata.Tags[0] != "test" {
		t.Errorf("Expected test, got %s", loadout.Metadata.Tags[0])
	}

	err = os.Remove(filePath)
	if err != nil {
		t.Errorf("Error removing %s: %s", filePath, err)
	}
}

func TestLoadoutExport(t *testing.T) {
	name := "TestLoadoutExport"
	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	err := AddEntryToLoadout(name, "test2", "test2", []string{"test"})
	if err != nil {
		t.Errorf("Error writing test data to %s: %s", filePath, err)
	}

	loadout, err := ReadLoadout(name)
	if err != nil {
		t.Errorf("Error reading test file %s: %s", filePath, err)
	}

	loadout.Export()

	err = os.Remove(filePath)
	if err != nil {
		t.Errorf("Error removing %s: %s", filePath, err)
	}
}

func TestRenameLoadout(t *testing.T) {
	name := "TestRenameLoadout"
	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	// Create test loadout
	err := AddEntryToLoadout(name, "test2", "test2", []string{"test"})
	if err != nil {
		t.Errorf("Error writing test data to %s: %s", filePath, err)
	}

	// Run RenameLoadout
	err = RenameLoadout(name, "TestRenameLoadout2")
	if err != nil {
		t.Errorf("Error renaming loadout %s: %s", name, err)
	}

	// Test (which also cleans up)
	filePath = filepath.Join(InitEnvtab(""), "TestRenameLoadout2.yaml")
	err = os.Remove(filePath)
	if err != nil {
		t.Errorf("Error removing %s: %s", filePath, err)
	}

}
