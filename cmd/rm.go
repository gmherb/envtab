/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/gmherb/envtab/internal/envtab"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm LOADOUT_NAME [LOADOUT_NAME ...]",
	Short: "Remove envtab loadout(s)",
	Long:  `Remove envtab loadout(s)`,
	Example: `  envtab rm myloadout
  envtab rm myloadout1 myloadout2 myloadout3`,
	Args:       cobra.MinimumNArgs(1),
	SuggestFor: []string{"delete", "del"},
	Aliases:    []string{"r", "remove"},
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
