/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/gmherb/envtab/internal/crypto"
	"github.com/gmherb/envtab/internal/envtab"
	"github.com/gmherb/envtab/internal/env"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all envtab loadouts",
	Long: `List all envtab loadouts. A glob pattern can be provided to narrow results. If the --long flag is provided, then
print the long listing format which includes the loadout name, tags, and other
metadata.`,
	Example: `  envtab ls
  envtab ls -l
  envtab ls --long
  envtab ls dev*
  envtab ls -l *staging*`,
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"l", "list"},
	Run: func(cmd *cobra.Command, args []string) {
		println("DEBUG: list called")

		long, _ := cmd.Flags().GetBool("long")
		if long {
			println("DEBUG: long listing format")
			if len(args) > 0 {
				ListEnvtabLoadouts(args[0])
			} else {
				ListEnvtabLoadouts("")
			}

		} else {
			println("DEBUG: short listing format")
			if len(args) > 0 {
				PrintEnvtabLoadouts(args[0])
			} else {
				PrintEnvtabLoadouts("")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.PersistentFlags().BoolP("long", "l", false, "Print long listing format")
}

func PrintEnvtabLoadouts(glob string) {
	envtabPath := envtab.InitEnvtab("")
	loadouts := envtab.GetEnvtabSlice(envtabPath)

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	for _, loadout := range loadouts {
		if len(glob) > 0 {
			matched, _ := filepath.Match(glob, loadout)
			if !matched {
				continue
			}
		}
		fmt.Fprintf(tw, "%s\t", loadout)
	}
	fmt.Fprintln(tw)
	tw.Flush()
}

func ListEnvtabLoadouts(glob string) {
	envtabSlice := envtab.GetEnvtabSlice("")
	environment := env.NewEnv()
	environment.Populate()

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, "UpdatedAt\tLoadedAt\tLogin\tTotal\tActive\tName\tTags\n")

	for _, loadout := range envtabSlice {

		if len(glob) > 0 {
			matched, _ := filepath.Match(glob, loadout)
			if !matched {
				continue
			}
		}

		lo, err := envtab.ReadLoadout(loadout)
		if err != nil {
			fmt.Printf("Error reading loadout %s: %s\n", loadout, err)
			os.Exit(1)
		}

		// Handle empty timestamps gracefully
		var updatedAt time.Time
		var loadedAt time.Time
		
		if lo.Metadata.UpdatedAt != "" {
			var err error
			updatedAt, err = time.Parse(time.RFC3339, lo.Metadata.UpdatedAt)
			if err != nil {
				fmt.Printf("Warning: Invalid updatedAt time for %s: %s (using current time)\n", loadout, err)
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
				fmt.Printf("Warning: Invalid loadedAt time for %s: %s (using current time)\n", loadout, err)
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
				return crypto.SOPSDecryptValue(encryptedValue)
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
			//strings.TrimPrefix(updatedAt.Format(time.DateOnly), "20"),
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
