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

func TestCreateEnvtab(t *testing.T) {
	envtabPath := getEnvtabPath()

	if _, err := os.Stat(envtabPath); os.IsNotExist(err) {
		os.Rename(envtabPath, envtabPath+".bak")
		defer os.Rename(envtabPath+".bak", envtabPath)
	}

	createEnvtabDir()

	if _, err := os.Stat(envtabPath); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", envtabPath)
	}
}

func TestInitEnvtab(t *testing.T) {
	envtabPath := getEnvtabPath()

	if _, err := os.Stat(envtabPath); os.IsNotExist(err) {
		os.Rename(envtabPath, envtabPath+".bak")
		defer os.Rename(envtabPath+".bak", envtabPath)
	}

	output := InitEnvtab()

	if _, err := os.Stat(envtabPath); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", envtabPath)
	}

	if output != envtabPath {
		t.Errorf("Expected %s, got %s", envtabPath, output)
	}
}

func TestListEnvtabEntries(t *testing.T) {
	listEnvtabEntries()
}
