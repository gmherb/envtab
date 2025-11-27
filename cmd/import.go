/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:     "import",
	Short:   "Import environment variables from a .env file",
	Long:    `Import environment variables from a .env file into an envtab loadout.`,
	Example: `  envtab import myloadout .env`,
	Args:    cobra.ExactArgs(2),
	Aliases: []string{"i", "im", "imp", "import"},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("import called")
		loadoutName := args[0]
		dotenvFile := args[1]
		lo, err := backends.ReadLoadout(loadoutName)
		if err != nil {
			logger.Error("failure reading loadout", "loadout", loadoutName, "error", err)
			os.Exit(1)
		}
		err = backends.ImportFromDotenv(lo, dotenvFile)
		if err != nil {
			logger.Error("failure importing from dotenv file", "file", dotenvFile, "error", err)
			os.Exit(1)
		}
		err = backends.WriteLoadout(loadoutName, lo)
		if err != nil {
			logger.Error("failure writing loadout", "loadout", loadoutName, "error", err)
			os.Exit(1)
		}
		fmt.Printf("Imported environment variables from [%s] into loadout [%s]\n", dotenvFile, loadoutName)
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
