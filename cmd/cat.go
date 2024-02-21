/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/gmherb/envtab/cmd/envtab"
	"github.com/spf13/cobra"
)

var catCmd = &cobra.Command{
	Use:   "cat LOADOUT_NAME [LOADOUT_NAME ...]",
	Short: "Concatenate envtab loadouts to stdout",
	Long:  `Concatenate envtab loadouts to stdout.`,
	Example: `  envtab cat myloadout
  envtab cat myloadout1 myloadout2 myloadout3`,
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"c", "ca", "print"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DEBUG: cat called")

		for _, arg := range args {

			loadout, err := envtab.ReadLoadout(arg)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				os.Exit(1)
			}

			loadout.PrintLoadout()
		}
	},
}

func init() {
	rootCmd.AddCommand(catCmd)
}
