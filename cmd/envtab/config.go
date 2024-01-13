package envtab

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const (
	ENVTAB_DIR = ".envtab"
)

type EnvMetadata struct {
	CreatedAt string   `json:"createdAt" yaml:"createdAt"`
	LoadedAt  string   `json:"loadedAt" yaml:"loadedAt"`
	UpdatedAt string   `json:"updatedAt" yaml:"updatedAt"`
	Login     bool     `json:"login" yaml:"login"`
	Tags      []string `json:"tags" yaml:"tags"`
}

type EnvTable struct {
	Metadata EnvMetadata       `json:"metadata" yaml:"metadata"`
	Entries  map[string]string `json:"entries" yaml:"entries"`
}

func (c *EnvTable) Save() error {
	return nil
}

func getEnvtabPath() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting user's home directory: %s\n", err)
		os.Exit(1)
	}

	return filepath.Join(usr.HomeDir, ENVTAB_DIR)
}

func InitEnvtab() string {
	envtabPath := getEnvtabPath()
	if _, err := os.Stat(envtabPath); os.IsNotExist(err) {
		os.Mkdir(envtabPath, 0700)
	}
	return envtabPath
}

func ListEnvtabEntries() []string {
	envtabPath := InitEnvtab()

	err := filepath.Walk(envtabPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fmt.Println(info.Name())
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error listing entries in the cache directory: %s\n", err)
		os.Exit(1)
	}
	return nil
}
