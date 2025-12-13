package env

import (
	"os"
	"strings"

	"github.com/gmherb/envtab/internal/loadout"
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

	// Expand environment variables in the value
	// Only expand if not encrypted (encrypted values will have "SOPS:" prefix)
	if !strings.HasPrefix(displayValue, "SOPS:") {
		displayValue = loadout.ExpandVariables(displayValue)
	}

	// Handle PATH specially - check if the expanded value is contained in current PATH
	if key == "PATH" {
		currentPath := e.Get("PATH")
		if currentPath != "" {
			// For PATH, check if all segments in displayValue are in currentPath
			pathSegments := strings.Split(displayValue, ":")
			allInPath := true
			for _, segment := range pathSegments {
				if segment != "" && !strings.Contains(currentPath, segment) {
					allInPath = false
					break
				}
			}
			if allInPath && len(pathSegments) > 0 {
				match = true
			}
		}
	} else {
		// For non-PATH variables, do exact match
		for k, v := range e.Env {
			if k == key && v == displayValue {
				match = true
				break
			}
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

	// Expand environment variables in the value
	// Only expand if not encrypted (encrypted values will have "SOPS:" prefix)
	if !strings.HasPrefix(displayValue, "SOPS:") {
		displayValue = loadout.ExpandVariables(displayValue)
	}

	// Handle PATH specially - check if the expanded value is contained in current PATH
	if key == "PATH" {
		currentPath := e.Get("PATH")
		if currentPath != "" {
			// For PATH, check if all segments in displayValue are in currentPath
			pathSegments := strings.Split(displayValue, ":")
			allInPath := true
			for _, segment := range pathSegments {
				if segment != "" && !strings.Contains(currentPath, segment) {
					allInPath = false
					break
				}
			}
			if allInPath && len(pathSegments) > 0 {
				match = true
			}
		}
	} else {
		// For non-PATH variables, do exact match
		for k, v := range e.Env {
			if k == key && v == displayValue {
				match = true
				break
			}
		}
	}
	return match
}
