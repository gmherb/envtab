package crypto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	
	yaml "gopkg.in/yaml.v2"
)

// SOPSEncryptFile encrypts a file using sops command-line tool
// Returns the encrypted content as bytes
// SOPS encrypts YAML files in-place, preserving structure and adding sops: metadata
// It does NOT wrap content in data: | - that only happens for binary/blob encryption
// For YAML files, SOPS encrypts values in-place and adds sops: metadata at the top
// It preserves the original structure (metadata:, entries:, etc.)
func SOPSEncryptFile(filePath string) ([]byte, error) {
	// Check if sops is available
	if _, err := exec.LookPath("sops"); err != nil {
		return nil, fmt.Errorf("sops command not found: %w", err)
	}

	// Use sops to encrypt the file directly
	// sops -e encrypts the file and outputs the encrypted YAML
	cmd := exec.Command("sops", "-e", filePath)
	cmd.Stderr = os.Stderr

	encrypted, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("sops encryption failed: %w", err)
	}

	return encrypted, nil
}

// SOPSDecryptFile decrypts a file using sops command-line tool
// Returns the decrypted content as bytes
// Handles key rotation errors gracefully
func SOPSDecryptFile(filePath string) ([]byte, error) {
	// Check if sops is available
	if _, err := exec.LookPath("sops"); err != nil {
		return nil, fmt.Errorf("sops command not found. Install SOPS: https://github.com/getsops/sops: %w", err)
	}

	// Use sops to decrypt the file
	// sops -d reads from file and outputs decrypted content
	cmd := exec.Command("sops", "-d", filePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	decrypted, err := cmd.Output()
	if err != nil {
		// Get stderr output for better error messages
		stderrStr := stderr.String()
		
		// Check if error is due to key rotation or missing keys
		if exitError, ok := err.(*exec.ExitError); ok {
			if contains(stderrStr, "no decryption key") || 
			   contains(stderrStr, "key not found") ||
			   contains(stderrStr, "access denied") ||
			   contains(stderrStr, "InvalidKeyException") ||
			   contains(stderrStr, "no decryption key found") {
				return nil, fmt.Errorf("decryption failed: keys may have been rotated or access denied. Try re-encrypting with current keys: %w", err)
			}
			// Check if file might not actually be SOPS-encrypted
			if contains(stderrStr, "no sops metadata found") ||
			   contains(stderrStr, "not a valid sops file") ||
			   contains(stderrStr, "Error decrypting") {
				return nil, fmt.Errorf("file may not be SOPS-encrypted or is corrupted. SOPS error: %s", stderrStr)
			}
			// Include stderr in error message for debugging
			if stderrStr != "" {
				return nil, fmt.Errorf("sops decryption failed: %s (exit status %d)", strings.TrimSpace(stderrStr), exitError.ExitCode())
			}
		}
		return nil, fmt.Errorf("sops decryption failed: %w", err)
	}

	return decrypted, nil
}

// SOPSCanDecrypt checks if a file can be decrypted with current keys
// Returns true if decryption is possible, false otherwise
func SOPSCanDecrypt(filePath string) bool {
	_, err := SOPSDecryptFile(filePath)
	return err == nil
}

// SOPSReencryptFile re-encrypts a file with current keys
// Useful when keys have been rotated
func SOPSReencryptFile(filePath string) error {
	// Check if sops is available
	if _, err := exec.LookPath("sops"); err != nil {
		return fmt.Errorf("sops command not found: %w", err)
	}

	// First, try to decrypt to verify we can read it
	// (might fail if keys rotated, but we'll try to re-encrypt anyway)
	
	// Use sops to re-encrypt in place
	// sops -i (in-place) re-encrypts with current keys
	cmd := exec.Command("sops", "-i", "-e", filePath)
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("sops re-encryption failed: %w", err)
	}

	return nil
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// IsSOPSEncrypted checks if a file is encrypted with sops
// by parsing the YAML/JSON and checking if a top-level "sops" or "data" key exists
// For file-level SOPS encryption:
//   - Files encrypted with --sops-file typically have "data:" as a top-level key (binary/blob mode)
//   - Files may also have "sops:" metadata at the top level
// For value-level SOPS encryption, values start with "SOPS:" prefix (handled separately)
// SOPS can encrypt both YAML and JSON files
func IsSOPSEncrypted(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	// Try parsing as YAML first (most common for envtab loadouts)
	var data map[string]interface{}
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		// If YAML parsing fails, try JSON (SOPS can encrypt JSON files too)
		err = json.Unmarshal(content, &data)
		if err != nil {
			// If both fail, it's not a valid YAML/JSON file, so not SOPS-encrypted
			return false
		}
	}

	// Check if "sops" key exists at the top level (SOPS metadata)
	_, hasSops := data["sops"]
	// Check if "data" key exists at the top level (indicates --sops-file binary/blob encryption)
	_, hasData := data["data"]
	
	// File is SOPS-encrypted if it has either sops metadata or data wrapper
	return hasSops || hasData
}

// SOPSEncryptValue encrypts a single value using sops
// Creates a temporary file, encrypts it, and returns the encrypted value
func SOPSEncryptValue(value string) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "envtab-sops-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write value to temp file as YAML
	_, err = tmpFile.WriteString(fmt.Sprintf("value: %s\n", value))
	if err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Encrypt the temp file
	encrypted, err := SOPSEncryptFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	// Return as base64 encoded string with prefix
	return "SOPS:" + string(encrypted), nil
}

// SOPSDecryptValue decrypts a SOPS-encrypted value
// The encrypted value contains the full SOPS-encrypted YAML structure including metadata
// This preserves all SOPS metadata needed for decryption
func SOPSDecryptValue(encryptedValue string) (string, error) {
	// Remove prefix if present
	encrypted := encryptedValue
	if len(encryptedValue) > 5 && encryptedValue[:5] == "SOPS:" {
		encrypted = encryptedValue[5:]
	}

	// Create a temporary file with encrypted content
	// The encrypted content is the full SOPS-encrypted YAML (with metadata)
	tmpFile, err := os.CreateTemp("", "envtab-sops-decrypt-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write the full SOPS-encrypted structure (includes metadata)
	_, err = tmpFile.WriteString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Decrypt the temp file (SOPS will use metadata in the file)
	decrypted, err := SOPSDecryptFile(tmpFile.Name())
	if err != nil {
		// Provide helpful error message for key rotation
		if contains(err.Error(), "keys may have been rotated") {
			return "", fmt.Errorf("cannot decrypt: encryption keys may have been rotated. The value was encrypted with different keys. %w", err)
		}
		return "", err
	}

	// Parse YAML to extract value
	// The decrypted content should be "value: <secret>"
	lines := bytes.Split(decrypted, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("value:")) {
			// Handle both "value: secret" and "value: 'secret'" formats
			valuePart := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("value:")))
			// Remove quotes if present
			valuePart = bytes.Trim(valuePart, `"'`)
			return string(valuePart), nil
		}
	}

	return "", fmt.Errorf("failed to extract value from decrypted content: expected 'value:' field not found")
}

// SOPSCanDecryptValue checks if a SOPS-encrypted value can be decrypted
func SOPSCanDecryptValue(encryptedValue string) bool {
	_, err := SOPSDecryptValue(encryptedValue)
	return err == nil
}

// GetSOPSConfigPath returns the path to .sops.yaml config file
// Checks current directory and home directory
func GetSOPSConfigPath() string {
	// Check current directory
	if _, err := os.Stat(".sops.yaml"); err == nil {
		absPath, _ := filepath.Abs(".sops.yaml")
		return absPath
	}

	// Check home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".sops.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return ""
}


