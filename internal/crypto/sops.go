package crypto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gmherb/envtab/internal/utils"
	yaml "gopkg.in/yaml.v2"
)

const sopsInstallURL = "https://github.com/getsops/sops"

// checkSOPSAvailable checks if the sops command is available
func checkSOPSAvailable() error {
	_, err := exec.LookPath("sops")
	if err != nil {
		return fmt.Errorf("sops command not found. Install SOPS: %s: %w", sopsInstallURL, err)
	}
	return nil
}

// SOPSEncryptFile encrypts a file using sops command-line tool
// Returns the encrypted content as bytes
// SOPS encrypts YAML files in-place, preserving structure and adding sops: metadata
func SOPSEncryptFile(filePath string) ([]byte, error) {
	if err := checkSOPSAvailable(); err != nil {
		return nil, err
	}

	cmd := exec.Command("sops", "-e", filePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	encrypted, err := cmd.Output()
	if err != nil {
		stderrStr := stderr.String()
		if stderrStr != "" {
			return nil, fmt.Errorf("sops encryption failed: %s: %w", strings.TrimSpace(stderrStr), err)
		}
		return nil, fmt.Errorf("sops encryption failed: %w", err)
	}

	return encrypted, nil
}

// SOPSDecryptFile decrypts a file using sops command-line tool
// Returns the decrypted content as bytes
// Handles key rotation errors gracefully
func SOPSDecryptFile(filePath string) ([]byte, error) {
	if err := checkSOPSAvailable(); err != nil {
		return nil, err
	}

	cmd := exec.Command("sops", "-d", filePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	decrypted, err := cmd.Output()
	if err != nil {
		stderrStr := stderr.String()

		if exitError, ok := err.(*exec.ExitError); ok {
			// Check for key rotation or access errors
			if utils.Contains(stderrStr, "no decryption key") ||
				utils.Contains(stderrStr, "key not found") ||
				utils.Contains(stderrStr, "access denied") ||
				utils.Contains(stderrStr, "InvalidKeyException") ||
				utils.Contains(stderrStr, "no decryption key found") {
				return nil, fmt.Errorf("decryption failed: keys may have been rotated or access denied. Try re-encrypting with current keys: %w", err)
			}
			// Check if file might not be SOPS-encrypted
			if utils.Contains(stderrStr, "no sops metadata found") ||
				utils.Contains(stderrStr, "not a valid sops file") ||
				utils.Contains(stderrStr, "Error decrypting") {
				return nil, fmt.Errorf("file may not be SOPS-encrypted or is corrupted. SOPS error: %s", stderrStr)
			}
			// Include stderr for debugging
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
	if err := checkSOPSAvailable(); err != nil {
		return err
	}

	// Use sops to re-encrypt in place with current keys
	cmd := exec.Command("sops", "-r", filePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if stderrStr != "" {
			return fmt.Errorf("sops re-encryption failed: %s: %w", strings.TrimSpace(stderrStr), err)
		}
		return fmt.Errorf("sops re-encryption failed: %w", err)
	}

	return nil
}

// IsSOPSEncrypted checks if a file is encrypted with sops
// by parsing the YAML/JSON and checking if a top-level "sops" or "data" key exists
// For file-level SOPS encryption:
//   - Files encrypted with --encrypt-file typically have "data:" as a top-level key (binary/blob mode)
//   - Files may also have "sops:" metadata at the top level
//
// For value-level SOPS encryption, values start with "SOPS:" prefix (handled separately)
func IsSOPSEncrypted(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		// Try JSON if YAML parsing fails
		if jsonErr := json.Unmarshal(content, &data); jsonErr != nil {
			return false
		}
	}

	_, hasSops := data["sops"]
	_, hasData := data["data"]
	return hasSops || hasData
}

// SOPSEncryptValue encrypts a single value using sops
// Creates a temporary file, encrypts it, and returns the encrypted value with "SOPS:" prefix
func SOPSEncryptValue(value string) (string, error) {
	tmpFile, err := os.CreateTemp("", "envtab-sops-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Use YAML marshaling to properly handle special characters
	data := map[string]string{"value": value}
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to marshal value to YAML: %w", err)
	}

	if _, err := tmpFile.Write(yamlData); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	encrypted, err := SOPSEncryptFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	return "SOPS:" + string(encrypted), nil
}

// SOPSDecryptValue decrypts a SOPS-encrypted value
// The encrypted value contains the full SOPS-encrypted YAML structure including metadata
// This preserves all SOPS metadata needed for decryption
func SOPSDecryptValue(encryptedValue string) (string, error) {
	// Remove "SOPS:" prefix if present
	encrypted := strings.TrimPrefix(encryptedValue, "SOPS:")
	if encrypted == "" {
		return "", fmt.Errorf("encrypted value is empty after removing prefix")
	}

	tmpFile, err := os.CreateTemp("", "envtab-sops-decrypt-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(encrypted); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	decrypted, err := SOPSDecryptFile(tmpFile.Name())
	if err != nil {
		if utils.Contains(err.Error(), "keys may have been rotated") {
			return "", fmt.Errorf("cannot decrypt: encryption keys may have been rotated. The value was encrypted with different keys. %w", err)
		}
		return "", err
	}

	// Parse YAML to extract value (more robust than line-by-line parsing)
	var data struct {
		Value string `yaml:"value"`
	}
	if err := yaml.Unmarshal(decrypted, &data); err != nil {
		// Fallback to line-by-line parsing if YAML parsing fails
		lines := bytes.Split(decrypted, []byte("\n"))
		for _, line := range lines {
			line = bytes.TrimSpace(line)
			if bytes.HasPrefix(line, []byte("value:")) {
				valuePart := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("value:")))
				valuePart = bytes.Trim(valuePart, `"'`)
				return string(valuePart), nil
			}
		}
		return "", fmt.Errorf("failed to extract value from decrypted content: %w", err)
	}

	return data.Value, nil
}

// GetSOPSConfigPath returns the path to .sops.yaml config file
// Checks current directory and home directory
func GetSOPSConfigPath() string {
	// Check current directory
	if absPath, err := filepath.Abs(".sops.yaml"); err == nil {
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
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
