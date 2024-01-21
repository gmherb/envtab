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

	envtabExists := false

	// Prep (if envtab exists, rename it)
	if _, err = os.Stat(envtabPath); err == nil {
		err = os.Rename(envtabPath, envtabPath+".bak")
		envtabExists = true
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

	if envtabExists {
		err = os.Rename(envtabPath+".bak", envtabPath)
		if err != nil {
			t.Errorf("Error renaming %s to %s: %s", envtabPath, envtabPath+".bak", err)
		}
	}

}

func TestGetEnvtabSlice(t *testing.T) {
	envtabPath := InitEnvtab()

	// Prep (if envtab exists, rename it)
	if _, err := os.Stat(envtabPath); err == nil {
		err = os.Rename(envtabPath, envtabPath+".bak")
		if err != nil {
			t.Errorf("Error renaming %s to %s: %s", envtabPath, envtabPath+".bak", err)
		}
	}
	// Create envtab directory
	err := os.Mkdir(envtabPath, 0700)
	if err != nil {
		t.Errorf("Error creating %s: %s", envtabPath, err)
	}

	// Create test files
	testFiles := []string{
		"test1.yaml",
		"test2.yaml",
		"test3.yaml",
	}

	for _, testFile := range testFiles {
		file, err := os.Create(filepath.Join(envtabPath, testFile))
		if err != nil {
			t.Errorf("Error creating %s: %s", testFile, err)
		}
		file.Close()
	}

	// Run function
	output := GetEnvtabSlice()

	// Test
	expected := []string{
		"test1",
		"test2",
		"test3",
	}

	if len(output) != len(expected) {
		t.Errorf("Expected %d, got %d", len(expected), len(output))
	}

	for i, _ := range expected {
		if output[i] != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], output[i])
		}
	}

	// Cleanup (remove test files)
	for _, testFile := range testFiles {
		err := os.Remove(filepath.Join(envtabPath, testFile))
		if err != nil {
			t.Errorf("Error removing %s: %s", testFile, err)
		}
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
