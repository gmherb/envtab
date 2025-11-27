package templates

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/config"
	"github.com/gmherb/envtab/internal/loadout"
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
	var isDotenvTemplate bool // Track if we loaded from .env file

	// First, check for user-provided .env template files (custom templates)
	dotenvPath := filepath.Join(config.InitEnvtab(""), "templates/"+templateName+".env")
	if _, err := os.Stat(dotenvPath); err == nil {
		slog.Debug("using .env template", "template", templateName, "path", dotenvPath)

		data, err := os.ReadFile(dotenvPath)
		if err != nil {
			fmt.Printf("ERROR: Failure reading template [%s]: %s\n", templateName, err)
			os.Exit(1)
		}
		// Parse .env file using reusable function from backends
		entries, err := backends.ParseDotenvContent(data)
		if err != nil {
			fmt.Printf("ERROR: Failure parsing .env template [%s]: %s\n", templateName, err)
			os.Exit(1)
		}
		// Populate loadout entries directly with values from .env
		for key, value := range entries {
			lo.Entries[key] = value
		}
		// Set description
		template.Description = fmt.Sprintf("Template from .env file: %s", templateName)
		isDotenvTemplate = true
		found = true
	} else {
		// Fall back to embedded templates
		if embeddedTemplate, exists := envtabTemplates.Templates[templateName]; exists {
			template = embeddedTemplate
			found = true
			slog.Debug("using embedded template", "template", templateName)
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

	// Only populate from template.Entries if we didn't load from .env file
	// (for .env files, we already populated with actual values above)
	if !isDotenvTemplate {
		for _, entry := range template.Entries {
			lo.Entries[entry] = ""
		}
	}

	return *lo
}
