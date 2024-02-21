/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/gmherb/envtab/cmd/envtab"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:     "rm LOADOUT_NAME [LOADOUT_NAME ...]",
	Short:   "Remove envtab loadout(s)",
	Long:    `Remove envtab loadout(s)`,
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"r", "remove", "delete", "del"},
	Run: func(cmd *cobra.Command, args []string) {
		println("DEBUG: rm called")
		for _, loadout := range args {
			println("DEBUG: removing loadout " + loadout)
			envtab.RemoveLoadout(loadout)
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
