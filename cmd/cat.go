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
	Use:   "cat <name>",
	Short: "Print an envtab loadout",
	Long:  `Print an envtab loadout`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DEBUG: cat called")

		loadout, err := envtab.ReadLoadout(args[0])
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}

		loadout.PrintLoadout()
	},
}

func init() {
	rootCmd.AddCommand(catCmd)
}
