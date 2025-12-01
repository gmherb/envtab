/*
Copyright © 2024 Greg Herbster
*/
package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cfgFile is the path to the config file
var cfgFile string

// Version information set at build time via ldflags
var (
	Version   string
	Commit    string
	BuildDate string
)

// parseLogLevelFromString parses a log level string to slog.Level
// Supported values: DEBUG, INFO, WARN, ERROR (case-insensitive)
// Defaults to ERROR if not set or invalid
func parseLogLevelFromString(levelStr string) slog.Level {
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
		return slog.LevelError
	}
}

// parseLogLevel parses the log level from config or environment variable
// Supported values: DEBUG, INFO, WARN, ERROR (case-insensitive)
// Defaults to ERROR if not set or invalid
func parseLogLevel() slog.Level {
	// Check config first, then environment variable
	levelStr := viper.GetString("log.level")
	if levelStr == "" {
		levelStr = os.Getenv("ENVTAB_LOG_LEVEL")
	}
	return parseLogLevelFromString(levelStr)
}

func init() {
	// Initialize default logger to write to stderr
	// This ensures debug logs don't interfere with normal command output
	// Log level can be configured via ENVTAB_LOG_LEVEL environment variable or config
	// Defaults to ERROR if not set
	levelStr := os.Getenv("ENVTAB_LOG_LEVEL")
	level := parseLogLevelFromString(levelStr)

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))

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
			slog.Error("failure getting user's home directory", "error", err)
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
	viper.SetDefault("log.level", "ERROR")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		slog.Debug("Using config file", "file", viper.ConfigFileUsed())
		// Update default logger with log level from config if it differs
		configLevel := parseLogLevel()
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: configLevel,
		})))
	} else {
		// Config file not found; ignore if missing
		slog.Debug("No config file found, using defaults", "error", err)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "envtab",
	Short:   "Take control of your environment.",
	Long:    `Take control of your environment.`,
	Version: getVersion(),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Check if ENVTAB_LOG_LEVEL is explicitly set - if so, respect it (don't override)
		if explicitLevel := os.Getenv("ENVTAB_LOG_LEVEL"); explicitLevel != "" {
			// ENVTAB_LOG_LEVEL was explicitly set, keep the level from initConfig()
			// (which already processed it)
			return
		}

		// No explicit ENVTAB_LOG_LEVEL, so apply --verbose flag logic
		// Access persistent flag from root command
		verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")
		if verbose {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		} else {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelError,
			})))
		}
	},
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
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Show verbose output (enables debug/info/warn logs)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// GetRootCmd exposes the root command for tooling (e.g., docs generators).
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// getVersion returns a formatted version string
func getVersion() string {
	version := Version
	// Append commit hash if it doesn't already contain it
	if Commit != "" {
		shortCommit := Commit
		if len(shortCommit) > 7 {
			shortCommit = shortCommit[:7]
		}
		// Check if version already contains this commit hash (from git describe)
		if !strings.Contains(version, shortCommit) {
			version += "+" + shortCommit
		}
	}
	// Append build date if available
	if BuildDate != "" {
		version += " (built " + BuildDate + ")"
	}
	return version
}
