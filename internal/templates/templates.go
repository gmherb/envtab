package templates

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

//go:embed embedded/*.env
var embeddedTemplatesFS embed.FS

var (
	embeddedTemplates     *LoadoutTemplates
	embeddedTemplatesOnce sync.Once
)

// getEmbeddedTemplates loads templates from embedded .env files
func getEmbeddedTemplates() LoadoutTemplates {
	embeddedTemplatesOnce.Do(func() {
		templates := make(map[string]LoadoutTemplate)

		entries, err := embeddedTemplatesFS.ReadDir("embedded")
		if err != nil {
			// If we can't read the embedded directory, return empty templates
			embeddedTemplates = &LoadoutTemplates{Templates: templates}
			return
		}

		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".env") {
				continue
			}

			templateName := strings.TrimSuffix(name, ".env")
			data, err := embeddedTemplatesFS.ReadFile("embedded/" + name)
			if err != nil {
				continue
			}

			// Parse .env file to get keys
			envEntries, err := backends.ParseDotenvContent(data)
			if err != nil {
				continue
			}

			// Convert map keys to slice (entries list)
			keys := make([]string, 0, len(envEntries))
			for key := range envEntries {
				keys = append(keys, key)
			}

			templates[templateName] = LoadoutTemplate{
				Description: fmt.Sprintf("Template: %s", templateName),
				Entries:     keys,
			}
		}

		embeddedTemplates = &LoadoutTemplates{Templates: templates}
	})

	return *embeddedTemplates
}

func MakeLoadoutFromTemplate(templateName string) loadout.Loadout {
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
			slog.Error("failure reading template", "template", templateName, "error", err)
			os.Exit(1)
		}
		// Parse .env file using reusable function from backends
		entries, err := backends.ParseDotenvContent(data)
		if err != nil {
			slog.Error("failure parsing .env template", "template", templateName, "error", err)
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
		embeddedTemplates := getEmbeddedTemplates()
		if embeddedTemplate, exists := embeddedTemplates.Templates[templateName]; exists {
			template = embeddedTemplate
			found = true
			slog.Debug("using embedded template", "template", templateName)
		}
	}

	if !found {
		// List available templates
		embeddedTemplates := getEmbeddedTemplates()
		templateNames := make([]string, 0, len(embeddedTemplates.Templates))
		for name := range embeddedTemplates.Templates {
			templateNames = append(templateNames, name)
		}
		slog.Error("template not found", "template", templateName, "available", templateNames)
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
