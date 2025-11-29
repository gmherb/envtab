/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"
	"net/url"
	"io"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/loadout"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import environment variables or loadouts",
	Long:  `Import environment variables from a .env file (merge) or a full loadout from YAML (.yaml|.yml). Supports local files and HTTP(S) URLs.`,
	Example: `  # Local .env into existing loadout (merge)
  envtab import myloadout ./config.env

  # Local YAML loadout (replace/create)
  envtab import myloadout ./prod.yaml

  # Remote .env (merge)
  envtab import myloadout --url https://raw.githubusercontent.com/org/repo/branch/config.env

  # Remote YAML loadout (replace/create)
  envtab import myloadout --url https://raw.githubusercontent.com/org/repo/branch/loadouts/prod.yaml`,
	Args: func(cmd *cobra.Command, args []string) error {
		if importURL != "" {
			if len(args) != 1 {
				return fmt.Errorf("when using --url, provide exactly one argument: LOADOUT_NAME")
			}
			return nil
		}
		if len(args) != 2 {
			return fmt.Errorf("expected 2 arguments: LOADOUT_NAME and FILE_PATH")
		}
		return nil
	},
	Aliases: []string{"i", "im", "imp", "import"},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("import called")
		loadoutName := args[0]

		// URL-based import path
		if importURL != "" {
			if err := importFromURL(loadoutName, importURL); err != nil {
				logger.Error("failure importing from URL", "url", importURL, "error", err)
				os.Exit(1)
			}
			return
		}

		// Local file import path (existing behavior for .env, extended to support .yaml|.yml)
		inputPath := args[1]
		ext := path.Ext(inputPath)
		switch ext {
		case ".env":
			lo, err := backends.ReadLoadout(loadoutName)
			if err != nil && !os.IsNotExist(err) {
				logger.Error("failure reading loadout", "loadout", loadoutName, "error", err)
				os.Exit(1)
			}
			if os.IsNotExist(err) {
				lo = loadout.InitLoadout()
			}
			if err := backends.ImportFromDotenv(lo, inputPath); err != nil {
				logger.Error("failure importing from dotenv file", "file", inputPath, "error", err)
				os.Exit(1)
			}
			if err := backends.WriteLoadout(loadoutName, lo); err != nil {
				logger.Error("failure writing loadout", "loadout", loadoutName, "error", err)
				os.Exit(1)
			}
			fmt.Printf("Imported environment variables from [%s] into loadout [%s]\n", inputPath, loadoutName)
		case ".yaml", ".yml":
			data, err := os.ReadFile(inputPath)
			if err != nil {
				logger.Error("failure reading YAML file", "file", inputPath, "error", err)
				os.Exit(1)
			}
			if err := loadout.ValidateLoadoutYAML(data); err != nil {
				logger.Error("invalid loadout YAML", "file", inputPath, "error", err)
				os.Exit(1)
			}
			var lo loadout.Loadout
			if err := yaml.Unmarshal(data, &lo); err != nil {
				logger.Error("failure parsing loadout YAML", "file", inputPath, "error", err)
				os.Exit(1)
			}
			if err := backends.WriteLoadout(loadoutName, &lo); err != nil {
				logger.Error("failure writing loadout", "loadout", loadoutName, "error", err)
				os.Exit(1)
			}
			fmt.Printf("Imported loadout YAML from [%s] into loadout [%s]\n", inputPath, loadoutName)
		default:
			logger.Error("unsupported file extension; expected .env, .yaml, or .yml", "file", inputPath)
			os.Exit(1)
		}
	},
}

var importURL string

func importFromURL(loadoutName string, rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("only http and https URLs are supported")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	ext := path.Ext(u.Path)
	switch ext {
	case ".env":
		lo, err := backends.ReadLoadout(loadoutName)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed reading loadout: %w", err)
		}
		if os.IsNotExist(err) {
			lo = loadout.InitLoadout()
		}
		entries, err := backends.ParseDotenvContent(data)
		if err != nil {
			return fmt.Errorf("failed parsing dotenv content: %w", err)
		}
		for k, v := range entries {
			lo.UpdateEntry(k, v)
		}
		if err := backends.WriteLoadout(loadoutName, lo); err != nil {
			return fmt.Errorf("failed writing loadout: %w", err)
		}
		fmt.Printf("Imported environment variables from URL [%s] into loadout [%s]\n", rawURL, loadoutName)
		return nil
	case ".yaml", ".yml":
		if err := loadout.ValidateLoadoutYAML(data); err != nil {
			return fmt.Errorf("invalid loadout YAML: %w", err)
		}
		var lo loadout.Loadout
		if err := yaml.Unmarshal(data, &lo); err != nil {
			return fmt.Errorf("failed parsing loadout YAML: %w", err)
		}
		if err := backends.WriteLoadout(loadoutName, &lo); err != nil {
			return fmt.Errorf("failed writing loadout: %w", err)
		}
		fmt.Printf("Imported loadout YAML from URL [%s] into loadout [%s]\n", rawURL, loadoutName)
		return nil
	default:
		return fmt.Errorf("unsupported URL extension; expected .env, .yaml, or .yml")
	}
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVarP(&importURL, "url", "u", "", "Import from HTTP(S) URL (.env merges, .yaml|.yml replaces/creates)")
}
