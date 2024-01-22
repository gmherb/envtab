package envtab

import (
	"fmt"
	"os"
	"os/user"
	"strings"
)

func detectLoginScript() string {
	shell := os.Getenv("SHELL")

	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting user's home directory: %s\n", err)
		os.Exit(1)
	}

	switch shell {
	case "/bin/bash":
		if _, err := os.Stat(usr.HomeDir + "/.bash_profile"); err == nil {
			return usr.HomeDir + "/.bash_profile"
		} else if _, err := os.Stat(usr.HomeDir + "/.bash_profile"); err == nil {
			return usr.HomeDir + "/.bash_login"
		} else {
			return usr.HomeDir + "/.profile"
		}
	case "/bin/zsh":
		return usr.HomeDir + "/.zprofile"
	case "/bin/tcsh":
		return usr.HomeDir + "/.login"
	case "/bin/csh":
		return usr.HomeDir + "/.login"
	default:
		return usr.HomeDir + "/.profile"
	}
}

func EnableLogin() {
	loginScript := detectLoginScript()
	println("DEBUG: Detected login script: " + loginScript)
	envtabLogin := "$(envtab login)"

	content, err := os.ReadFile(loginScript)
	if err != nil {
		fmt.Printf("Error reading login script %s: %s\n", loginScript, err)
		os.Exit(1)
	}

	if strings.Contains(string(content), envtabLogin) {
		fmt.Printf("Login script %s already contains %s\n", loginScript, envtabLogin)
		os.Exit(0)
	}

	f, err := os.OpenFile(loginScript, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("Error opening login script %s: %s\n", loginScript, err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err = f.WriteString("\n" + envtabLogin); err != nil {
		fmt.Printf("Error writing to login script %s: %s\n", loginScript, err)
		os.Exit(1)
	}

}

func DisableLogin() {
	loginScript := detectLoginScript()
	println("DEBUG: Detected login script: " + loginScript)
	envtabLogin := "$(envtab login)"

	content, err := os.ReadFile(loginScript)
	if err != nil {
		fmt.Printf("Error reading login script %s: %s\n", loginScript, err)
		os.Exit(1)
	}

	if !strings.Contains(string(content), envtabLogin) {
		fmt.Printf("Login script %s does not contain %s\n", loginScript, envtabLogin)
		os.Exit(0)
	}

	// iterate over the lines, looking for `envtabLogin`
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, envtabLogin) {
			// remove the line
			lines[i] = lines[len(lines)-1]
			lines[len(lines)-1] = ""
			lines = lines[:len(lines)-1]

		}
	}
	output := strings.Join(lines, "\n")

	// Overwrite the login script with the updated content
	f, err := os.OpenFile(loginScript, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error opening login script %s: %s\n", loginScript, err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err = f.WriteString(output); err != nil {
		fmt.Printf("Error writing to login script %s: %s\n", loginScript, err)
		os.Exit(1)
	}

}
