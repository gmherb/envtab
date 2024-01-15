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
	USAGE = `Usage: envtab add <name> <key>=<value> [tag1 tag2 ...]`
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an envtab entry to a loadout",
	Long: `Add an environment variable and its value, KEY=value, as an entry in
an envtab .

` + USAGE + `

The first argument is the name of the entry followed by the key and value
of the environment variable. Optionally, you can add tags to the envtab
table by adding them after the key value pair (multiple can be provided
using space as a separator). By default, the table is a	YAML file in the
envtab directory which resides in the user's home directory (~/.envtab).`,
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
			fmt.Println(USAGE)
			os.Exit(1)
		}

		if len(args) == 2 && !strings.Contains(args[1], "=") {
			fmt.Println("DEBUG: No value provided for your envtab entry. No equal sign detected and only 2 args provided.")
			fmt.Println(USAGE)
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

		fmt.Printf("DEBUG: Name: %s, Key: %s, Value: %s, tags: %s.", name, key, value, tags)

		err := envtab.WriteEntryToLoadout(name, key, value, tags)
		if err != nil {
			fmt.Printf("Error writing entry to file [%s]: %s\n", name, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
