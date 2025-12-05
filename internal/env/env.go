package env

import (
	"log/slog"
	"os"
	"strings"

	"github.com/gmherb/envtab/internal/sops"
)

type Env struct {
	Env map[string]string
}

func NewEnv() *Env {
	return &Env{Env: make(map[string]string)}
}

func (e *Env) Set(ev string) {
	pair := strings.Split(ev, "=")
	e.Env[pair[0]] = pair[1]
}

func (e *Env) Populate() {
	for _, ev := range os.Environ() {
		e.Set(ev)
	}
}

func (e *Env) Dump() {
	for k, v := range e.Env {
		println(k, ":", v)
	}
}

func (e *Env) Get(key string) string {
	return e.Env[key]
}

func (e *Env) CompareRawValue(key string, value string) bool {
	return sops.SOPSDisplayValue(e.Get(key), false) == value
}

func (e *Env) CompareSOPSEncryptedValue(key string, value string) bool {
	match := false

	displayValue := sops.SOPSDisplayValue(value, true)

	if strings.Contains(displayValue, "$PATH") {
		slog.Debug("entry contains $PATH", "key", key, "value", value, "env", e.Env)
		displayValue = strings.Replace(displayValue, "$PATH", "", 1)
		displayValue = strings.Trim(displayValue, ":")
	}

	for k, v := range e.Env {
		if k == key && v == displayValue {
			match = true
			break
		} else if key == "PATH" && strings.Contains(v, displayValue) {
			match = true
			break
		}
	}
	return match
}

// IsEntryActive checks if a loadout entry is currently active in the environment
// It compares the entry's value (decrypted if encrypted) with what's in the environment
func (e *Env) IsEntryActive(key string, value string) bool {
	match := false

	// Decrypt value if it's encrypted
	displayValue := sops.SOPSDisplayValue(value, true)

	if strings.Contains(displayValue, "$PATH") {
		slog.Debug("entry contains $PATH", "key", key, "value", value, "env", e.Env)
		displayValue = strings.Replace(displayValue, "$PATH", "", 1)
		displayValue = strings.Trim(displayValue, ":")
	}

	for k, v := range e.Env {
		if k == key && v == displayValue {
			match = true
			break
		} else if key == "PATH" && strings.Contains(v, displayValue) {
			match = true
			break
		}
	}
	return match
}
