package envtab

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tagz "github.com/gmherb/envtab/pkg/tags"
	"github.com/gmherb/envtab/pkg/utils"
	yaml "gopkg.in/yaml.v2"
)

type LoadoutMetadata struct {
	CreatedAt   string   `json:"createdAt" yaml:"createdAt"`
	LoadedAt    string   `json:"loadedAt" yaml:"loadedAt"`
	UpdatedAt   string   `json:"updatedAt" yaml:"updatedAt"`
	Login       bool     `json:"login" yaml:"login"`
	Tags        []string `json:"tags" yaml:"tags"`
	Description string   `json:"description" yaml:"description"`
}

type Loadout struct {
	Metadata LoadoutMetadata   `json:"metadata" yaml:"metadata"`
	Entries  map[string]string `json:"entries" yaml:"entries"`
}

func (l Loadout) Export() {
	for key, value := range l.Entries {
		fmt.Printf("export %s=%s\n", key, value)
	}
	l.UpdateLoadedAt()
}

func (l Loadout) UpdateEntry(key, value string) error {
	println("DEBUG: UpdateEntry called")
	l.Entries[key] = value
	return nil
}

func (l Loadout) UpdateTags(tags []string) error {
	println("DEBUG: UpdateTags called")
	l.Metadata.Tags = tagz.MergeTags(l.Metadata.Tags, tags)
	return nil
}

func (l Loadout) ReplaceTags(tags []string) error {
	println("DEBUG: ReplaceTags called")
	l.Metadata.Tags = tags
	return nil
}

func (l Loadout) UpdateDescription(description string) error {
	println("DEBUG: UpdateDescription called")
	l.Metadata.Description = description
	return nil
}

func (l Loadout) UpdateLogin(login bool) error {
	println("DEBUG: UpdateLogin called")
	l.Metadata.Login = login
	return nil
}

func (l Loadout) UpdateUpdatedAt() error {
	println("DEBUG: UpdateUpdatedAt called")
	l.Metadata.UpdatedAt = utils.GetCurrentTime()
	return nil
}

func (l Loadout) UpdateLoadedAt() error {
	println("DEBUG: UpdateLoadedAt called")
	l.Metadata.LoadedAt = utils.GetCurrentTime()
	return nil
}

// Print a loadout file to stdout
func (l Loadout) PrintLoadout() error {

	data, err := yaml.Marshal(l)
	if err != nil {
		return err
	}

	fmt.Printf("%s", string(data))

	return nil
}

// Initialize a new Loadout struct
func InitLoadout() *Loadout {

	loadout := &Loadout{
		Metadata: LoadoutMetadata{
			CreatedAt:   utils.GetCurrentTime(),
			LoadedAt:    utils.GetCurrentTime(),
			UpdatedAt:   utils.GetCurrentTime(),
			Login:       false,
			Tags:        []string{},
			Description: "",
		},
		Entries: map[string]string{},
	}

	return loadout
}

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

func WriteLoadout(name string, loadout *Loadout) error {

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	loadout.UpdateUpdatedAt()

	data, err := yaml.Marshal(loadout)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0700)
	if err != nil {
		return err
	}

	return nil
}

// Write a key-value pair to a loadout (and optionally add tags)
func AddEntryToLoadout(name, key, value string, tags []string) error {

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

func DeleteLoadout(name string) error {

	filePath := filepath.Join(InitEnvtab(""), name+".yaml")

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}
