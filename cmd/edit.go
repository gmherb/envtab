/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/gmherb/envtab/cmd/envtab"
	tagz "github.com/gmherb/envtab/pkg/tags"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [loadout]",
	Short: "Edit envtab loadout",
	Long: `Edit envtab loadout name, description, tags, and whether its enabled on login.
If no options are provided, enter editor to manually edit a envtab loadout.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DEBUG: edit called")

		if len(args) > 1 {
			fmt.Printf("ERROR: edit only accepts one loadout name\n")
			os.Exit(1)
		}

		if len(args) == 0 {
			fmt.Printf("ERROR: edit requires a loadout name\n")
			os.Exit(1)
		}

		loadoutName := args[0]
		println("DEBUG: Loadout: " + loadoutName)

		loadoutModified := false

		// If --name is set, rename the loadout
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			fmt.Printf("DEBUG: Renaming loadout [%s] to [%s]\n", loadoutName, name)
			err := envtab.RenameLoadout(loadoutName, name)
			if err != nil {
				fmt.Printf("ERROR: Failure renaming loadout [%s] to [%s]: %s\n", loadoutName, name, err)
				os.Exit(1)
			}
			loadoutName = name
			loadoutModified = true
		}

		// load the loadout
		loadout, err := envtab.ReadLoadout(loadoutName)
		if err != nil {
			fmt.Printf("ERROR: Failure reading loadout [%s]: %s\n", loadoutName, err)
			os.Exit(1)
		}

		// If --description is set, update the loadout description
		if description, _ := cmd.Flags().GetString("description"); description != "" {
			fmt.Printf("DEBUG: Updating loadout [%s] description to [%s]\n", loadoutName, description)
			loadout.UpdateDescription(description)
			loadoutModified = true
		}

		// If --login is set, enable loadout on login
		if login, _ := cmd.Flags().GetBool("login"); login {

			// If --login and --no-login are both set, error
			if noLogin, _ := cmd.Flags().GetBool("no-login"); noLogin {
				fmt.Printf("ERROR: --login and --no-login are mutually exclusive\n")
				os.Exit(1)
			}

			fmt.Printf("DEBUG: Enabling loadout [%s] on login\n", loadoutName)
			loadout.UpdateLogin(true)
			loadoutModified = true
		}

		// If --no-login is set, disable loadout on login
		if noLogin, _ := cmd.Flags().GetBool("no-login"); noLogin {
			fmt.Printf("DEBUG: Disabling loadout [%s] on login\n", loadoutName)
			loadout.UpdateLogin(false)
			loadoutModified = true
		}

		// If --tags is set, update the loadout tags
		if tags, _ := cmd.Flags().GetString("tags"); tags != "" {

			newTags := []string{tags}

			newTags = tagz.SplitTags(newTags)
			newTags = tagz.RemoveEmptyTags(newTags)
			newTags = tagz.RemoveDuplicateTags(newTags)

			fmt.Printf("DEBUG: Updating loadout [%s] tags to %s\n", loadoutName, newTags)

			loadout.UpdateTags(newTags)
			loadoutModified = true
		}

		if loadoutModified {
			println("DEBUG: Writing loadout")
			//loadout.PrintLoadout()

			err = envtab.WriteLoadout(loadoutName, loadout)
			if err != nil {
				fmt.Printf("ERROR: Failure writing loadout [%s]: %s\n", loadoutName, err)
				os.Exit(1)
			}
		} else {
			// TODO: Do not edit directly, instead copy to a temp file and edit that
			// and then copy back to the original file if valid
			editLoadout(loadoutName)
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringP("name", "n", "", "loadout name")
	editCmd.Flags().StringP("description", "d", "", "loadout description")
	editCmd.Flags().BoolP("login", "l", false, "enable loadout on login")
	editCmd.Flags().BoolP("no-login", "L", false, "disable loadout on login (default)")
	editCmd.Flags().StringP("tags", "t", "", "loadout tags (comma or space separated)")
}

func editLoadout(loadoutName string) {

	envtabPath := envtab.InitEnvtab("")
	loadoutPath := envtabPath + `/` + loadoutName + `.yaml`

	if _, err := os.Stat(loadoutPath); os.IsNotExist(err) {
		fmt.Printf("ERROR: Loadout [%s] does not exist\n", loadoutName)
		os.Exit(1)
	}

	err := envtab.EditLoadout(loadoutName)
	if err != nil {
		fmt.Printf("ERROR: Failure editing loadout [%s]: %s\n", loadoutName, err)
		os.Exit(1)
	}

}
