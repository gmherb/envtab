/*
Copyright © 2024 Greg Herbster
*/
package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// logger is a structured logger that writes to stderr
// This allows separating debug logs from normal output (which goes to stdout)
var logger *slog.Logger

// parseLogLevel parses the log level from an environment variable
// Supported values: DEBUG, INFO, WARN, ERROR (case-insensitive)
// Defaults to INFO if not set or invalid
func parseLogLevel() slog.Level {
	levelStr := strings.ToUpper(os.Getenv("ENVTAB_LOG_LEVEL"))
	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		// Default to INFO if not set or invalid
		return slog.LevelInfo
	}
}

func init() {
	// Initialize logger to write to stderr
	// This ensures debug logs don't interfere with normal command output
	// Log level can be configured via ENVTAB_LOG_LEVEL environment variable
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: parseLogLevel(),
	}))
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "envtab",
	Short:   "Take control of your environment.",
	Long:    `Take control of your environment.`,
	Version: "0.0.0",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.envtab.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
