package env

import (
	"os"
	"strings"
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

func (e *Env) Compare(key string, value string) bool {
	match := false

	if strings.Contains(value, "$PATH") {
		//println("DEBUG: entry contains $PATH")
		value = strings.Replace(value, "$PATH", "", 1)
		value = strings.Trim(value, ":")
		//println("DEBUG: trying to match:", value)
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
