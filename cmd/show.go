/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/gmherb/envtab/cmd/envtab"
	"github.com/gmherb/envtab/pkg/env"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:                   "show",
	Short:                 "Show active loadouts",
	Long:                  `Show each loadout with active entries (environment variables).`,
	Args:                  cobra.NoArgs,
	SuggestFor:            []string{"status"},
	Aliases:               []string{"s", "sh", "sho"},
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		println("DEBUG: show called")
		showActiveLoadouts()
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func showActiveLoadouts() {
	envtabSlice := envtab.GetEnvtabSlice("")
	environment := env.NewEnv()
	environment.Populate()

	for _, loadout := range envtabSlice {

		lo, err := envtab.ReadLoadout(loadout)
		if err != nil {
			fmt.Printf("Error reading loadout %s: %s\n", loadout, err)
			os.Exit(1)
		}

		var activeEntries = []string{}
		for key, value := range lo.Entries {
			if environment.Compare(key, value) {
				activeEntries = append(activeEntries, key+"="+value)
			}
		}

		green := color.New(color.FgGreen).SprintFunc()
		purple := color.New(color.FgHiMagenta).SprintFunc()

		activeEntryCount := len(activeEntries)
		totalEntryCount := len(lo.Entries)

		// If a loadout has fewer active entries than total entries, colorize the count red
		var countColor func(a ...interface{}) string
		if len(activeEntries) < len(lo.Entries) {
			countColor = color.New(color.FgRed).SprintFunc()
		} else {
			countColor = color.New(color.FgBlue).SprintFunc()
		}

		if len(activeEntries) > 0 {

			d := 80 - // term width
				len(loadout) -
				len(fmt.Sprint(activeEntryCount)) -
				len(fmt.Sprint(totalEntryCount)) -
				10 // magic number
			fmt.Println(
				green(loadout),
				strings.Repeat(purple("-"), d),
				"[", countColor(len(activeEntries)), "/", countColor(totalEntryCount), "]",
			)
			for _, entry := range activeEntries {
				fmt.Println("  ", entry)
			}
			fmt.Println()
		}
	}

}
