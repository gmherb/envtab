package crypto

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SOPSEncryptFile encrypts a file using sops command-line tool
// Returns the encrypted content as bytes
func SOPSEncryptFile(filePath string) ([]byte, error) {
	// Check if sops is available
	if _, err := exec.LookPath("sops"); err != nil {
		return nil, fmt.Errorf("sops command not found: %w", err)
	}

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Use sops to encrypt the content
	// sops -e reads from stdin and outputs encrypted content
	cmd := exec.Command("sops", "-e", "/dev/stdin")
	cmd.Stdin = bytes.NewReader(content)
	cmd.Stderr = os.Stderr

	encrypted, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("sops encryption failed: %w", err)
	}

	return encrypted, nil
}

// SOPSDecryptFile decrypts a file using sops command-line tool
// Returns the decrypted content as bytes
func SOPSDecryptFile(filePath string) ([]byte, error) {
	// Check if sops is available
	if _, err := exec.LookPath("sops"); err != nil {
		return nil, fmt.Errorf("sops command not found: %w", err)
	}

	// Use sops to decrypt the file
	// sops -d reads from file and outputs decrypted content
	cmd := exec.Command("sops", "-d", filePath)
	cmd.Stderr = os.Stderr

	decrypted, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("sops decryption failed: %w", err)
	}

	return decrypted, nil
}

// IsSOPSEncrypted checks if a file is encrypted with sops
// by checking if it starts with sops metadata
func IsSOPSEncrypted(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	// SOPS encrypted files typically start with "sops:" in YAML
	// or have specific JSON structure
	return bytes.Contains(content, []byte("sops:")) ||
		bytes.Contains(content, []byte("\"sops\""))
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
func SOPSDecryptValue(encryptedValue string) (string, error) {
	// Remove prefix if present
	encrypted := encryptedValue
	if len(encryptedValue) > 5 && encryptedValue[:5] == "SOPS:" {
		encrypted = encryptedValue[5:]
	}

	// Create a temporary file with encrypted content
	tmpFile, err := os.CreateTemp("", "envtab-sops-decrypt-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Decrypt the temp file
	decrypted, err := SOPSDecryptFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	// Parse YAML to extract value
	// Simple extraction - assumes format "value: <decrypted>"
	lines := bytes.Split(decrypted, []byte("\n"))
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("value: ")) {
			return string(bytes.TrimPrefix(line, []byte("value: "))), nil
		}
	}

	return "", fmt.Errorf("failed to extract value from decrypted content")
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

