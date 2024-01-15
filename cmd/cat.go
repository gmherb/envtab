/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/gmherb/envtab/cmd/envtab"
	"github.com/spf13/cobra"
)

const (
	CAT_USAGE = `Usage: envtab cat <name>`
)

var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "Print an envtab loadout",
	Long: `Print the YAML file which contains the envtab loadout with the name
the provided argument.

` + CAT_USAGE + `

`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DEBUG: cat called")

		if len(args) != 1 {
			fmt.Println("ERROR: Insufficient number of arguments provided")
			fmt.Println(CAT_USAGE)
			os.Exit(1)
		}

		envtabPath := envtab.InitEnvtab()

		content, err := os.ReadFile(envtabPath + "/" + args[0] + ".yaml")
		if err != nil {

			if os.IsNotExist(err) {
				fmt.Printf("ERROR: Loadout [%s] does not exist\n", args[0])
				os.Exit(1)
			}

			fmt.Printf("ERROR: Error reading %s: %s\n", envtabPath, err)
			os.Exit(1)
		}
		fmt.Printf("%s", string(content))

	},
}

func init() {
	rootCmd.AddCommand(catCmd)
}
