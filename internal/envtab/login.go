package envtab

import (
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"strings"
)

var loginScripts = []string{
	".bash_profile",
	".bash_login",
	".profile",
	".zprofile",
	".login",
}
var envtabLoginLine = "$(envtab login)"

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
		} else if _, err := os.Stat(usr.HomeDir + "/.bash_login"); err == nil {
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
	slog.Debug("detected login script", "script", loginScript)
	envtabLogin := "$(envtab login)"

	content, err := os.ReadFile(loginScript)
	if err != nil {
		fmt.Printf("Error reading login script %s: %s\n", loginScript, err)
		os.Exit(1)
	}

	if strings.Contains(string(content), envtabLogin) {
		fmt.Printf("DEBUG: Login script %s already contains %s\n", loginScript, envtabLogin)
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
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting user's home directory: %s\n", err)
		os.Exit(1)
	}
	for _, loginScript := range loginScripts {
		removeEnvtabFromScript(usr.HomeDir + "/" + loginScript)
	}
}

func removeEnvtabFromScript(loginScript string) {
	fmt.Printf("DEBUG: Removing envtab from login script [%s]\n", loginScript)
	content, err := os.ReadFile(loginScript)

	// ignore error if file doesn't exist
	if os.IsNotExist(err) {
		fmt.Printf("DEBUG: Login script [%s] does not exist\n", loginScript)
		return
	} else if err != nil {
		fmt.Printf("Error reading login script [%s]: %s\n", loginScript, err)
		os.Exit(1)
	}

	// ignore if login script doesn't contain `envtabLoginLine`
	if !strings.Contains(string(content), envtabLoginLine) {
		fmt.Printf("DEBUG: Login script [%s] does not contain [%s]\n", loginScript, envtabLoginLine)
		return
	}
	fmt.Printf("DEBUG: Login script [%s] contains [%s]\n", loginScript, envtabLoginLine)
	// iterate over the lines, looking for `envtabLoginLine`
	lines := strings.Split(string(content), "\n")
	newlines := []string{}
	for i, line := range lines {
		if !strings.Contains(line, envtabLoginLine) {
			newlines = append(newlines, line)
			fmt.Printf("DEBUG: Keeping [%s] from line %d\n", line, i)
		} else {
			fmt.Printf("DEBUG: Removing [%s] from line %d\n", envtabLoginLine, i)
		}
	}
	output := strings.Join(newlines, "\n")

	// Overwrite the login script with the updated content
	f, err := os.OpenFile(loginScript, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Error opening login script [%s]: %s\n", loginScript, err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err = f.WriteString(output); err != nil {
		fmt.Printf("Error writing to login script [%s]: %s\n", loginScript, err)
		os.Exit(1)
	}

}

func ShowLoginStatus() {
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting user's home directory: %s\n", err)
		os.Exit(1)
	}
	var loginScriptPath string
	for _, loginScript := range loginScripts {
		loginScriptPath = usr.HomeDir + "/" + loginScript

		fmt.Printf("DEBUG: Checking login script [%s] for envtab\n", loginScript)
		content, err := os.ReadFile(loginScriptPath)

		// ignore error if file doesn't exist
		if os.IsNotExist(err) {
			fmt.Printf("DEBUG: Login script [%s] does not exist\n", loginScript)
		} else if err != nil {
			fmt.Printf("Error reading login script [%s]: %s\n", loginScript, err)
			os.Exit(1)
		}

		// Print enabled if the login script contains `envtabLoginLine`
		if strings.Contains(string(content), envtabLoginLine) {
			fmt.Printf("DEBUG: Login script [%s] contains [%s]\n", loginScript, envtabLoginLine)
			fmt.Printf("enabled\n")
			return

		}
	}
	fmt.Printf("disabled\n")
	return
}
