// Test file for env.go
package env

import (
	"testing"
)

func TestParseKeyValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expectedKey   string
		expectedValue string
	}{
		{"simple key-value", "KEY=value", "KEY", "value"},
		{"value with equals", "KEY=value=with=equals", "KEY", "value=with=equals"},
		{"empty value", "KEY=", "KEY", ""},
		{"no equals sign", "KEY", "", ""},
		{"multiple equals", "KEY=val1=val2=val3", "KEY", "val1=val2=val3"},
		{"key with leading whitespace", "  KEY=value", "KEY", "value"},
		{"key with trailing whitespace", "KEY  =value", "KEY", "value"},
		{"key with both whitespace", "  KEY  =value", "KEY", "value"},
		{"value with leading whitespace", "KEY=  value", "KEY", "value"},
		{"value with trailing whitespace", "KEY=value  ", "KEY", "value"},
		{"value with both whitespace", "KEY=  value  ", "KEY", "value"},
		{"both key and value with whitespace", "  KEY  =  value  ", "KEY", "value"},
		{"value with internal whitespace", "KEY=value with spaces", "KEY", "value with spaces"},
		{"empty value with whitespace", "KEY=  ", "KEY", ""},
		{"key with whitespace, empty value", "  KEY  =", "KEY", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value := ParseKeyValue(tt.input)
			if key != tt.expectedKey {
				t.Errorf("ParseKeyValue() key = %v, want %v", key, tt.expectedKey)
			}
			if value != tt.expectedValue {
				t.Errorf("ParseKeyValue() value = %v, want %v", value, tt.expectedValue)
			}
		})
	}
}

func TestSet(t *testing.T) {
	e := NewEnv()
	e.Set("TEST=TEST")
	if e.Env["TEST"] != "TEST" {
		t.Errorf("Expected %s, got %s", "TEST", e.Env["TEST"])
	}

	// Test value with equals sign
	e.Set("TEST2=value=with=equals")
	if e.Env["TEST2"] != "value=with=equals" {
		t.Errorf("Expected %s, got %s", "value=with=equals", e.Env["TEST2"])
	}
}

func TestPopulate(t *testing.T) {
	e := NewEnv()
	e.Populate()
	if e.Env["PATH"] == "" {
		t.Errorf("Expected %s, got %s", "PATH", e.Env["PATH"])
	}
	if e.Env["HOME"] == "" {
		t.Errorf("Expected %s, got %s", "HOME", e.Env["HOME"])
	}
	if e.Env["USER"] == "" {
		t.Errorf("Expected %s, got %s", "USER", e.Env["USER"])
	}
	if e.Env["PWD"] == "" {
		t.Errorf("Expected %s, got %s", "PWD", e.Env["PWD"])
	}
	if e.Env["SHLVL"] == "" {
		t.Errorf("Expected %s, got %s", "SHLVL", e.Env["SHLVL"])
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

func TestCompareRawValue(t *testing.T) {
	e := NewEnv()
	e.Set("TEST_KEY=test_value")

	// Test exact match
	if !e.CompareRawValue("TEST_KEY", "test_value") {
		t.Error("CompareRawValue() should return true for exact match")
	}

	// Test non-match
	if e.CompareRawValue("TEST_KEY", "wrong_value") {
		t.Error("CompareRawValue() should return false for non-match")
	}

	// Test non-existent key
	if e.CompareRawValue("NON_EXISTENT", "value") {
		t.Error("CompareRawValue() should return false for non-existent key")
	}
}

func TestCompareSOPSEncryptedValue(t *testing.T) {
	e := NewEnv()
	e.Set("TEST_KEY=test_value")
	e.Set("PATH=/usr/bin:/usr/local/bin")

	// Test exact match with non-SOPS value
	if !e.CompareSOPSEncryptedValue("TEST_KEY", "test_value") {
		t.Error("CompareSOPSEncryptedValue() should return true for exact match")
	}

	// Test non-match
	if e.CompareSOPSEncryptedValue("TEST_KEY", "wrong_value") {
		t.Error("CompareSOPSEncryptedValue() should return false for non-match")
	}

	// Test non-existent key
	if e.CompareSOPSEncryptedValue("NON_EXISTENT", "value") {
		t.Error("CompareSOPSEncryptedValue() should return false for non-existent key")
	}
}
