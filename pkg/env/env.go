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

func (e *Env) Print() {
	for k, v := range e.Env {
		println(k, ":", v)
	}
}

func (e *Env) Get(key string) string {
	return e.Env[key]
}
