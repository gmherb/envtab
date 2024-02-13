// Test file for env.go
package env

import (
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
