/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gmherb/envtab/cmd/envtab"
	tagz "github.com/gmherb/envtab/pkg/tags"
	"github.com/spf13/cobra"
)

const (
	ADD_USAGE = `Usage: envtab add <name> <key>=<value> [tag1 tag2 ...]`
)

var addCmd = &cobra.Command{
	Use:   "add <name> <key>=<value> [tag1 tag2 ...]",
	Short: "Add an envtab entry to a loadout",
	Long: `Add an environment variable and its value, KEY=value, as an entry in
an envtab loadout.

Optionally, you can add tags to the envtab loadout by adding them after the key
and value. Multiple tags can be provided using space as a separator.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DEBUG: add command called")

		var (
			name  string   // Envtab entry name
			key   string   // Environment variable key
			value string   // Environment variable value
			tags  []string // Tags for the envtab entry

		)

		if len(args) < 2 {
			fmt.Println("DEBUG: Insufficient number of arguments provided")
			fmt.Println(ADD_USAGE)
			os.Exit(1)
		}

		if len(args) == 2 && !strings.Contains(args[1], "=") {
			fmt.Println("DEBUG: No value provided for your envtab entry. No equal sign detected and only 2 args provided.")
			fmt.Println(ADD_USAGE)
			os.Exit(1)
		}

		name = args[0]

		if strings.Contains(args[1], "=") {
			fmt.Println("DEBUG: Equal sign detected in second argument. Splitting into key and value.")
			key, value = strings.Split(args[1], "=")[0], strings.Split(args[1], "=")[1]
			tags = args[2:]

		} else {
			fmt.Println("DEBUG: No equal sign detected in second argument. Assigning second argument as key.")
			key = args[1]
			value = args[2]
			tags = args[3:]
		}

		tags = tagz.SplitTags(tags)
		tags = tagz.RemoveEmptyTags(tags)
		tags = tagz.RemoveDuplicateTags(tags)

		fmt.Printf("DEBUG: Name: %s, Key: %s, Value: %s, tags: %s.", name, key, value, tags)

		err := envtab.WriteEntryToLoadout(name, key, value, tags)
		if err != nil {
			fmt.Printf("ERROR: Error writing entry to file [%s]: %s\n", name, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
