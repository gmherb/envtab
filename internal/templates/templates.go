package templates

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gmherb/envtab/internal/config"
	"github.com/gmherb/envtab/internal/loadout"
	yaml "gopkg.in/yaml.v2"
)

type LoadoutTemplate struct {
	Entries     []string `json:"entries" yaml:"entries"`
	Description string   `json:"description" yaml:"description"`
}

type LoadoutTemplates struct {
	Templates map[string]LoadoutTemplate `json:"templates" yaml:"templates"`
}

func MakeLoadoutFromTemplate(templateName string, force bool) loadout.Loadout {
	lo := loadout.InitLoadout()
	var template LoadoutTemplate
	var found bool

	// First, check embedded templates
	if embeddedTemplate, exists := envtabTemplates.Templates[templateName]; exists {
		template = embeddedTemplate
		found = true
		slog.Debug("using embedded template", "template", templateName)
	} else {
		// Fall back to file-based templates
		templatePath := filepath.Join(config.InitEnvtab(""), "templates/"+templateName+".yml")
		if _, err := os.Stat(templatePath); err == nil {
			slog.Debug("using file template", "template", templateName, "path", templatePath)

			data, err := os.ReadFile(templatePath)
			if err != nil {
				fmt.Printf("ERROR: Failure reading template [%s]: %s\n", templateName, err)
				os.Exit(1)
			}
			// load yaml file into LoadoutTemplate struct
			err = yaml.Unmarshal(data, &template)
			if err != nil {
				fmt.Printf("ERROR: Failure parsing template [%s]: %s\n", templateName, err)
				os.Exit(1)
			}
			found = true
		}
	}

	if !found {
		fmt.Printf("ERROR: Template [%s] not found. Available templates: ", templateName)
		// List available templates
		templateNames := make([]string, 0, len(envtabTemplates.Templates))
		for name := range envtabTemplates.Templates {
			templateNames = append(templateNames, name)
		}
		for i, name := range templateNames {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(name)
		}
		fmt.Println()
		os.Exit(1)
	}

	lo.Metadata.Description = template.Description

	for _, entry := range template.Entries {
		lo.Entries[entry] = ""
	}

	return *lo
}
