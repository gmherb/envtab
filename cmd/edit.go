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

var editCmd = &cobra.Command{
	Use:   "edit [loadout]",
	Short: "Edit envtab loadout",
	Long:  `Enter configured user editor to manually edit a envtab loadout.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DEBUG: edit called")

		if len(args) > 1 {
			fmt.Printf("ERROR: edit only accepts one loadout name\n")
			os.Exit(1)
		}

		if len(args) == 0 {
			fmt.Printf("ERROR: edit requires a loadout name\n")
			os.Exit(1)
		}

		envtabPath := envtab.InitEnvtab("")

		loadoutName := args[0]
		loadoutPath := envtabPath + `/` + loadoutName + `.yaml`

		println("DEBUG: loadoutPath:" + loadoutPath + ", loadoutName: " + loadoutPath)

		if _, err := os.Stat(loadoutPath); os.IsNotExist(err) {
			fmt.Printf("ERROR: Loadout [%s] does not exist\n", loadoutName)
			os.Exit(1)
		}

		err := envtab.EditLoadout(loadoutName)
		if err != nil {
			fmt.Printf("ERROR: Failure editing loadout [%s]: %s\n", loadoutName, err)
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
