/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/login"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Export all login loadouts",
	Long: `Export all loadouts which are enabled on login.

This is typically sourced from a login script such as ~/.profile.

To setup login automatically, run:
  envtab login --enable

To disable login, run:
  envtab login --disable

To show the status of login, run:
  envtab login --status`,
	Args:    cobra.NoArgs,
	Aliases: []string{"lo", "log", "logi"},
	Example: `  envtab login
  envtab login --status
  envtab login --enable
  envtab login --disable`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("login called")

		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		status, _ := cmd.Flags().GetBool("status")

		if enable {
			slog.Debug("enabling login")
			login.EnableLogin()
			return
		}
		if disable {
			slog.Debug("disabling login")
			login.DisableLogin()
			return
		}
		if status {
			slog.Debug("showing status")
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
		slog.Error("failure listing loadouts", "error", err)
		os.Exit(1)
	}

	for _, loadout := range loadouts {
		lo, err := backends.ReadLoadout(loadout)
		if err != nil {
			// Skip loadout if SOPS is not installed (for encrypted loadouts)
			errStr := err.Error()
			if strings.Contains(errStr, "SOPS_NOT_INSTALLED") {
				slog.Warn("skipping loadout - SOPS not installed", "loadout", loadout)
				continue
			}
			slog.Error("failure reading loadout", "loadout", loadout, "error", err)
			os.Exit(1)
		}

		if lo.Metadata.Login {
			slog.Debug("loadout has login enabled", "loadout", loadout)
			lo.Export()
		} else {
			slog.Debug("loadout has login disabled", "loadout", loadout)
		}
	}
}
