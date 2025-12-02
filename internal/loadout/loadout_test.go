package loadout

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gmherb/envtab/internal/utils"
	"github.com/gmherb/envtab/pkg/sops"
)

func TestValidateLoadout(t *testing.T) {
	tests := []struct {
		name    string
		loadout *Loadout
		wantErr bool
	}{
		{
			name: "valid loadout",
			loadout: &Loadout{
				Metadata: LoadoutMetadata{},
				Entries: map[string]string{
					"KEY1": "value1",
					"KEY2": "value2",
				},
			},
			wantErr: false,
		},
		{
			name:    "nil loadout",
			loadout: nil,
			wantErr: true,
		},
		{
			name: "loadout with empty key",
			loadout: &Loadout{
				Metadata: LoadoutMetadata{},
				Entries: map[string]string{
					"":     "value1",
					"KEY2": "value2",
				},
			},
			wantErr: true,
		},
		{
			name: "empty entries",
			loadout: &Loadout{
				Metadata: LoadoutMetadata{},
				Entries:  map[string]string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLoadout(tt.loadout)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLoadout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLoadoutYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid YAML without duplicates",
			yaml: `metadata:
  createdAt: "2023-01-01T00:00:00Z"
entries:
  KEY1: value1
  KEY2: value2
  KEY3: value3`,
			wantErr: false,
		},
		{
			name: "YAML with duplicate keys in entries",
			yaml: `metadata:
  createdAt: "2023-01-01T00:00:00Z"
entries:
  KEY1: value1
  KEY2: value2
  KEY1: value3`,
			wantErr: true,
		},
		{
			name: "YAML with comments",
			yaml: `metadata:
  createdAt: "2023-01-01T00:00:00Z"
entries:
  # This is a comment
  KEY1: value1
  KEY2: value2`,
			wantErr: false,
		},
		{
			name: "YAML with empty lines",
			yaml: `metadata:
  createdAt: "2023-01-01T00:00:00Z"

entries:

  KEY1: value1

  KEY2: value2`,
			wantErr: false,
		},
		{
			name: "YAML without entries section",
			yaml: `metadata:
  createdAt: "2023-01-01T00:00:00Z"`,
			wantErr: false,
		},
		{
			name: "YAML with nested entries",
			yaml: `metadata:
  createdAt: "2023-01-01T00:00:00Z"
entries:
  KEY1: value1
  KEY2:
    nested: value`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLoadoutYAML([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLoadoutYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInitLoadout(t *testing.T) {
	loadout := InitLoadout()

	if loadout == nil {
		t.Fatal("InitLoadout() returned nil")
	}

	if loadout.Metadata.CreatedAt == "" {
		t.Error("InitLoadout() should set CreatedAt")
	}

	if loadout.Metadata.LoadedAt == "" {
		t.Error("InitLoadout() should set LoadedAt")
	}

	if loadout.Metadata.UpdatedAt == "" {
		t.Error("InitLoadout() should set UpdatedAt")
	}

	if loadout.Metadata.Login != false {
		t.Error("InitLoadout() should set Login to false")
	}

	if loadout.Metadata.Tags == nil {
		t.Error("InitLoadout() should initialize Tags as empty slice")
	}

	if len(loadout.Metadata.Tags) != 0 {
		t.Error("InitLoadout() should initialize Tags as empty slice")
	}

	if loadout.Metadata.Description != "" {
		t.Error("InitLoadout() should set Description to empty string")
	}

	if loadout.Entries == nil {
		t.Error("InitLoadout() should initialize Entries map")
	}

	if len(loadout.Entries) != 0 {
		t.Error("InitLoadout() should initialize Entries as empty map")
	}

	// Verify timestamps are valid RFC3339
	_, err := time.Parse(time.RFC3339, loadout.Metadata.CreatedAt)
	if err != nil {
		t.Errorf("InitLoadout() CreatedAt is not valid RFC3339: %v", err)
	}
}

func TestUpdateEntry(t *testing.T) {
	loadout := InitLoadout()

	err := loadout.UpdateEntry("TEST_KEY", "test_value")
	if err != nil {
		t.Errorf("UpdateEntry() error = %v", err)
	}

	if loadout.Entries["TEST_KEY"] != "test_value" {
		t.Errorf("UpdateEntry() failed to update entry, got %v, want test_value", loadout.Entries["TEST_KEY"])
	}

	if loadout.Metadata.UpdatedAt == "" {
		t.Error("UpdateEntry() should update UpdatedAt")
	}
}

func TestUpdateTags(t *testing.T) {
	loadout := InitLoadout()
	loadout.Metadata.Tags = []string{"tag1", "tag2"}

	err := loadout.UpdateTags([]string{"tag3", "tag4"})
	if err != nil {
		t.Errorf("UpdateTags() error = %v", err)
	}

	// Check that all tags are present
	expectedTags := []string{"tag1", "tag2", "tag3", "tag4"}
	for _, expectedTag := range expectedTags {
		found := false
		for _, tag := range loadout.Metadata.Tags {
			if tag == expectedTag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("UpdateTags() missing tag: %s", expectedTag)
		}
	}

	if loadout.Metadata.UpdatedAt == "" {
		t.Error("UpdateTags() should update UpdatedAt")
	}
}

func TestReplaceTags(t *testing.T) {
	loadout := InitLoadout()
	loadout.Metadata.Tags = []string{"tag1", "tag2"}

	err := loadout.ReplaceTags([]string{"tag3", "tag4"})
	if err != nil {
		t.Errorf("ReplaceTags() error = %v", err)
	}

	if len(loadout.Metadata.Tags) != 2 {
		t.Errorf("ReplaceTags() should have 2 tags, got %d", len(loadout.Metadata.Tags))
	}

	if loadout.Metadata.Tags[0] != "tag3" || loadout.Metadata.Tags[1] != "tag4" {
		t.Errorf("ReplaceTags() failed to replace tags, got %v", loadout.Metadata.Tags)
	}

	if loadout.Metadata.UpdatedAt == "" {
		t.Error("ReplaceTags() should update UpdatedAt")
	}
}

func TestRemoveTags(t *testing.T) {
	loadout := InitLoadout()
	loadout.Metadata.Tags = []string{"tag1", "tag2", "tag3", "tag4"}

	err := loadout.RemoveTags([]string{"tag2", "tag4"})
	if err != nil {
		t.Errorf("RemoveTags() error = %v", err)
	}

	if len(loadout.Metadata.Tags) != 2 {
		t.Errorf("RemoveTags() should have 2 tags, got %d", len(loadout.Metadata.Tags))
	}

	// Check that removed tags are not present
	for _, tag := range loadout.Metadata.Tags {
		if tag == "tag2" || tag == "tag4" {
			t.Errorf("RemoveTags() should have removed tag %s", tag)
		}
	}

	// Check that remaining tags are present
	expectedTags := []string{"tag1", "tag3"}
	for _, expectedTag := range expectedTags {
		found := false
		for _, tag := range loadout.Metadata.Tags {
			if tag == expectedTag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RemoveTags() missing expected tag: %s", expectedTag)
		}
	}

	if loadout.Metadata.UpdatedAt == "" {
		t.Error("RemoveTags() should update UpdatedAt")
	}
}

func TestUpdateDescription(t *testing.T) {
	loadout := InitLoadout()

	err := loadout.UpdateDescription("test description")
	if err != nil {
		t.Errorf("UpdateDescription() error = %v", err)
	}

	if loadout.Metadata.Description != "test description" {
		t.Errorf("UpdateDescription() failed, got %v, want test description", loadout.Metadata.Description)
	}

	if loadout.Metadata.UpdatedAt == "" {
		t.Error("UpdateDescription() should update UpdatedAt")
	}
}

func TestUpdateLogin(t *testing.T) {
	loadout := InitLoadout()

	err := loadout.UpdateLogin(true)
	if err != nil {
		t.Errorf("UpdateLogin() error = %v", err)
	}

	if loadout.Metadata.Login != true {
		t.Errorf("UpdateLogin() failed, got %v, want true", loadout.Metadata.Login)
	}

	err = loadout.UpdateLogin(false)
	if err != nil {
		t.Errorf("UpdateLogin() error = %v", err)
	}

	if loadout.Metadata.Login != false {
		t.Errorf("UpdateLogin() failed, got %v, want false", loadout.Metadata.Login)
	}

	if loadout.Metadata.UpdatedAt == "" {
		t.Error("UpdateLogin() should update UpdatedAt")
	}
}

func TestUpdateUpdatedAt(t *testing.T) {
	loadout := InitLoadout()
	originalTime := loadout.Metadata.UpdatedAt

	// Wait a bit to ensure time difference (at least 1 second for RFC3339 precision)
	time.Sleep(1100 * time.Millisecond)

	err := loadout.UpdateUpdatedAt()
	if err != nil {
		t.Errorf("UpdateUpdatedAt() error = %v", err)
	}

	if loadout.Metadata.UpdatedAt == originalTime {
		t.Error("UpdateUpdatedAt() should update the timestamp")
	}

	// Verify it's valid RFC3339
	_, err = time.Parse(time.RFC3339, loadout.Metadata.UpdatedAt)
	if err != nil {
		t.Errorf("UpdateUpdatedAt() UpdatedAt is not valid RFC3339: %v", err)
	}
}

func TestUpdateLoadedAt(t *testing.T) {
	loadout := InitLoadout()
	originalTime := loadout.Metadata.LoadedAt

	// Wait a bit to ensure time difference (at least 1 second for RFC3339 precision)
	time.Sleep(1100 * time.Millisecond)

	err := loadout.UpdateLoadedAt()
	if err != nil {
		t.Errorf("UpdateLoadedAt() error = %v", err)
	}

	if loadout.Metadata.LoadedAt == originalTime {
		t.Error("UpdateLoadedAt() should update the timestamp")
	}

	// Verify it's valid RFC3339
	_, err = time.Parse(time.RFC3339, loadout.Metadata.LoadedAt)
	if err != nil {
		t.Errorf("UpdateLoadedAt() LoadedAt is not valid RFC3339: %v", err)
	}
}

func TestCompareLoadouts(t *testing.T) {
	baseTime := "2023-01-01T00:00:00Z"

	tests := []struct {
		name string
		old  Loadout
		new  Loadout
		want bool
	}{
		{
			name: "identical loadouts",
			old: Loadout{
				Metadata: LoadoutMetadata{
					CreatedAt:   baseTime,
					LoadedAt:    baseTime,
					UpdatedAt:   baseTime,
					Login:       false,
					Tags:        []string{"tag1"},
					Description: "desc",
				},
				Entries: map[string]string{"KEY1": "value1"},
			},
			new: Loadout{
				Metadata: LoadoutMetadata{
					CreatedAt:   baseTime,
					LoadedAt:    baseTime,
					UpdatedAt:   baseTime,
					Login:       false,
					Tags:        []string{"tag1"},
					Description: "desc",
				},
				Entries: map[string]string{"KEY1": "value1"},
			},
			want: false,
		},
		{
			name: "different CreatedAt",
			old: Loadout{
				Metadata: LoadoutMetadata{CreatedAt: baseTime},
			},
			new: Loadout{
				Metadata: LoadoutMetadata{CreatedAt: "2023-01-02T00:00:00Z"},
			},
			want: true,
		},
		{
			name: "different LoadedAt",
			old: Loadout{
				Metadata: LoadoutMetadata{LoadedAt: baseTime},
			},
			new: Loadout{
				Metadata: LoadoutMetadata{LoadedAt: "2023-01-02T00:00:00Z"},
			},
			want: true,
		},
		{
			name: "different UpdatedAt",
			old: Loadout{
				Metadata: LoadoutMetadata{UpdatedAt: baseTime},
			},
			new: Loadout{
				Metadata: LoadoutMetadata{UpdatedAt: "2023-01-02T00:00:00Z"},
			},
			want: true,
		},
		{
			name: "different Login",
			old: Loadout{
				Metadata: LoadoutMetadata{Login: false},
			},
			new: Loadout{
				Metadata: LoadoutMetadata{Login: true},
			},
			want: true,
		},
		{
			name: "different Tags length",
			old: Loadout{
				Metadata: LoadoutMetadata{Tags: []string{"tag1"}},
			},
			new: Loadout{
				Metadata: LoadoutMetadata{Tags: []string{"tag1", "tag2"}},
			},
			want: true,
		},
		{
			name: "different Tags content",
			old: Loadout{
				Metadata: LoadoutMetadata{Tags: []string{"tag1"}},
			},
			new: Loadout{
				Metadata: LoadoutMetadata{Tags: []string{"tag2"}},
			},
			want: true,
		},
		{
			name: "different Description",
			old: Loadout{
				Metadata: LoadoutMetadata{Description: "desc1"},
			},
			new: Loadout{
				Metadata: LoadoutMetadata{Description: "desc2"},
			},
			want: true,
		},
		{
			name: "different Entries length",
			old: Loadout{
				Entries: map[string]string{"KEY1": "value1"},
			},
			new: Loadout{
				Entries: map[string]string{"KEY1": "value1", "KEY2": "value2"},
			},
			want: true,
		},
		{
			name: "different Entries values",
			old: Loadout{
				Entries: map[string]string{"KEY1": "value1"},
			},
			new: Loadout{
				Entries: map[string]string{"KEY1": "value2"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareLoadouts(tt.old, tt.new)
			if got != tt.want {
				t.Errorf("CompareLoadouts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintLoadout(t *testing.T) {
	loadout := InitLoadout()
	loadout.Entries["TEST_KEY"] = "test_value"

	// Capture output would require more complex setup, so we just test for errors
	err := loadout.PrintLoadout()
	if err != nil {
		t.Errorf("PrintLoadout() error = %v", err)
	}
}

func TestExport(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set a test PATH
	testPath := "/test/path1:/test/path2"
	os.Setenv("PATH", testPath)

	loadout := InitLoadout()
	loadout.Entries["TEST_VAR"] = "test_value"
	loadout.Entries["PATH"] = "/new/path:$PATH"

	// Export should not panic
	loadout.Export()

	// Verify PATH was updated
	currentPath := os.Getenv("PATH")
	if !strings.Contains(currentPath, "/new/path") {
		t.Error("Export() should add new path to PATH")
	}

	// Test with empty value (should be skipped)
	loadout2 := InitLoadout()
	loadout2.Entries["EMPTY_VAR"] = ""
	loadout2.Export() // Should not panic

	// Verify LoadedAt was updated
	if loadout.Metadata.LoadedAt == "" {
		t.Error("Export() should update LoadedAt")
	}
}

func TestExportWithSOPSEncryptedPATH(t *testing.T) {
	// This test verifies the fix from 0.1.4-alpha:
	// SOPS-encrypted PATH values should be decrypted before PATH expansion
	// Skip test if SOPS is not available

	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set a test PATH
	testPath := "/test/path1:/test/path2"
	os.Setenv("PATH", testPath)

	// Create a SOPS-encrypted PATH value that contains $PATH
	// First, encrypt the value "/new/path:$PATH"
	plainValue := "/new/path:$PATH"
	encrypted, err := sops.SOPSEncryptValue(plainValue)
	if err != nil {
		t.Skipf("Cannot encrypt value for test (SOPS may not be configured): %v", err)
	}

	loadout := InitLoadout()
	loadout.Entries["PATH"] = encrypted

	// Export should decrypt the value first, then expand $PATH
	loadout.Export()

	// Verify PATH was updated with both the new path and existing paths
	currentPath := os.Getenv("PATH")
	if !strings.Contains(currentPath, "/new/path") {
		t.Error("Export() should add new path from SOPS-encrypted PATH value")
	}
	if !strings.Contains(currentPath, "/test/path1") {
		t.Error("Export() should preserve existing paths when expanding $PATH in SOPS-encrypted value")
	}
	if !strings.Contains(currentPath, "/test/path2") {
		t.Error("Export() should preserve existing paths when expanding $PATH in SOPS-encrypted value")
	}

	// Verify LoadedAt was updated
	if loadout.Metadata.LoadedAt == "" {
		t.Error("Export() should update LoadedAt")
	}
}

func TestExportWithEmptyValues(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set a test PATH
	testPath := "/test/path1:/test/path2"
	os.Setenv("PATH", testPath)

	tests := []struct {
		name           string
		entries        map[string]string
		expectedInPath []string
		notInPath      []string
		shouldExport   map[string]bool // key -> should be exported
	}{
		{
			name: "empty PATH entry should be skipped",
			entries: map[string]string{
				"PATH": "",
				"VAR1": "value1",
			},
			expectedInPath: []string{"/test/path1", "/test/path2"},
			notInPath:      []string{},
			shouldExport: map[string]bool{
				"VAR1": true,
				"PATH": false, // Empty PATH should not be exported
			},
		},
		{
			name: "PATH with empty segments should skip empty parts",
			entries: map[string]string{
				"PATH": "/new/path1::/new/path2",
			},
			expectedInPath: []string{"/test/path1", "/test/path2", "/new/path1", "/new/path2"},
			notInPath:      []string{""},
			shouldExport: map[string]bool{
				"PATH": true,
			},
		},
		{
			name: "PATH with leading/trailing colons should be trimmed",
			entries: map[string]string{
				"PATH": ":/new/path1:/new/path2:",
			},
			expectedInPath: []string{"/test/path1", "/test/path2", "/new/path1", "/new/path2"},
			notInPath:      []string{""},
			shouldExport: map[string]bool{
				"PATH": true,
			},
		},
		{
			name: "PATH with only empty segments should preserve existing PATH",
			entries: map[string]string{
				"PATH": "::",
			},
			expectedInPath: []string{"/test/path1", "/test/path2"},
			notInPath:      []string{""},
			shouldExport: map[string]bool{
				"PATH": true,
			},
		},
		{
			name: "empty non-PATH entries should be skipped",
			entries: map[string]string{
				"EMPTY_VAR": "",
				"VAR1":      "value1",
				"VAR2":      "value2",
			},
			expectedInPath: []string{"/test/path1", "/test/path2"},
			notInPath:      []string{},
			shouldExport: map[string]bool{
				"EMPTY_VAR": false,
				"VAR1":      true,
				"VAR2":      true,
			},
		},
		{
			name: "PATH with $PATH expansion and empty segments",
			entries: map[string]string{
				"PATH": "/new/path1::$PATH:/new/path2:",
			},
			expectedInPath: []string{"/test/path1", "/test/path2", "/new/path1", "/new/path2"},
			notInPath:      []string{""},
			shouldExport: map[string]bool{
				"PATH": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset PATH for each test
			os.Setenv("PATH", testPath)

			loadout := InitLoadout()
			for k, v := range tt.entries {
				loadout.Entries[k] = v
			}

			// Capture stdout output
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			// Capture output in a goroutine
			var capturedOutput bytes.Buffer
			done := make(chan error)
			go func() {
				_, err := io.Copy(&capturedOutput, r)
				r.Close()
				done <- err
			}()

			// Call Export() which will write to stdout
			loadout.Export()

			// Close write end and restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Wait for capture to complete
			if err := <-done; err != nil {
				t.Fatalf("Failed to capture output: %v", err)
			}

			// Parse captured output to extract exported variables
			exportedVars := make(map[string]bool)
			output := capturedOutput.String()
			// Match lines like "export VAR=value" or "export PATH=..."
			exportRegex := regexp.MustCompile(`^export (\w+)=`)
			for _, line := range strings.Split(output, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				matches := exportRegex.FindStringSubmatch(line)
				if len(matches) > 1 {
					varName := matches[1]
					exportedVars[varName] = true
				}
			}

			// Validate shouldExport expectations
			if tt.shouldExport != nil {
				for varName, shouldBeExported := range tt.shouldExport {
					wasExported := exportedVars[varName]
					if shouldBeExported && !wasExported {
						t.Errorf("Expected %q to be exported, but it was not. Output:\n%s", varName, output)
					}
					if !shouldBeExported && wasExported {
						t.Errorf("Expected %q NOT to be exported, but it was. Output:\n%s", varName, output)
					}
				}
			}

			// Verify PATH contents
			currentPath := os.Getenv("PATH")
			for _, expected := range tt.expectedInPath {
				if !strings.Contains(currentPath, expected) {
					t.Errorf("Expected PATH to contain %q, got %q", expected, currentPath)
				}
			}

			for _, notExpected := range tt.notInPath {
				if notExpected != "" && strings.Contains(currentPath, notExpected) {
					t.Errorf("Expected PATH not to contain empty segment, got %q", currentPath)
				}
			}

			// Verify PATH doesn't have double colons
			if strings.Contains(currentPath, "::") {
				t.Errorf("PATH should not contain double colons, got %q", currentPath)
			}

			// Verify LoadedAt was updated
			if loadout.Metadata.LoadedAt == "" {
				t.Error("Export() should update LoadedAt")
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name string
		s    string
		sub  string
		want bool
	}{
		{"contains substring", "Hello World", "world", true},
		{"contains substring case insensitive", "Hello World", "WORLD", true},
		{"does not contain", "Hello World", "foo", false},
		{"empty string", "", "test", false},
		{"empty substring", "test", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.Contains(tt.s, tt.sub)
			if got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
