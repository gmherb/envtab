package config

import (
	"fmt"
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
	var err error

	usr, err := user.Current()
	if err != nil {
		t.Errorf("Error getting user's home directory: %s", err)
	}

	testPath := filepath.Join(usr.HomeDir, ".envtab_test")

	// Run function
	output := InitEnvtab(testPath)
	fmt.Println(output)

	// Test directory creation
	if _, err = os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", testPath)
	}

	// Test output
	if output != testPath {
		t.Errorf("Expected %s, got %s", testPath, output)
	}

	// Cleanup (rename envtab back to original name)
	err = os.Remove(output)
	if err != nil {
		t.Errorf("Error removing %s: %s", testPath, err)
	}
}

func TestGetEnvtabSlice(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Errorf("Error getting user's home directory: %s", err)
	}

	testPath := filepath.Join(usr.HomeDir, ".envtab_test2")

	envtabPath := InitEnvtab(testPath)

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
	output := GetEnvtabSlice(testPath)

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

	// Cleanup
	err = os.Remove(envtabPath)
	if err != nil {
		t.Errorf("Error removing %s: %s", envtabPath, err)
	}
}

