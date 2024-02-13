package envtab

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type LoadoutTemplate struct {
	Entries     []string `json:"entries" yaml:"entries"`
	Description string   `json:"description" yaml:"description"`
}

func MakeLoadoutFromTemplate(templateName string) Loadout {

	templatePath := filepath.Join(InitEnvtab(""), "templates/"+templateName+".yml")
	fmt.Printf("DEBUG: templatePath: %s\n", templatePath)

	loadout := InitLoadout()

	data, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("ERROR: Failure reading template [%s]: %s\n", templateName, err)
		os.Exit(1)
	}
	// load yaml file into LoadoutTemplate struct
	var template LoadoutTemplate
	err = yaml.Unmarshal(data, &template)
	if err != nil {
		fmt.Printf("ERROR: Failure parsing template [%s]: %s\n", templateName, err)
		os.Exit(1)
	}

	loadout.Metadata.Description = template.Description

	for _, entry := range template.Entries {
		loadout.Entries[entry] = ""
	}

	return *loadout

}
