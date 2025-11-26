// Test file for env.go
package env

import (
	"fmt"
	"strings"
	"testing"
)

func TestSet(t *testing.T) {
	e := NewEnv()
	e.Set("TEST=TEST")
	if e.Env["TEST"] != "TEST" {
		t.Errorf("Expected %s, got %s", "TEST", e.Env["TEST"])
	}
	println("This is a", e.Env["TEST"])
}

func TestPopulate(t *testing.T) {
	e := NewEnv()
	e.Populate()
	if e.Env["PATH"] == "" {
		t.Errorf("Expected %s, got %s", "PATH", e.Env["PATH"])
	} else {
		println("PATH:", e.Env["PATH"])
	}
	if e.Env["HOME"] == "" {
		t.Errorf("Expected %s, got %s", "HOME", e.Env["HOME"])
	} else {
		println("HOME:", e.Env["HOME"])
	}
	if e.Env["USER"] == "" {
		t.Errorf("Expected %s, got %s", "USER", e.Env["USER"])
	} else {
		println("USER:", e.Env["USER"])
	}
	if e.Env["PWD"] == "" {
		t.Errorf("Expected %s, got %s", "PWD", e.Env["PWD"])
	} else {
		println("PWD:", e.Env["PWD"])
	}
	if e.Env["SHLVL"] == "" {
		t.Errorf("Expected %s, got %s", "SHLVL", e.Env["SHLVL"])
	} else {
		println("SHLVL:", e.Env["SHLVL"])
	}
}

func TestPrint(t *testing.T) {
	e := NewEnv()
	e.Populate()
	e.Dump()
}

func TestGet(t *testing.T) {
	e := NewEnv()
	e.Set("TEST_KEY=test_value")

	value := e.Get("TEST_KEY")
	if value != "test_value" {
		t.Errorf("Get() = %v, want test_value", value)
	}

	// Test non-existent key
	value = e.Get("NON_EXISTENT")
	if value != "" {
		t.Errorf("Get() for non-existent key should return empty string, got %v", value)
	}
}

func TestCompare(t *testing.T) {
	e := NewEnv()
	e.Set("TEST_KEY=test_value")
	e.Set("PATH=/usr/bin:/usr/local/bin")

	// Test exact match
	if !e.Compare("TEST_KEY", "test_value") {
		t.Error("Compare() should return true for exact match")
	}

	// Test non-match
	if e.Compare("TEST_KEY", "wrong_value") {
		t.Error("Compare() should return false for non-match")
	}

	// Test PATH matching with $PATH
	if !e.Compare("PATH", "/usr/bin:$PATH") {
		t.Error("Compare() should handle $PATH substitution")
	}
}

func TestCompareWithDecrypt(t *testing.T) {
	e := NewEnv()
	e.Set("TEST_KEY=test_value")

	// Test without decrypt function (should work like Compare)
	if !e.CompareWithDecrypt("TEST_KEY", "test_value", nil) {
		t.Error("CompareWithDecrypt() should work without decrypt function")
	}

	// Test with decrypt function for non-encrypted value
	decryptFunc := func(s string) (string, error) {
		return strings.TrimPrefix(s, "SOPS:"), nil
	}
	if !e.CompareWithDecrypt("TEST_KEY", "test_value", decryptFunc) {
		t.Error("CompareWithDecrypt() should work with decrypt function for non-encrypted value")
	}

	// Test with decrypt function for encrypted value
	if !e.CompareWithDecrypt("TEST_KEY", "SOPS:test_value", decryptFunc) {
		t.Error("CompareWithDecrypt() should decrypt SOPS: prefixed values")
	}

	// Test with decrypt function that fails
	failingDecryptFunc := func(s string) (string, error) {
		return "", fmt.Errorf("decryption failed")
	}
	if e.CompareWithDecrypt("TEST_KEY", "SOPS:test_value", failingDecryptFunc) {
		t.Error("CompareWithDecrypt() should return false when decryption fails")
	}
}
