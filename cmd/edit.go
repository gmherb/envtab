/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gmherb/envtab/internal/crypto"
	"github.com/gmherb/envtab/internal/envtab"
	"github.com/gmherb/envtab/internal/tags"
	"github.com/gmherb/envtab/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var editCmd = &cobra.Command{
	Use:   "edit LOADOUT_NAME",
	Short: "Edit envtab loadout",
	Long: `Edit envtab loadout name, description, tags, and whether its enabled
on login.

If no options are provided, enter editor to manually edit a envtab loadout.`,
	Example: `  envtab edit myloadout                                  # edit loadout in editor
  envtab edit myloadout --name newname                   # rename loadout
  envtab edit myloadout --description "new description"  # update description
  envtab edit myloadout --tags "tag1, tag2, tag3"        # update tags
  envtab edit myloadout --login                          # enable login on loadout
  envtab edit myloadout --no-login                       # disable login on loadout
  envtab edit myloadout -n newloadout -d "blah bla" -l   # update multiple fields`,
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"ed", "edi"},
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

		// Check if file is SOPS-encrypted to preserve encryption on save
		envtabPath := envtab.InitEnvtab("")
		loadoutPath := filepath.Join(envtabPath, loadoutName+".yaml")
		isSOPSEncrypted := crypto.IsSOPSEncrypted(loadoutPath)

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
		if tagsStr, _ := cmd.Flags().GetString("tags"); tagsStr != "" {
			newTags := []string{tagsStr}

			newTags = tags.SplitTags(newTags)
			newTags = tags.RemoveEmptyTags(newTags)
			newTags = tags.RemoveDuplicateTags(newTags)

			fmt.Printf("DEBUG: Updating loadout [%s] tags to %s\n", loadoutName, newTags)

			loadout.UpdateTags(newTags)
			loadoutModified = true
		}

		if loadoutModified {
			println("DEBUG: Writing loadout")

			// Preserve SOPS encryption if the file was originally encrypted
			if isSOPSEncrypted {
				err = envtab.WriteLoadoutWithEncryption(loadoutName, loadout, true)
			} else {
				err = envtab.WriteLoadout(loadoutName, loadout)
			}
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

	envtabPath := envtab.InitEnvtab("")
	loadoutPath := filepath.Join(envtabPath, loadoutName+".yaml")

	// Check if the loadout exists
	if _, err := os.Stat(loadoutPath); os.IsNotExist(err) {
		fmt.Printf("ERROR: Loadout [%s] does not exist\n", loadoutName)
		os.Exit(1)
	}

	// Check if file is SOPS-encrypted to preserve encryption on save
	isSOPSEncrypted := crypto.IsSOPSEncrypted(loadoutPath)

	// Read the loadout (handles SOPS decryption automatically)
	loadout, err := envtab.ReadLoadout(loadoutName)
	if err != nil {
		return err
	}

	// Decrypt all SOPS-encrypted values for editing
	// Keep track of which keys were encrypted so we can re-encrypt them on save
	encryptedKeys, err := loadout.DecryptSOPSValues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Some SOPS values could not be decrypted: %s\n", err)
	}

	// Marshal to get YAML for editing (now with decrypted values)
	data, err := yaml.Marshal(loadout)
	if err != nil {
		return err
	}

	// Save the original timestamps
	createdAt := loadout.Metadata.CreatedAt
	updatedAt := loadout.Metadata.LoadedAt

	tempFilePath := loadoutPath + ".tmp"

	// Write the Loadout struct to a temp file
	err = os.WriteFile(tempFilePath, data, 0600)
	if err != nil {
		return err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	var editedLoadout *envtab.Loadout

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
		editedLoadout = &envtab.Loadout{}
		err = yaml.Unmarshal(data, editedLoadout)

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
	if envtab.CompareLoadouts(*loadout, *editedLoadout) {
		editedLoadout.UpdateUpdatedAt()

		// Re-encrypt values that were originally SOPS-encrypted
		if len(encryptedKeys) > 0 {
			err := editedLoadout.ReencryptSOPSValues(encryptedKeys)
			if err != nil {
				return fmt.Errorf("failed to re-encrypt SOPS values: %w", err)
			}
		}

		// Preserve SOPS encryption if the file was originally encrypted
		if isSOPSEncrypted {
			return envtab.WriteLoadoutWithEncryption(loadoutName, editedLoadout, true)
		}
		return envtab.WriteLoadout(loadoutName, editedLoadout)
	}

	// Remove the temp file
	os.Remove(tempFilePath)
	return nil
}
