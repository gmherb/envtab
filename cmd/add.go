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

var addCmd = &cobra.Command{
	Use:   "add LOADOUT_NAME KEY=VALUE [TAG1 TAG2,TAG3 ...]",
	Short: "Add an entry to a envtab loadout",
	Long: `Add an environment variable key-pair as an entry in an envtab
loadout.

Add tags to your envtab loadout by adding them after the key and value. Multiple
tags can be provided using space or comma as a separator.`,
	Example: `  envtab add myloadout MY_ENV_VAR=myvalue
  envtab add myloadout MY_ENV_VAR=myvalue tag1,tag2,tag3
  envtab add myloadout MY_ENV_VAR=myvalue tag1 tag2 tag3
  envtab add myloadout MY_ENV_VAR=myvalue tag1,tag2, tag3 tag4,tag5`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MinimumNArgs(2),
	Aliases:               []string{"a", "ad"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DEBUG: add command called")

		var (
			name  string   // envtab loadout name
			key   string   // Environment variable key
			value string   // Environment variable value
			tags  []string // Tags to append to envtab loadout
		)

		if len(args) == 2 && !strings.Contains(args[1], "=") {
			fmt.Println("DEBUG: No value provided for your envtab entry. No equal sign detected and only 2 args provided.")
			cmd.Usage()
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

		fmt.Printf("DEBUG: Name: %s, Key: %s, Value: %s, tags: %s.\n", name, key, value, tags)

		err := envtab.AddEntryToLoadout(name, key, value, tags)
		if err != nil {
			fmt.Printf("ERROR: Error writing entry to file [%s]: %s\n", name, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
