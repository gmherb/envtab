package env

import (
	"log/slog"
	"os"
	"strings"
)

// DecryptFunc is a function type for decrypting values
// This allows the env package to work with encrypted values without
// directly depending on crypto implementations
type DecryptFunc func(string) (string, error)

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

func (e *Env) Compare(key string, value string) bool {
	return e.CompareWithDecrypt(key, value, nil)
}

// CompareWithDecrypt compares a key-value pair, optionally decrypting the value first
// If decryptFunc is provided and value is encrypted (starts with "SOPS:"), it will decrypt first
func (e *Env) CompareWithDecrypt(key string, value string, decryptFunc DecryptFunc) bool {
	match := false

	// Decrypt if needed and decryptFunc is provided
	if decryptFunc != nil {
		if strings.HasPrefix(value, "SOPS:") {
			decrypted, err := decryptFunc(value)
			if err != nil {
				// If decryption fails, can't match - return false
				return false
			}
			value = decrypted
		}
	}

	if strings.Contains(value, "$PATH") {
		slog.Debug("entry contains $PATH", "key", key, "value", value, "env", e.Env)
		value = strings.Replace(value, "$PATH", "", 1)
		value = strings.Trim(value, ":")
	}

	for k, v := range e.Env {
		if k == key && v == value {
			match = true
			break
		} else if key == "PATH" && strings.Contains(v, value) {
			match = true
			break
		}
	}
	return match
}
