/*
Copyright © 2024 Greg Herbster
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/env"
	"github.com/gmherb/envtab/pkg/sops"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [LOADOUT_PATTERN...]",
	Short: "List all envtab loadouts",
	Long: `List all envtab loadouts. Optional glob patterns can be provided to
narrow results. If multiple patterns are provided, loadouts matching any
pattern will be shown. If the --long flag is provided, then print the long
listing format which includes the loadout name, tags, and other metadata.`,
	Example: `  envtab list
  envtab list -l
  envtab list --long
  envtab list dev*
  envtab list -l *staging*
  envtab list aws-* gcp-*`,
	Args:    cobra.ArbitraryArgs,
	Aliases: []string{"l", "ls", "lis"},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("list called with args", "args", args)

		if long, _ := cmd.Flags().GetBool("long"); long {
			slog.Debug("long listing format")
			ListEnvtabLoadouts(args)

		} else {
			slog.Debug("short listing format")
			PrintEnvtabLoadouts(args)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.PersistentFlags().BoolP("long", "l", false, "Print long listing format")
}

func PrintEnvtabLoadouts(patterns []string) {
	loadouts, err := backends.ListLoadouts()
	if err != nil {
		slog.Error("failure listing loadouts", "error", err)
		os.Exit(1)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	for _, loadout := range loadouts {
		// Filter by patterns if provided
		if len(patterns) > 0 {
			matched := false
			for _, pattern := range patterns {
				m, _ := filepath.Match(pattern, loadout)
				if m {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		fmt.Fprintf(tw, "%s\t", loadout)
	}
	fmt.Fprintln(tw)
	tw.Flush()
}

func ListEnvtabLoadouts(patterns []string) {
	envtabSlice, err := backends.ListLoadouts()
	if err != nil {
		slog.Error("failure listing loadouts", "error", err)
		os.Exit(1)
	}
	environment := env.NewEnv()
	environment.Populate()

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, "UpdatedAt\tLoadedAt\tLogin\tTotal\tActive\tName\tTags\n")

	for _, loadout := range envtabSlice {

		// Filter by patterns if provided
		if len(patterns) > 0 {
			matched := false
			for _, pattern := range patterns {
				m, _ := filepath.Match(pattern, loadout)
				if m {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

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

		// Handle empty timestamps gracefully
		var updatedAt time.Time
		var loadedAt time.Time

		if lo.Metadata.UpdatedAt != "" {
			var err error
			updatedAt, err = time.Parse(time.RFC3339, lo.Metadata.UpdatedAt)
			if err != nil {
				slog.Warn("invalid updatedAt time, using current time", "loadout", loadout, "error", err)
				updatedAt = time.Now()
			}
		} else {
			// Use current time if empty
			updatedAt = time.Now()
		}

		if lo.Metadata.LoadedAt != "" {
			var err error
			loadedAt, err = time.Parse(time.RFC3339, lo.Metadata.LoadedAt)
			if err != nil {
				slog.Warn("invalid loadedAt time, using current time", "loadout", loadout, "error", err)
				loadedAt = time.Now()
			}
		} else {
			// Use current time if empty
			loadedAt = time.Now()
		}

		var activeEntries = []string{}
		// Create decrypt function for comparing encrypted values
		decryptFunc := func(encryptedValue string) (string, error) {
			if strings.HasPrefix(encryptedValue, "SOPS:") {
				return sops.SOPSDecryptValue(encryptedValue)
			}
			return encryptedValue, nil
		}

		for key, value := range lo.Entries {
			if environment.CompareWithDecrypt(key, value, decryptFunc) {
				activeEntries = append(activeEntries, key+"="+value)
			}
		}

		var updatedAtTime string
		var loadedAtTime string
		if time.Since(updatedAt) > 24*time.Hour {
			loadedAtTime = strings.TrimPrefix(loadedAt.Format(time.DateOnly), "20")
			updatedAtTime = strings.TrimPrefix(updatedAt.Format(time.DateOnly), "20")
		} else {
			loadedAtTime = loadedAt.Format(time.TimeOnly)
			updatedAtTime = updatedAt.Format(time.TimeOnly)
		}

		fmt.Fprintf(tw, "%s\t%s\t%t\t%d\t%d\t%s\t%s\n",
			updatedAtTime,
			loadedAtTime,
			lo.Metadata.Login,
			len(lo.Entries),
			len(activeEntries),
			loadout,
			lo.Metadata.Tags)
	}
	fmt.Fprintln(tw)
	tw.Flush()
}
