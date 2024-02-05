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

var exportCmd = &cobra.Command{
	Use:   "export <loadout>",
	Short: "Export envtab loadout",
	Long: `Print export statements for provided loadout to be sourced into your
environment.

Example: $(envtab export myloadout)`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		println("DEBUG: export called")

		envtabPath := envtab.InitEnvtab("")

		loadoutName := args[0]
		loadoutPath := envtabPath + `/` + loadoutName + `.yaml`

		println("DEBUG: loadoutPath:" + loadoutPath + ", loadoutName: " + loadoutPath)

		if _, err := os.Stat(loadoutPath); os.IsNotExist(err) {
			fmt.Printf("ERROR: Loadout [%s] does not exist\n", loadoutName)
			os.Exit(1)
		}

		loadout, err := envtab.ReadLoadout(loadoutName)
		if err != nil {
			fmt.Printf("ERROR: Failure reading loadout [%s]: %s\n", loadoutName, err)
			os.Exit(1)
		}

		loadout.Export()
		envtab.WriteLoadout(loadoutName, loadout)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
