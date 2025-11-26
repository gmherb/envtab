/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/gmherb/envtab/internal/envtab"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export LOADOUT_NAME [LOADOUT_NAME ...]",
	Short: "Export envtab loadout(s)",
	Long: `Print export statements for provided loadouts to be sourced into
	your environment.`,
	Example: `  $(envtab export myloadout)
  $(envtab export myloadout1 myloadout2 myloadout3)`,
	Args:                  cobra.MinimumNArgs(1),
	DisableFlagsInUseLine: true,
	SuggestFor:            []string{"load", "source", "."},
	Aliases:               []string{"ex", "exp", "expo"},
	Run: func(cmd *cobra.Command, args []string) {
		println("DEBUG: export called")

		envtabPath := envtab.InitEnvtab("")

		for _, arg := range args {

			loadoutName := arg
			loadoutPath := envtabPath + `/` + loadoutName + `.yaml`

			println("DEBUG: loadoutName: " + loadoutName + ", loadoutPath: " + loadoutPath)

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
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
