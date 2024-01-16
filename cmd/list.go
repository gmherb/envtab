/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/gmherb/envtab/cmd/envtab"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all envtab loadouts",
	Long: `List all envtab loadouts.  If the --long flag is provided, then
print the long listing format which includes the loadout name, tags, and other
metadata.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DEBUG: list called")

		if cmd.Flag("long").Value.String() == "true" {
			fmt.Println("DEBUG: long listing format")
			envtab.ListEnvtabLoadouts()

		} else {
			fmt.Println("DEBUG: short listing format")
			envtab.PrintEnvtabLoadouts()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.PersistentFlags().BoolP("long", "l", false, "Print long listing format")
}
