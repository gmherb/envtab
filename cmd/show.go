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
	Use:   "show",
	Short: "Show active loadouts",
	Long:  `Show each loadout with active entries (environment variables).`,
	Run: func(cmd *cobra.Command, args []string) {
		println("DEBUG: show called")
		showActiveLoadouts()
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func showActiveLoadouts() {
	envtabSlice := envtab.GetEnvtabSlice()
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

		// If a loadout has fewer active entries than total entries, colorize the count red
		var countColor func(a ...interface{}) string
		if len(activeEntries) < len(lo.Entries) {
			countColor = color.New(color.FgRed).SprintFunc()
		} else {
			countColor = color.New(color.FgBlue).SprintFunc()
		}

		if len(activeEntries) > 0 {

			// 80 is term width, 9 is the length of the [x/y] string
			t := 80 - len(loadout) + 9
			fmt.Println(
				green(loadout), strings.Repeat(purple("-"), t),
				"[", countColor(len(activeEntries)), "/", countColor(len(lo.Entries)), "]",
			)
			for _, entry := range activeEntries {
				fmt.Println("  ", entry)
			}
			fmt.Println()
		}
	}

}
