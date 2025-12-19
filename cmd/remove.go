package cmd

import (
	"log/slog"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove LOADOUT_NAME [LOADOUT_NAME ...]",
	Short: "Remove envtab loadout(s)",
	Long:  `Remove envtab loadout(s)`,
	Example: `  envtab remove myloadout
  envtab remove myloadout1 myloadout2 myloadout3`,
	Args:       cobra.MinimumNArgs(1),
	SuggestFor: []string{"delete", "del"},
	Aliases:    []string{"r", "rm"},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("remove called")
		for _, loadout := range args {
			slog.Debug("removing loadout", "loadout", loadout)
			backends.RemoveLoadout(loadout)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
