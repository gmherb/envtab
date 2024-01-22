package envtab

import (
	"os"
	"os/exec"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

// Rename a loadout file
func RenameLoadout(oldName, newName string) error {

	envtabPath := InitEnvtab("")
	oldFilePath := filepath.Join(envtabPath, oldName+".yaml")
	newFilePath := filepath.Join(envtabPath, newName+".yaml")

	err := os.Rename(oldFilePath, newFilePath)
	if err != nil {
		return err
	}

	return nil
}

// Read a loadout from file and return a Loadout struct
func ReadLoadout(name string) (*Loadout, error) {

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var loadout Loadout
	err = yaml.Unmarshal(content, &loadout)
	if err != nil {
		return nil, err
	}

	return &loadout, nil
}

// Write a Loadout struct to file
func WriteLoadout(name string, loadout *Loadout) error {

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	data, err := yaml.Marshal(loadout)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

// Write a key-value pair to a loadout (and optionally any tags)
func AddEntryToLoadout(name string, key string, value string, tags []string) error {

	// Read the existing entries if file exists
	loadout, err := ReadLoadout(name)
	if err != nil && !os.IsNotExist(err) {
		return err

	} else if os.IsNotExist(err) {
		loadout = InitLoadout()
	}

	loadout.UpdateEntry(key, value)
	loadout.UpdateTags(tags)

	return WriteLoadout(name, loadout)
}

// Enter an interactive session to edit a loadout file
func EditLoadout(name string) error {

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Delete a loadout file
func DeleteLoadout(name string) error {

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}
