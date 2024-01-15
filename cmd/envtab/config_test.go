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
