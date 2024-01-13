/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gmherb/envtab/cmd/envtab"
	"github.com/gmherb/envtab/pkg/env"

	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("show called")

		envtabPath := envtab.InitEnvtab()

		if len(args) < 1 {
			e := env.NewEnv()
			e.Populate()
			e.Print()
		} else {
			content, err := ioutil.ReadFile(envtabPath + "/" + args[0])
			if err != nil {
				fmt.Printf("Error reading %s: %s\n", envtabPath, err)
				os.Exit(1)
			}
			fmt.Printf("%s", string(content))

		}

	},
}

func init() {
	rootCmd.AddCommand(showCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// showCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
