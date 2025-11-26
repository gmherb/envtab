/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/login"
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
		logger.Debug("login called")

		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		status, _ := cmd.Flags().GetBool("status")

		if enable {
			logger.Debug("enabling login")
			login.EnableLogin()
			return
		}
		if disable {
			logger.Debug("disabling login")
			login.DisableLogin()
			return
		}
		if status {
			logger.Debug("showing status")
			login.ShowLoginStatus()
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
	loadouts, err := backends.ListLoadouts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failure listing loadouts: %s\n", err)
		os.Exit(1)
	}

	for _, loadout := range loadouts {

		lo, err := backends.ReadLoadout(loadout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failure reading loadout [%s]: %s\n", loadout, err)
			os.Exit(1)
		}

		if lo.Metadata.Login {
			logger.Debug("loadout has login enabled", "loadout", loadout)
			lo.Export()
		} else {
			logger.Debug("loadout has login disabled", "loadout", loadout)
		}
	}
}
