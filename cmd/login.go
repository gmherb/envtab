/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/gmherb/envtab/internal/envtab"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Export all login loadouts",
	Long: `Export all loadouts which are enabled on login. This is typically
run from a login script such as ~/.bash_profile or ~/.zprofile and can be setup
automatically by running "envtab login --enable". You can disable login
by running "envtab login --disable".`,
	Args:    cobra.NoArgs,
	Aliases: []string{"lo", "log", "logi"},
	Run: func(cmd *cobra.Command, args []string) {
		println("DEBUG: login called")

		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		status, _ := cmd.Flags().GetBool("status")

		if enable {
			println("DEBUG: Enabling login")
			envtab.EnableLogin()
			return
		}
		if disable {
			println("DEBUG: Disabling login")
			envtab.DisableLogin()
			return
		}
		if status {
			println("DEBUG: Showing status")
			envtab.ShowLoginStatus()
			return
		}

		exportLoginLoadouts()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().BoolP("enable", "e", false, "Setup envtab to load on shell login")
	loginCmd.Flags().BoolP("disable", "d", false, "Remove envtab from your login scripts")
	loginCmd.Flags().BoolP("status", "s", false, "Show the status of envtab in your login scripts")
	loginCmd.MarkFlagsMutuallyExclusive("enable", "disable", "status")
}

func exportLoginLoadouts() {
	loadouts := envtab.GetEnvtabSlice("")

	for _, loadout := range loadouts {

		lo, err := envtab.ReadLoadout(loadout)
		if err != nil {
			println("ERROR: Failure reading loadout [%s]: %s\n", loadout, err)
			os.Exit(1)
		}

		if lo.Metadata.Login {
			println("DEBUG: Loadout [" + loadout + "] has login enabled")
			lo.Export()
		} else {
			println("DEBUG: Loadout [" + loadout + "] has login disabled")
		}
	}
}
