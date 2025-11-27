package templates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gmherb/envtab/internal/config"
)

func TestMakeLoadoutFromTemplate_EmbeddedTemplate(t *testing.T) {
	tests := []struct {
		name           string
		templateName   string
		expectedKeys   []string
		expectedDesc   string
		expectedValues map[string]string
	}{
		{
			name:         "aws template",
			templateName: "aws",
			expectedKeys: []string{
				"AWS_ACCESS_KEY_ID",
				"AWS_SECRET_ACCESS_KEY",
				"AWS_DEFAULT_REGION",
				"AWS_PROFILE",
			},
			expectedDesc:   "Amazon Web Services Template",
			expectedValues: map[string]string{},
		},
		{
			name:         "gcp template",
			templateName: "gcp",
			expectedKeys: []string{
				"GOOGLE_APPLICATION_CREDENTIALS",
				"GCLOUD_PROJECT",
				"GOOGLE_CLOUD_PROJECT",
			},
			expectedDesc:   "Google Cloud Platform Template",
			expectedValues: map[string]string{},
		},
		{
			name:         "pgsql template",
			templateName: "pgsql",
			expectedKeys: []string{
				"PGHOST",
				"PGPORT",
				"PGDATABASE",
				"PGUSER",
				"PGPASSWORD",
			},
			expectedDesc:   "PostgreSQL Database Template",
			expectedValues: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a temporary directory to avoid conflicts with actual envtab config
			tmpDir, err := os.MkdirTemp("", "envtab-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Save original HOME
			originalHome := os.Getenv("HOME")
			defer os.Setenv("HOME", originalHome)

			// Set HOME to temp directory so config.InitEnvtab("") uses it
			os.Setenv("HOME", tmpDir)

			// Initialize envtab directory
			envtabPath := config.InitEnvtab("")

			// Ensure no .env template file exists for this test (clean state)
			templatesDir := filepath.Join(envtabPath, "templates")
			envFile := filepath.Join(templatesDir, tt.templateName+".env")
			os.Remove(envFile) // Remove if exists, ignore error

			// Call the function
			lo := MakeLoadoutFromTemplate(tt.templateName, false)

			// Verify loadout is not nil
			if lo.Entries == nil {
				t.Fatal("MakeLoadoutFromTemplate() returned loadout with nil Entries")
			}

			// Verify description
			if lo.Metadata.Description != tt.expectedDesc {
				t.Errorf("MakeLoadoutFromTemplate() description = %v, want %v",
					lo.Metadata.Description, tt.expectedDesc)
			}

			// Verify expected keys are present
			for _, key := range tt.expectedKeys {
				if _, exists := lo.Entries[key]; !exists {
					t.Errorf("MakeLoadoutFromTemplate() missing expected key: %s", key)
				}
			}

			// Verify that entries have empty values (embedded templates don't set values)
			for key, value := range lo.Entries {
				if value != "" {
					t.Errorf("MakeLoadoutFromTemplate() entry %s has value %v, want empty string", key, value)
				}
			}

			// Verify loadout has proper structure
			if lo.Metadata.CreatedAt == "" {
				t.Error("MakeLoadoutFromTemplate() should set CreatedAt")
			}
			if lo.Metadata.LoadedAt == "" {
				t.Error("MakeLoadoutFromTemplate() should set LoadedAt")
			}
			if lo.Metadata.UpdatedAt == "" {
				t.Error("MakeLoadoutFromTemplate() should set UpdatedAt")
			}
		})
	}
}

func TestMakeLoadoutFromTemplate_DotenvTemplate(t *testing.T) {
	// Use a temporary directory to avoid conflicts with actual envtab config
	tmpDir, err := os.MkdirTemp("", "envtab-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to temp directory so config.InitEnvtab("") uses it
	os.Setenv("HOME", tmpDir)

	// Initialize envtab directory
	envtabPath := config.InitEnvtab("")

	// Create templates directory
	templatesDir := filepath.Join(envtabPath, "templates")
	err = os.MkdirAll(templatesDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Create a test .env template file
	templateName := "test-template"
	envFile := filepath.Join(templatesDir, templateName+".env")
	envContent := `KEY1=value1
KEY2=value2
KEY3=value with spaces
KEY4=value_with_equals=sign
# This is a comment
KEY5=value5
`
	err = os.WriteFile(envFile, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write .env template file: %v", err)
	}

	// Call the function
	lo := MakeLoadoutFromTemplate(templateName, false)

	// Verify loadout is not nil
	if lo.Entries == nil {
		t.Fatal("MakeLoadoutFromTemplate() returned loadout with nil Entries")
	}

	// Verify entries from .env file are present with values
	expectedEntries := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
		"KEY3": "value with spaces",
		"KEY4": "value_with_equals=sign",
		"KEY5": "value5",
	}

	for key, expectedValue := range expectedEntries {
		if actualValue, exists := lo.Entries[key]; !exists {
			t.Errorf("MakeLoadoutFromTemplate() missing expected key: %s", key)
		} else if actualValue != expectedValue {
			t.Errorf("MakeLoadoutFromTemplate() entry %s = %v, want %v", key, actualValue, expectedValue)
		}
	}

	// Verify description is set for .env template
	expectedDesc := "Template from .env file: " + templateName
	if lo.Metadata.Description != expectedDesc {
		t.Errorf("MakeLoadoutFromTemplate() description = %v, want %v",
			lo.Metadata.Description, expectedDesc)
	}

	// Verify comment was not added as an entry
	if _, exists := lo.Entries["# This is a comment"]; exists {
		t.Error("MakeLoadoutFromTemplate() should not include comments as entries")
	}
}

func TestMakeLoadoutFromTemplate_DotenvTemplateTakesPrecedence(t *testing.T) {
	// Use a temporary directory to avoid conflicts with actual envtab config
	tmpDir, err := os.MkdirTemp("", "envtab-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to temp directory so config.InitEnvtab("") uses it
	os.Setenv("HOME", tmpDir)

	// Initialize envtab directory
	envtabPath := config.InitEnvtab("")

	// Create templates directory
	templatesDir := filepath.Join(envtabPath, "templates")
	err = os.MkdirAll(templatesDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Create a .env template file with the same name as an embedded template
	templateName := "aws"
	envFile := filepath.Join(templatesDir, templateName+".env")
	envContent := `CUSTOM_KEY=custom_value
ANOTHER_KEY=another_value
`
	err = os.WriteFile(envFile, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write .env template file: %v", err)
	}

	// Call the function
	lo := MakeLoadoutFromTemplate(templateName, false)

	// Verify that .env template takes precedence (should have custom keys, not AWS keys)
	if lo.Entries["CUSTOM_KEY"] != "custom_value" {
		t.Error("MakeLoadoutFromTemplate() should use .env template when both exist")
	}

	if lo.Entries["AWS_ACCESS_KEY_ID"] != "" {
		t.Error("MakeLoadoutFromTemplate() should not use embedded template when .env template exists")
	}

	// Verify description indicates it's from .env file
	expectedDesc := "Template from .env file: " + templateName
	if lo.Metadata.Description != expectedDesc {
		t.Errorf("MakeLoadoutFromTemplate() description = %v, want %v",
			lo.Metadata.Description, expectedDesc)
	}
}

func TestMakeLoadoutFromTemplate_InvalidDotenvTemplate(t *testing.T) {
	// Use a temporary directory to avoid conflicts with actual envtab config
	tmpDir, err := os.MkdirTemp("", "envtab-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to temp directory so config.InitEnvtab("") uses it
	os.Setenv("HOME", tmpDir)

	// Initialize envtab directory
	envtabPath := config.InitEnvtab("")

	// Create templates directory
	templatesDir := filepath.Join(envtabPath, "templates")
	err = os.MkdirAll(templatesDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Create a test .env template file with invalid content
	templateName := "invalid-template"
	envFile := filepath.Join(templatesDir, templateName+".env")
	// Write a file that exists but will cause parse error
	// (empty file should be fine, but let's test with something that might cause issues)
	envContent := `KEY1=value1
INVALID LINE WITHOUT EQUALS
KEY2=value2
`
	err = os.WriteFile(envFile, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write .env template file: %v", err)
	}

	// Note: This test verifies that the function handles parsing errors
	// The function calls os.Exit(1) on error, so we can't test the error path directly
	// But we can verify that valid entries are still parsed
	lo := MakeLoadoutFromTemplate(templateName, false)

	// The function should still parse valid entries
	if lo.Entries["KEY1"] != "value1" {
		t.Error("MakeLoadoutFromTemplate() should parse valid entries even with some invalid lines")
	}
	if lo.Entries["KEY2"] != "value2" {
		t.Error("MakeLoadoutFromTemplate() should parse valid entries even with some invalid lines")
	}
}

func TestMakeLoadoutFromTemplate_AllEmbeddedTemplates(t *testing.T) {
	// Test that all embedded templates can be loaded
	// Get list of all embedded templates
	allTemplates := []string{
		"aws", "gcp", "azure", "pgsql", "mysql", "mongodb",
		"elasticsearch", "kafka", "rabbitmq", "redis", "memcached",
		"docker", "k8s", "vault", "consul", "terraform", "terragrunt",
		"helm", "ansible", "packer", "vagrant", "jira-cli",
		"python", "go", "rust", "c", "git", "github", "gitlab",
		"proxy", "wireguard", "sops", "yq", "jq", "jo",
		"etcd", "k6",
	}

	// Use a temporary directory to avoid conflicts with actual envtab config
	tmpDir, err := os.MkdirTemp("", "envtab-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to temp directory so config.InitEnvtab("") uses it
	os.Setenv("HOME", tmpDir)

	// Initialize envtab directory
	envtabPath := config.InitEnvtab("")
	templatesDir := filepath.Join(envtabPath, "templates")

	for _, templateName := range allTemplates {
		t.Run(templateName, func(t *testing.T) {
			// Ensure no .env template file exists for this test (clean state)
			envFile := filepath.Join(templatesDir, templateName+".env")
			os.Remove(envFile) // Remove if exists, ignore error

			lo := MakeLoadoutFromTemplate(templateName, false)

			// Verify loadout is initialized
			if lo.Entries == nil {
				t.Fatal("MakeLoadoutFromTemplate() returned loadout with nil Entries")
			}

			// Verify description is set
			if lo.Metadata.Description == "" {
				t.Error("MakeLoadoutFromTemplate() should set description for embedded template")
			}

			// Verify at least one entry exists (all templates should have entries)
			if len(lo.Entries) == 0 {
				t.Error("MakeLoadoutFromTemplate() should create loadout with entries")
			}

			// Verify all entries have empty values (embedded templates don't set values)
			for key, value := range lo.Entries {
				if value != "" {
					t.Errorf("MakeLoadoutFromTemplate() entry %s has value %v, want empty string", key, value)
				}
			}
		})
	}
}

func TestMakeLoadoutFromTemplate_LoadoutStructure(t *testing.T) {
	// Use a temporary directory to avoid conflicts with actual envtab config
	tmpDir, err := os.MkdirTemp("", "envtab-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to temp directory so config.InitEnvtab("") uses it
	os.Setenv("HOME", tmpDir)

	// Initialize envtab directory
	_ = config.InitEnvtab("")

	// Test with embedded template
	lo := MakeLoadoutFromTemplate("aws", false)

	// Verify it's a valid loadout structure
	if lo.Metadata.CreatedAt == "" {
		t.Error("MakeLoadoutFromTemplate() should set CreatedAt")
	}
	if lo.Metadata.LoadedAt == "" {
		t.Error("MakeLoadoutFromTemplate() should set LoadedAt")
	}
	if lo.Metadata.UpdatedAt == "" {
		t.Error("MakeLoadoutFromTemplate() should set UpdatedAt")
	}
	if lo.Metadata.Login != false {
		t.Error("MakeLoadoutFromTemplate() should set Login to false")
	}
	if lo.Metadata.Tags == nil {
		t.Error("MakeLoadoutFromTemplate() should initialize Tags")
	}
	if lo.Entries == nil {
		t.Error("MakeLoadoutFromTemplate() should initialize Entries")
	}
}

func TestMakeLoadoutFromTemplate_ForceParameter(t *testing.T) {
	// The force parameter is currently not used in the function
	// but we test that it doesn't cause issues
	tmpDir, err := os.MkdirTemp("", "envtab-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to temp directory so config.InitEnvtab("") uses it
	os.Setenv("HOME", tmpDir)

	// Initialize envtab directory
	_ = config.InitEnvtab("")

	// Test with force=true
	lo1 := MakeLoadoutFromTemplate("aws", true)
	if lo1.Entries == nil {
		t.Fatal("MakeLoadoutFromTemplate() with force=true returned loadout with nil Entries")
	}

	// Test with force=false
	lo2 := MakeLoadoutFromTemplate("aws", false)
	if lo2.Entries == nil {
		t.Fatal("MakeLoadoutFromTemplate() with force=false returned loadout with nil Entries")
	}

	// Both should work the same way (force is not currently used)
	if len(lo1.Entries) != len(lo2.Entries) {
		t.Error("MakeLoadoutFromTemplate() should behave the same regardless of force parameter")
	}
}

// TestLoadoutTemplate_Structure tests the LoadoutTemplate struct
func TestLoadoutTemplate_Structure(t *testing.T) {
	template := LoadoutTemplate{
		Entries:     []string{"KEY1", "KEY2"},
		Description: "Test template",
	}

	if len(template.Entries) != 2 {
		t.Errorf("LoadoutTemplate.Entries length = %d, want 2", len(template.Entries))
	}
	if template.Description != "Test template" {
		t.Errorf("LoadoutTemplate.Description = %v, want Test template", template.Description)
	}
}

// TestLoadoutTemplates_Structure tests the LoadoutTemplates struct
func TestLoadoutTemplates_Structure(t *testing.T) {
	templates := LoadoutTemplates{
		Templates: map[string]LoadoutTemplate{
			"test": {
				Entries:     []string{"KEY1"},
				Description: "Test",
			},
		},
	}

	if len(templates.Templates) != 1 {
		t.Errorf("LoadoutTemplates.Templates length = %d, want 1", len(templates.Templates))
	}
	if templates.Templates["test"].Description != "Test" {
		t.Errorf("LoadoutTemplates.Templates[\"test\"].Description = %v, want Test",
			templates.Templates["test"].Description)
	}
}

