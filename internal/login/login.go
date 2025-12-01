package login

import (
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

var loginScripts = []string{
	".bash_profile",
	".bash_login",
	".profile",
	".zprofile",
	".login",
}

// getEnvtabLoginLine returns the login line with the absolute path to the binary
func getEnvtabLoginLine() string {
	execPath, err := os.Executable()
	if err != nil {
		slog.Error("failure getting executable path", "error", err)
		os.Exit(1)
	}
	// os.Executable() may return a relative path on some systems, so ensure it's absolute
	absPath, err := filepath.Abs(execPath)
	if err != nil {
		slog.Error("failure getting absolute path", "error", err)
		os.Exit(1)
	}
	return fmt.Sprintf("$(%s login)", absPath)
}

func detectLoginScript() string {
	shell := os.Getenv("SHELL")

	usr, err := user.Current()
	if err != nil {
		slog.Error("failure getting user's home directory", "error", err)
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
	envtabLogin := getEnvtabLoginLine()

	content, err := os.ReadFile(loginScript)
	if err != nil {
		slog.Error("failure reading login script", "script", loginScript, "error", err)
		os.Exit(1)
	}

	if strings.Contains(string(content), envtabLogin) {
		slog.Debug("login script already contains envtab", "script", loginScript)
		os.Exit(0)
	}

	f, err := os.OpenFile(loginScript, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		slog.Error("failure opening login script", "script", loginScript, "error", err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err = f.WriteString("\n" + envtabLogin); err != nil {
		slog.Error("failure writing to login script", "script", loginScript, "error", err)
		os.Exit(1)
	}

}

func DisableLogin() {
	usr, err := user.Current()
	if err != nil {
		slog.Error("failure getting user's home directory", "error", err)
		os.Exit(1)
	}
	for _, loginScript := range loginScripts {
		removeEnvtabFromScript(usr.HomeDir + "/" + loginScript)
	}
}

func removeEnvtabFromScript(loginScript string) {
	slog.Debug("removing envtab from login script", "script", loginScript)
	content, err := os.ReadFile(loginScript)

	// ignore error if file doesn't exist
	if os.IsNotExist(err) {
		slog.Debug("login script does not exist", "script", loginScript)
		return
	} else if err != nil {
		slog.Error("failure reading login script", "script", loginScript, "error", err)
		os.Exit(1)
	}

	// Get file info to preserve permissions
	fileInfo, err := os.Stat(loginScript)
	if err != nil {
		slog.Error("failure getting file info", "script", loginScript, "error", err)
		os.Exit(1)
	}

	envtabLoginLine := getEnvtabLoginLine()
	// ignore if login script doesn't contain `envtabLoginLine`
	if !strings.Contains(string(content), envtabLoginLine) {
		slog.Debug("login script does not contain envtab", "script", loginScript)
		return
	}
	slog.Debug("login script contains envtab", "script", loginScript)
	// iterate over the lines, looking for `envtabLoginLine`
	lines := strings.Split(string(content), "\n")
	newlines := []string{}
	for i, line := range lines {
		if !strings.Contains(line, envtabLoginLine) {
			newlines = append(newlines, line)
			slog.Debug("keeping line from login script", "script", loginScript, "line", i, "content", line)
		} else {
			slog.Debug("removing line from login script", "script", loginScript, "line", i)
		}
	}
	output := strings.Join(newlines, "\n")

	// Overwrite the login script with the updated content, preserving permissions
	f, err := os.OpenFile(loginScript, os.O_WRONLY|os.O_TRUNC, fileInfo.Mode())
	if err != nil {
		slog.Error("failure opening login script", "script", loginScript, "error", err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err = f.WriteString(output); err != nil {
		slog.Error("failure writing to login script", "script", loginScript, "error", err)
		os.Exit(1)
	}

}

func ShowLoginStatus() {
	usr, err := user.Current()
	if err != nil {
		slog.Error("failure getting user's home directory", "error", err)
		os.Exit(1)
	}
	var loginScriptPath string
	for _, loginScript := range loginScripts {
		loginScriptPath = usr.HomeDir + "/" + loginScript

		slog.Debug("checking login script for envtab", "script", loginScript)
		content, err := os.ReadFile(loginScriptPath)

		// ignore error if file doesn't exist
		if os.IsNotExist(err) {
			slog.Debug("login script does not exist", "script", loginScript)
			continue
		} else if err != nil {
			slog.Error("failure reading login script", "script", loginScript, "error", err)
			os.Exit(1)
		}

		envtabLoginLine := getEnvtabLoginLine()
		// Print enabled if the login script contains `envtabLoginLine`
		if strings.Contains(string(content), envtabLoginLine) {
			slog.Debug("login script contains envtab", "script", loginScript)
			fmt.Printf("enabled\n")
			return

		}
	}
	fmt.Printf("disabled\n")
}
