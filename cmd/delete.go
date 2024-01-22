/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/gmherb/envtab/cmd/envtab"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <loadout> [loadout]...",
	Short: "Delete envtab loadout(s)",
	Long:  `Delete envtab loadout(s)`,
	Run: func(cmd *cobra.Command, args []string) {
		println("DEBUG: delete called")

		if len(args) < 1 {
			println("ERROR: Must specify loadout(s) to delete")
			os.Exit(1)
		}

		for _, loadout := range args {
			println("DEBUG: deleting loadout " + loadout)
			envtab.DeleteLoadout(loadout)
		}

	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
