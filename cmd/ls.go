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

	"github.com/gmherb/envtab/cmd/envtab"
	"github.com/gmherb/envtab/pkg/env"

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

		if cmd.Flag("long").Value.String() == "true" {
			println("DEBUG: long listing format")
			ListEnvtabLoadouts()

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
	loadoutSlice := envtab.GetEnvtabSlice(envtabPath)

	var loadouts []string
	if len(glob) > 0 {
		println("DEBUG: glob pattern matching ", glob)
		for _, l := range loadoutSlice {
			matched, _ := filepath.Match(glob, l)
			if matched {
				loadouts = append(loadouts, l)
			}
		}
	} else {
		loadouts = loadoutSlice
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	for i := 0; i < len(loadouts); i++ {
		fmt.Fprintf(tw, "%s\t", loadouts[i])
	}
	fmt.Fprintln(tw)
	tw.Flush()
}

func ListEnvtabLoadouts() {
	envtabSlice := envtab.GetEnvtabSlice("")
	environment := env.NewEnv()
	environment.Populate()

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, "UpdatedAt\tLoadedAt\tLogin\tActive\tTotal\tName\tTags\n")

	for _, loadout := range envtabSlice {

		lo, err := envtab.ReadLoadout(loadout)
		if err != nil {
			fmt.Printf("Error reading loadout %s: %s\n", loadout, err)
			os.Exit(1)
		}

		updatedAt, err := time.Parse(time.RFC3339, lo.Metadata.UpdatedAt)
		if err != nil {
			fmt.Printf("Error parsing updatedAt time %s: %s\n", lo.Metadata.UpdatedAt, err)
			os.Exit(1)
		}

		loadedAt, err := time.Parse(time.RFC3339, lo.Metadata.LoadedAt)
		if err != nil {
			fmt.Printf(
				"Error parsing loadedAt time %s: %s\n",
				lo.Metadata.UpdatedAt, err,
			)
			os.Exit(1)
		}

		var activeEntries = []string{}
		for key, value := range lo.Entries {

			if environment.Compare(key, value) {
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
			len(activeEntries),
			len(lo.Entries),
			loadout,
			lo.Metadata.Tags)
	}
	fmt.Fprintln(tw)
	tw.Flush()
}
