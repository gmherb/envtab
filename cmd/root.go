/*
Copyright © 2024 Greg Herbster
*/
package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gmherb/envtab/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cfgFile is the path to the config file (provided by flag)
var cfgFile string

const (
	ENVTAB_DIR         = ".envtab"
	ENVTAB_CONFIG      = ".envtab"
	ENVTAB_CONFIG_TYPE = "yaml"
)

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

func init() {
	// Initialize default logger with ERROR level
	// Log level will be updated in initConfig() after viper is configured
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	})))

	// Initialize Viper configuration (cobra will call this automatically)
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and environment variables if set.
// Priority order for config file:
// 1. Command-line flag (--config)
// 2. Environment variable (ENVTAB_CONFIG)
// 3. Project config: CWD/.envtab.yaml and walk up the directory tree
// 4. User config: ~/.envtab.yaml or $XDG_CONFIG_HOME/envtab/.envtab.yaml
// 5. System config: /etc/envtab.yaml
func initConfig() {
	if cfgFile != "" {
		// 1. Command-line flag (--config)
		viper.SetConfigFile(cfgFile)
	} else if envConfig := os.Getenv("ENVTAB_CONFIG"); envConfig != "" {
		// 2. Environment variable (ENVTAB_CONFIG)
		viper.SetConfigFile(envConfig)
	} else {
		// Use hierarchical config paths
		viper.SetConfigName(ENVTAB_CONFIG)
		viper.SetConfigType(ENVTAB_CONFIG_TYPE)

		// 3. Project config: walk up from CWD to find .envtab.yaml
		projectConfig := config.FindProjectConfig()
		if projectConfig != "" {
			projectDir := filepath.Dir(projectConfig)
			viper.AddConfigPath(projectDir)
			slog.Debug("Found project config", "path", projectConfig)
		}

		// 4. User config: ~/.envtab.yaml or $XDG_CONFIG_HOME/envtab/.envtab.yaml
		userConfigPath := config.GetUserConfigPath()
		userConfigDir := filepath.Dir(userConfigPath)
		viper.AddConfigPath(userConfigDir)

		// 5. System config: /etc/envtab.yaml
		viper.AddConfigPath("/etc")
	}

	viper.SetEnvPrefix("ENVTAB")
	viper.AutomaticEnv()
	viper.SetDefault("log.level", "ERROR")

	// Read config file if found
	if err := viper.ReadInConfig(); err == nil {
		slog.Debug("Using config file", "file", viper.ConfigFileUsed())
	}

	// Always set logger from viper (reads from env var, config file, or default)
	logLevelStr := viper.GetString("log.level")
	logLevel := parseLogLevelFromString(logLevelStr)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})))
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
		// --verbose flag overrides config file setting
		if verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose"); verbose {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		}
		// Otherwise, keep the level set in initConfig() (from config file or env var)
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
	// Define persistent flags (global for all commands)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (overrides project/user/system config precedence)")
	rootCmd.PersistentFlags().BoolP("verbose", "", false, "Show verbose output (enables debug/info/warn logs)")
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
