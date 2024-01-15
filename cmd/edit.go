/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [loadout]",
	Short: "Edit a envtab loadout",
	Long: `Enter configured user editor to manually edit a envtab loadout.

If no loadout is specified, the active loadout with the most matching entries
will be edited.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DEBUG: edit called")

	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
