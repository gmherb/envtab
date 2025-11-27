/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/gmherb/envtab/internal/backends"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "remove LOADOUT_NAME [LOADOUT_NAME ...]",
	Short: "Remove envtab loadout(s)",
	Long:  `Remove envtab loadout(s)`,
	Example: `  envtab remove myloadout
  envtab remove myloadout1 myloadout2 myloadout3`,
	Args:       cobra.MinimumNArgs(1),
	SuggestFor: []string{"delete", "del"},
	Aliases:    []string{"r", "rm"},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("rm called")
		for _, loadout := range args {
			logger.Debug("removing loadout", "loadout", loadout)
			backends.RemoveLoadout(loadout)
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
