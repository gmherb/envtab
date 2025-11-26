/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/gmherb/envtab/internal/crypto"
	"github.com/gmherb/envtab/internal/envtab"
	"github.com/gmherb/envtab/internal/env"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:                   "show",
	Short:                 "Show active loadouts",
	Long:                  `Show each loadout with active entries (environment variables).`,
	Args:                  cobra.NoArgs,
	SuggestFor:            []string{"status"},
	Aliases:               []string{"sh", "sho"},
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("show called")
		showSensitive, _ := cmd.Flags().GetBool("sensitive")
		showActiveLoadouts(showSensitive)
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().BoolP("sensitive", "s", false, "Show decrypted sensitive values (SOPS encrypted)")
}

func showActiveLoadouts(showSensitive bool) {
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
		// Create decrypt function for comparing encrypted values
		decryptFunc := func(encryptedValue string) (string, error) {
			if strings.HasPrefix(encryptedValue, "SOPS:") {
				return crypto.SOPSDecryptValue(encryptedValue)
			}
			return encryptedValue, nil
		}

		// Create display function that conditionally shows decrypted values
		displayValue := func(value string) string {
			if strings.HasPrefix(value, "SOPS:") {
				if showSensitive {
					decrypted, err := crypto.SOPSDecryptValue(value)
					if err != nil {
						return "***encrypted***"
					}
					return decrypted
				}
				return "***encrypted***"
			}
			return value
		}

		for key, value := range lo.Entries {
			if environment.CompareWithDecrypt(key, value, decryptFunc) {
				// Display value (decrypted if showSensitive is true)
				displayVal := displayValue(value)
				activeEntries = append(activeEntries, key+"="+displayVal)
			}
		}

		// TODO: Add different color pattern/options
		entryColor := color.New(color.FgHiWhite).SprintFunc()
		dashColor := color.New(color.FgHiBlack).SprintFunc()
		loColor := color.New(color.FgGreen).SprintFunc()

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

			dashCount := 80 - // term width
				len(loadout) -
				len(fmt.Sprint(activeEntryCount)) -
				len(fmt.Sprint(totalEntryCount)) -
				10 // magic number

			fmt.Println(
				loColor(loadout),
				strings.Repeat(dashColor("-"), dashCount),
				"[", countColor(len(activeEntries)), "/", countColor(totalEntryCount), "]",
			)
			padding := "   "
			for _, entry := range activeEntries {
				fmt.Println(padding, entryColor(entry))
			}
		}
	}

}
