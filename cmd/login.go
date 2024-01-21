/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/gmherb/envtab/cmd/envtab"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Export all loadouts with login: true",
	Long: `Export all loadouts which are enabled on login. This is typically
run from a login script such as ~/.bash_profile or ~/.zprofile.`,
	Run: func(cmd *cobra.Command, args []string) {
		println("DEBUG: login called")

		exportLoginLoadouts()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func exportLoginLoadouts() {
	loadouts := envtab.GetEnvtabSlice()

	for _, loadout := range loadouts {

		loadoutStruct, err := envtab.ReadLoadout(loadout)
		if err != nil {
			println("ERROR: Failure reading loadout [%s]: %s\n", loadout, err)
			os.Exit(1)
		}

		if loadoutStruct.Metadata.Login == true {
			println("DEBUG: Loadout [" + loadout + "] has login enabled")
			loadoutStruct.Export()
		} else {
			println("DEBUG: Loadout [" + loadout + "] has login disabled")
		}
	}
}
