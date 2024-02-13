/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gmherb/envtab/cmd/envtab"
	tagz "github.com/gmherb/envtab/pkg/tags"
	"github.com/gmherb/envtab/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var editCmd = &cobra.Command{
	Use:   "edit LOADOUT_NAME [flags]",
	Short: "Edit envtab loadout",
	Long: `Edit envtab loadout name, description, tags, and whether its enabled on login.
If no options are provided, enter editor to manually edit a envtab loadout.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		println("DEBUG: edit called")

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

			err = envtab.WriteLoadout(loadoutName, loadout)
			if err != nil {
				fmt.Printf("ERROR: Failure writing loadout [%s]: %s\n", loadoutName, err)
				os.Exit(1)
			}
		} else {
			editLoadout(loadoutName)
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringP("name", "n", "", "set loadout name")
	editCmd.Flags().StringP("description", "d", "", "set loadout description")
	editCmd.Flags().StringP("tags", "t", "", "set loadout tags (separated by comma or space)")

	editCmd.Flags().BoolP("login", "l", false, "enable loadout on login (mutually exclusive with --no-login)")
	editCmd.Flags().BoolP("no-login", "L", false, "disable loadout on login (mutually exclusive with --login)")
	editCmd.MarkFlagsMutuallyExclusive("login", "no-login")
}

func editLoadout(loadoutName string) error {

	var loadout envtab.Loadout
	var editedLoadout envtab.Loadout

	envtabPath := envtab.InitEnvtab("")
	loadoutPath := filepath.Join(envtabPath, loadoutName+".yaml")
	tempFilePath := loadoutPath + ".tmp"

	// Check if the loadout exists
	if _, err := os.Stat(loadoutPath); os.IsNotExist(err) {
		fmt.Printf("ERROR: Loadout [%s] does not exist\n", loadoutName)
		os.Exit(1)
	}

	// Read the loadout file
	data, err := os.ReadFile(loadoutPath)
	if err != nil {
		return err
	}

	// Load loadout yaml file into a Loadout struct
	err = yaml.Unmarshal(data, &loadout)
	if err != nil {
		return err
	}

	// Save the original timestamps
	createdAt := loadout.Metadata.CreatedAt
	updatedAt := loadout.Metadata.LoadedAt

	// Write the Loadout struct to a temp file
	err = os.WriteFile(tempFilePath, data, 0600)
	if err != nil {
		return err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// Loop until a valid loadout is provided or user aborts
	for {

		// Open the temp file in the editor
		cmd := exec.Command(editor, tempFilePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Run()

		// Read the temp file
		data, err = os.ReadFile(tempFilePath)
		if err != nil {
			return err
		}

		// Load yaml file into a Loadout struct
		err = yaml.Unmarshal(data, &editedLoadout)

		// If the contents of the file could not be parsed
		// Ask the user to continue editing the file or abort
		if err != nil {

			usersChoice := utils.PromptForAnswer("The file could not be parsed. Do you want to continue editing to try to fix the errors? Enter 'yes' to continue to edit or 'no' to abort and discard changes?")
			if !usersChoice {
				return nil
			}
		}

		// If the contents of the file could be parsed
		// Break the loop
		if err == nil {
			break
		}
	}

	// Restore the original timestamps
	editedLoadout.Metadata.CreatedAt = createdAt
	editedLoadout.Metadata.UpdatedAt = updatedAt

	// Only overwrite the loadout when modified
	if envtab.CompareLoadouts(loadout, editedLoadout) {
		editedLoadout.UpdateUpdatedAt()

		return envtab.WriteLoadout(loadoutName, &editedLoadout)
	}

	// Remove the temp file
	os.Remove(tempFilePath)
	return nil
}
