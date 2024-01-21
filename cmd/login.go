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

		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		if enable && disable {
			println("ERROR: Cannot enable and disable login at the same time")
			os.Exit(1)
		}

		if enable {
			println("DEBUG: Enabling login")
			envtab.EnableLogin()
			return
		} else if disable {
			println("DEBUG: Disabling login")
			envtab.DisableLogin()
			return
		}

		exportLoginLoadouts()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().BoolP("enable", "e", false, "Enable login")
	loginCmd.Flags().BoolP("disable", "d", false, "Disable login")
}

func exportLoginLoadouts() {
	loadouts := envtab.GetEnvtabSlice()

	for _, loadout := range loadouts {

		lo, err := envtab.ReadLoadout(loadout)
		if err != nil {
			println("ERROR: Failure reading loadout [%s]: %s\n", loadout, err)
			os.Exit(1)
		}

		if lo.Metadata.Login == true {
			println("DEBUG: Loadout [" + loadout + "] has login enabled")
			lo.Export()
		} else {
			println("DEBUG: Loadout [" + loadout + "] has login disabled")
		}
	}
}
