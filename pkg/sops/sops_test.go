package sops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gmherb/envtab/internal/utils"
	yaml "gopkg.in/yaml.v2"
)

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
		{"case insensitive match", "Error: no decryption key", "NO DECRYPTION KEY", true},
		{"case insensitive match 2", "InvalidKeyException occurred", "invalidkeyexception", true},
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

func TestIsSOPSEncrypted(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "sops-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		content  string
		format   string
		expected bool
	}{
		{
			name: "YAML with sops metadata",
			content: `sops:
  kms: []
  gcp_kms: []
entries:
  key: value`,
			format:   "yaml",
			expected: true,
		},
		{
			name: "YAML with data key",
			content: `data: encrypted_blob_here
metadata:
  other: value`,
			format:   "yaml",
			expected: true,
		},
		{
			name: "YAML without sops",
			content: `metadata:
  createdAt: "2023-01-01"
entries:
  key: value`,
			format:   "yaml",
			expected: false,
		},
		{
			name:     "JSON with sops metadata",
			content:  `{"sops": {}, "entries": {"key": "value"}}`,
			format:   "json",
			expected: true,
		},
		{
			name:     "JSON with data key",
			content:  `{"data": "encrypted", "other": "value"}`,
			format:   "json",
			expected: true,
		},
		{
			name:     "JSON without sops",
			content:  `{"metadata": {}, "entries": {"key": "value"}}`,
			format:   "json",
			expected: false,
		},
		{
			name:     "empty file",
			content:  "",
			format:   "yaml",
			expected: false,
		},
		{
			name:     "invalid YAML/JSON",
			content:  "not valid yaml or json {",
			format:   "yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filename string
			if tt.format == "json" {
				filename = filepath.Join(tmpDir, "test.json")
			} else {
				filename = filepath.Join(tmpDir, "test.yaml")
			}

			err := os.WriteFile(filename, []byte(tt.content), 0600)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			result := IsSOPSEncrypted(filename)
			if result != tt.expected {
				t.Errorf("IsSOPSEncrypted() = %v, want %v", result, tt.expected)
			}
		})
	}

	// Test non-existent file
	result := IsSOPSEncrypted(filepath.Join(tmpDir, "nonexistent.yaml"))
	if result != false {
		t.Errorf("IsSOPSEncrypted() for non-existent file should return false, got %v", result)
	}
}

func TestGetSOPSConfigPath(t *testing.T) {
	// Test when no config exists (should return empty string or existing path)
	// This test just verifies the function doesn't crash
	path := GetSOPSConfigPath()
	_ = path

	// Note: We don't test with actual directory changes because:
	// 1. Changing directories can cause race conditions in parallel test execution
	// 2. The function checks both current directory and home directory
	// 3. Testing with actual directory changes would require careful cleanup
	// The function is simple enough that testing it doesn't crash is sufficient
	// Real-world usage will be tested through integration tests
}

func TestCheckSOPSAvailable(t *testing.T) {
	// This test depends on whether sops is installed
	err := checkSOPSAvailable()
	// We can't assert the result since it depends on the environment
	// Just verify the function doesn't panic
	_ = err
}

func TestSOPSEncryptValue(t *testing.T) {
	// Skip if sops is not available
	if err := checkSOPSAvailable(); err != nil {
		t.Skipf("Skipping test: sops not available: %v", err)
	}

	tests := []struct {
		name  string
		value string
	}{
		{"simple value", "test_value"},
		{"value with special chars", "test@value#123"},
		{"value with newlines", "line1\nline2\nline3"},
		{"value with quotes", `value with "quotes"`},
		{"empty value", ""},
		{"unicode value", "测试值 🎉"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := SOPSEncryptValue(tt.value)
			if err != nil {
				t.Errorf("SOPSEncryptValue() error = %v", err)
				return
			}

			// Verify it has SOPS: prefix
			if !strings.HasPrefix(encrypted, "SOPS:") {
				t.Errorf("SOPSEncryptValue() should return value with SOPS: prefix")
			}

			// Verify we can decrypt it back
			decrypted, err := SOPSDecryptValue(encrypted)
			if err != nil {
				t.Errorf("SOPSDecryptValue() error = %v", err)
				return
			}

			if decrypted != tt.value {
				t.Errorf("SOPSDecryptValue() = %v, want %v", decrypted, tt.value)
			}
		})
	}
}

func TestSOPSDecryptValue(t *testing.T) {
	// Skip if sops is not available
	if err := checkSOPSAvailable(); err != nil {
		t.Skipf("Skipping test: sops not available: %v", err)
	}

	// Test decrypting a value we encrypt
	originalValue := "test_decrypt_value"
	encrypted, err := SOPSEncryptValue(originalValue)
	if err != nil {
		t.Fatalf("SOPSEncryptValue() error = %v", err)
	}

	decrypted, err := SOPSDecryptValue(encrypted)
	if err != nil {
		t.Fatalf("SOPSDecryptValue() error = %v", err)
	}

	if decrypted != originalValue {
		t.Errorf("SOPSDecryptValue() = %v, want %v", decrypted, originalValue)
	}

	// Test with value that already has SOPS: prefix (should handle gracefully)
	decrypted2, err := SOPSDecryptValue(encrypted)
	if err != nil {
		t.Fatalf("SOPSDecryptValue() error with already prefixed value = %v", err)
	}
	if decrypted2 != originalValue {
		t.Errorf("SOPSDecryptValue() with prefixed value = %v, want %v", decrypted2, originalValue)
	}

	// Test error cases
	tests := []struct {
		name           string
		encryptedValue string
		wantErr        bool
	}{
		{"empty string", "", true},
		{"empty after prefix", "SOPS:", true},
		{"invalid encrypted data", "SOPS:invalid_data", true},
		{"not SOPS encrypted", "not_encrypted", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SOPSDecryptValue(tt.encryptedValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("SOPSDecryptValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSOPSEncryptFile(t *testing.T) {
	// Skip if sops is not available
	if err := checkSOPSAvailable(); err != nil {
		t.Skipf("Skipping test: sops not available: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "sops-encrypt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.yaml")
	testContent := `metadata:
  createdAt: "2023-01-01"
entries:
  KEY1: value1
  KEY2: value2`
	err = os.WriteFile(testFile, []byte(testContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	encrypted, err := SOPSEncryptFile(testFile)
	if err != nil {
		t.Fatalf("SOPSEncryptFile() error = %v", err)
	}

	if len(encrypted) == 0 {
		t.Error("SOPSEncryptFile() should return non-empty encrypted data")
	}

	// Verify it contains sops metadata when parsed
	var data map[string]interface{}
	err = yaml.Unmarshal(encrypted, &data)
	if err != nil {
		t.Fatalf("Failed to parse encrypted YAML: %v", err)
	}

	_, hasSops := data["sops"]
	if !hasSops {
		t.Error("SOPSEncryptFile() should add sops metadata")
	}

	// Test error case: non-existent file
	_, err = SOPSEncryptFile(filepath.Join(tmpDir, "nonexistent.yaml"))
	if err == nil {
		t.Error("SOPSEncryptFile() should return error for non-existent file")
	}
}

func TestSOPSDecryptFile(t *testing.T) {
	// Skip if sops is not available
	if err := checkSOPSAvailable(); err != nil {
		t.Skipf("Skipping test: sops not available: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "sops-decrypt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and encrypt a test file
	testFile := filepath.Join(tmpDir, "test.yaml")
	testContent := `metadata:
  createdAt: "2023-01-01"
entries:
  KEY1: value1`
	err = os.WriteFile(testFile, []byte(testContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	encrypted, err := SOPSEncryptFile(testFile)
	if err != nil {
		t.Fatalf("SOPSEncryptFile() error = %v", err)
	}

	// Write encrypted content to file
	encryptedFile := filepath.Join(tmpDir, "encrypted.yaml")
	err = os.WriteFile(encryptedFile, encrypted, 0600)
	if err != nil {
		t.Fatalf("Failed to write encrypted file: %v", err)
	}

	// Decrypt it
	decrypted, err := SOPSDecryptFile(encryptedFile)
	if err != nil {
		t.Fatalf("SOPSDecryptFile() error = %v", err)
	}

	// Verify decrypted content matches original (ignoring sops metadata)
	var decryptedData map[string]interface{}
	err = yaml.Unmarshal(decrypted, &decryptedData)
	if err != nil {
		t.Fatalf("Failed to parse decrypted YAML: %v", err)
	}

	// Check that entries are present
	entries, ok := decryptedData["entries"].(map[interface{}]interface{})
	if !ok {
		t.Error("SOPSDecryptFile() should preserve entries in decrypted content")
	}

	if entries["KEY1"] != "value1" {
		t.Errorf("SOPSDecryptFile() decrypted value mismatch, got %v, want value1", entries["KEY1"])
	}

	// Test error case: non-existent file
	_, err = SOPSDecryptFile(filepath.Join(tmpDir, "nonexistent.yaml"))
	if err == nil {
		t.Error("SOPSDecryptFile() should return error for non-existent file")
	}

	// Test error case: non-encrypted file
	_, err = SOPSDecryptFile(testFile)
	if err == nil {
		t.Error("SOPSDecryptFile() should return error for non-encrypted file")
	}
}

func TestSOPSCanDecrypt(t *testing.T) {
	// Skip if sops is not available
	if err := checkSOPSAvailable(); err != nil {
		t.Skipf("Skipping test: sops not available: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "sops-can-decrypt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and encrypt a test file
	testFile := filepath.Join(tmpDir, "test.yaml")
	testContent := `metadata:
  createdAt: "2023-01-01"
entries:
  KEY1: value1`
	err = os.WriteFile(testFile, []byte(testContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	encrypted, err := SOPSEncryptFile(testFile)
	if err != nil {
		t.Fatalf("SOPSEncryptFile() error = %v", err)
	}

	// Write encrypted content to file
	encryptedFile := filepath.Join(tmpDir, "encrypted.yaml")
	err = os.WriteFile(encryptedFile, encrypted, 0600)
	if err != nil {
		t.Fatalf("Failed to write encrypted file: %v", err)
	}

	// Should be able to decrypt
	if !SOPSCanDecrypt(encryptedFile) {
		t.Error("SOPSCanDecrypt() should return true for valid encrypted file")
	}

	// Should not be able to decrypt plain file
	if SOPSCanDecrypt(testFile) {
		t.Error("SOPSCanDecrypt() should return false for non-encrypted file")
	}

	// Should return false for non-existent file
	if SOPSCanDecrypt(filepath.Join(tmpDir, "nonexistent.yaml")) {
		t.Error("SOPSCanDecrypt() should return false for non-existent file")
	}
}

func TestSOPSReencryptFile(t *testing.T) {
	// Skip if sops is not available
	if err := checkSOPSAvailable(); err != nil {
		t.Skipf("Skipping test: sops not available: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "sops-reencrypt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and encrypt a test file
	testFile := filepath.Join(tmpDir, "test.yaml")
	testContent := `metadata:
  createdAt: "2023-01-01"
entries:
  KEY1: value1`
	err = os.WriteFile(testFile, []byte(testContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	encrypted, err := SOPSEncryptFile(testFile)
	if err != nil {
		t.Fatalf("SOPSEncryptFile() error = %v", err)
	}

	// Write encrypted content to file
	encryptedFile := filepath.Join(tmpDir, "encrypted.yaml")
	err = os.WriteFile(encryptedFile, encrypted, 0600)
	if err != nil {
		t.Fatalf("Failed to write encrypted file: %v", err)
	}

	// Re-encrypt it
	err = SOPSReencryptFile(encryptedFile)
	if err != nil {
		t.Fatalf("SOPSReencryptFile() error = %v", err)
	}

	// Verify it can still be decrypted
	decrypted, err := SOPSDecryptFile(encryptedFile)
	if err != nil {
		t.Fatalf("SOPSDecryptFile() after re-encryption error = %v", err)
	}

	// Verify content is preserved
	var decryptedData map[string]interface{}
	err = yaml.Unmarshal(decrypted, &decryptedData)
	if err != nil {
		t.Fatalf("Failed to parse decrypted YAML: %v", err)
	}

	entries, ok := decryptedData["entries"].(map[interface{}]interface{})
	if !ok {
		t.Error("SOPSReencryptFile() should preserve entries")
	}

	if entries["KEY1"] != "value1" {
		t.Errorf("SOPSReencryptFile() preserved value mismatch, got %v, want value1", entries["KEY1"])
	}

	// Test error case: non-existent file
	err = SOPSReencryptFile(filepath.Join(tmpDir, "nonexistent.yaml"))
	if err == nil {
		t.Error("SOPSReencryptFile() should return error for non-existent file")
	}
}

func TestSOPSEncryptDecryptRoundTrip(t *testing.T) {
	// Skip if sops is not available
	if err := checkSOPSAvailable(); err != nil {
		t.Skipf("Skipping test: sops not available: %v", err)
	}

	originalValue := "round_trip_test_value_123"

	// Encrypt
	encrypted, err := SOPSEncryptValue(originalValue)
	if err != nil {
		t.Fatalf("SOPSEncryptValue() error = %v", err)
	}

	// Decrypt
	decrypted, err := SOPSDecryptValue(encrypted)
	if err != nil {
		t.Fatalf("SOPSDecryptValue() error = %v", err)
	}

	// Verify round trip
	if decrypted != originalValue {
		t.Errorf("Round trip failed: got %v, want %v", decrypted, originalValue)
	}
}

func TestSOPSDecryptValueWithFallbackParsing(t *testing.T) {
	// Skip if sops is not available
	if err := checkSOPSAvailable(); err != nil {
		t.Skipf("Skipping test: sops not available: %v", err)
	}

	// Test with a value that might trigger fallback parsing
	originalValue := "value:with:colons"

	encrypted, err := SOPSEncryptValue(originalValue)
	if err != nil {
		t.Fatalf("SOPSEncryptValue() error = %v", err)
	}

	decrypted, err := SOPSDecryptValue(encrypted)
	if err != nil {
		t.Fatalf("SOPSDecryptValue() error = %v", err)
	}

	if decrypted != originalValue {
		t.Errorf("Decrypt with special chars failed: got %v, want %v", decrypted, originalValue)
	}
}
