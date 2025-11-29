/*
Copyright © 2024 Greg Herbster
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// logger is a structured logger that writes to stderr
// This allows separating debug logs from normal output (which goes to stdout)
var logger *slog.Logger

// cfgFile is the path to the config file
var cfgFile string

// parseLogLevel parses the log level from an environment variable or config
// Supported values: DEBUG, INFO, WARN, ERROR (case-insensitive)
// Defaults to INFO if not set or invalid
func parseLogLevel() slog.Level {
	// Check config first, then environment variable
	levelStr := viper.GetString("log.level")
	if levelStr == "" {
		levelStr = strings.ToUpper(os.Getenv("ENVTAB_LOG_LEVEL"))
	}

	levelStr = strings.ToUpper(levelStr)
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
	// Log level can be configured via ENVTAB_LOG_LEVEL environment variable or config
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: parseLogLevel(),
	}))

	// Initialize Viper
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Priority order:
	// 1. Command-line flag (--config)
	// 2. Environment variable (ENVTAB_CONFIG)
	// 3. Default paths

	if cfgFile != "" {
		// Use config file from the flag (highest priority).
		viper.SetConfigFile(cfgFile)
	} else if envConfig := os.Getenv("ENVTAB_CONFIG"); envConfig != "" {
		// Use config file from environment variable
		viper.SetConfigFile(envConfig)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting user's home directory: %s\n", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".envtab" (without extension).
		viper.AddConfigPath(filepath.Join(home, ".envtab"))
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".envtab")
	}

	// Set environment variable prefix
	viper.SetEnvPrefix("ENVTAB")
	viper.AutomaticEnv() // read in environment variables that match

	// Set defaults
	viper.SetDefault("log.level", "INFO")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Debug("Using config file", "file", viper.ConfigFileUsed())
	} else {
		// Config file not found; ignore if missing
		logger.Debug("No config file found, using defaults", "error", err)
	}
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.envtab/.envtab.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// GetRootCmd exposes the root command for tooling (e.g., docs generators).
func GetRootCmd() *cobra.Command {
	return rootCmd
}
