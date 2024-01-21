/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gmherb/envtab/cmd/envtab"
	"github.com/gmherb/envtab/pkg/env"
	"github.com/gmherb/envtab/pkg/utils"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all envtab loadouts",
	Long: `List all envtab loadouts.  If the --long flag is provided, then
print the long listing format which includes the loadout name, tags, and other
metadata.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DEBUG: list called")

		if cmd.Flag("long").Value.String() == "true" {
			fmt.Println("DEBUG: long listing format")
			ListEnvtabLoadouts()

		} else {
			fmt.Println("DEBUG: short listing format")
			PrintEnvtabLoadouts()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.PersistentFlags().BoolP("long", "l", false, "Print long listing format")
}

func PrintEnvtabLoadouts() {
	loadouts := envtab.GetEnvtabSlice()
	for _, loadouts := range loadouts {
		fmt.Println(loadouts)
	}
}

func ListEnvtabLoadouts() {
	envtabSlice := envtab.GetEnvtabSlice()
	environment := env.NewEnv()
	environment.Populate()

	fmt.Println("UpdatedAt LoadedAt  Login Active Total  Name               Tags")
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

		fmt.Println(
			// TODO: Determine if time is under 24 hours and print TimeOnly instead of DateOnly
			strings.TrimPrefix(updatedAt.Format(time.DateOnly), "20"), "",
			strings.TrimPrefix(loadedAt.Format(time.DateOnly), "20"), "",
			lo.Metadata.Login, " [",
			len(activeEntries), " / ",
			len(lo.Entries), "] ",
			utils.PadString(loadout, 17), "",
			lo.Metadata.Tags)

	}
}
