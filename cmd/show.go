/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/gmherb/envtab/pkg/env"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Display environment and highlight active envtab entries",
	Long: `Print each environment variable currently set in the environment
	and highlight those that match an entry in the active envtab file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("show called")

		e := env.NewEnv()
		e.Populate()
		e.Print()
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
