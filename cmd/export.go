/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/config"
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
		slog.Debug("export called")

		envtabPath := config.InitEnvtab("")

		for _, arg := range args {

			loadoutName := arg
			loadoutPath := envtabPath + `/` + loadoutName + `.yaml`

			slog.Debug("exporting loadout", "loadout", loadoutName, "path", loadoutPath)

		if _, err := os.Stat(loadoutPath); os.IsNotExist(err) {
			slog.Error("loadout does not exist", "loadout", loadoutName)
			os.Exit(1)
		}

		loadout, err := backends.ReadLoadout(loadoutName)
		if err != nil {
			// Skip loadout if SOPS is not installed (for encrypted loadouts)
			errStr := err.Error()
			if strings.Contains(errStr, "SOPS_NOT_INSTALLED") {
				slog.Warn("skipping loadout - SOPS not installed", "loadout", loadoutName)
				continue
			}
			slog.Error("failure reading loadout", "loadout", loadoutName, "error", err)
			os.Exit(1)
		}

			loadout.Export()
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
