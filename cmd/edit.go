/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/config"
	"github.com/gmherb/envtab/pkg/sops"
	"github.com/gmherb/envtab/internal/loadout"
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
		logger.Debug("edit called")

		loadoutName := args[0]
		logger.Debug("editing loadout", "loadout", loadoutName)

		loadoutModified := false

		// If --name is set, rename the loadout
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			logger.Debug("renaming loadout", "old", loadoutName, "new", name)
			err := backends.RenameLoadout(loadoutName, name)
			if err != nil {
				fmt.Printf("ERROR: Failure renaming loadout [%s] to [%s]: %s\n", loadoutName, name, err)
				os.Exit(1)
			}
			loadoutName = name
			loadoutModified = true
		}

		// Check if file is SOPS-encrypted to preserve encryption on save
		envtabPath := config.InitEnvtab("")
		loadoutPath := filepath.Join(envtabPath, loadoutName+".yaml")
		isSOPSEncrypted := sops.IsSOPSEncrypted(loadoutPath)

		// load the loadout
		lo, err := backends.ReadLoadout(loadoutName)
		if err != nil {
			fmt.Printf("ERROR: Failure reading loadout [%s]: %s\n", loadoutName, err)
			os.Exit(1)
		}

		// If --description is set, update the loadout description
		if description, _ := cmd.Flags().GetString("description"); description != "" {
			logger.Debug("updating loadout description", "loadout", loadoutName, "description", description)
			lo.UpdateDescription(description)
			loadoutModified = true
		}

		// If --login is set, enable loadout on login
		if login, _ := cmd.Flags().GetBool("login"); login {
			logger.Debug("enabling loadout on login", "loadout", loadoutName)
			lo.UpdateLogin(true)
			loadoutModified = true
		}

		// If --no-login is set, disable loadout on login
		if noLogin, _ := cmd.Flags().GetBool("no-login"); noLogin {
			logger.Debug("disabling loadout on login", "loadout", loadoutName)
			lo.UpdateLogin(false)
			loadoutModified = true
		}

		// If --tags is set, update the loadout tags
		if tagsStr, _ := cmd.Flags().GetString("tags"); tagsStr != "" {
			newTags := []string{tagsStr}

			newTags = tags.SplitTags(newTags)
			newTags = tags.RemoveEmptyTags(newTags)
			newTags = tags.RemoveDuplicateTags(newTags)

			logger.Debug("updating loadout tags", "loadout", loadoutName, "tags", newTags)

			lo.UpdateTags(newTags)
			loadoutModified = true
		}

		if loadoutModified {
			logger.Debug("writing loadout", "loadout", loadoutName)

			// Preserve SOPS encryption if the file was originally encrypted
			if isSOPSEncrypted {
				err = backends.WriteLoadoutWithEncryption(loadoutName, lo, true)
			} else {
				err = backends.WriteLoadout(loadoutName, lo)
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

	envtabPath := config.InitEnvtab("")
	loadoutPath := filepath.Join(envtabPath, loadoutName+".yaml")

	// Check if the loadout exists
	if _, err := os.Stat(loadoutPath); os.IsNotExist(err) {
		fmt.Printf("ERROR: Loadout [%s] does not exist\n", loadoutName)
		os.Exit(1)
	}

	// Check if file is SOPS-encrypted to preserve encryption on save
	isSOPSEncrypted := sops.IsSOPSEncrypted(loadoutPath)

	// Read the loadout (handles SOPS decryption automatically)
	lo, err := backends.ReadLoadout(loadoutName)
	if err != nil {
		return err
	}

	// Decrypt all SOPS-encrypted values for editing
	// Keep track of which keys were encrypted so we can re-encrypt them on save
	encryptedKeys, err := lo.DecryptSOPSValues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Some SOPS values could not be decrypted: %s\n", err)
	}

	// Marshal to get YAML for editing (now with decrypted values)
	data, err := yaml.Marshal(lo)
	if err != nil {
		return err
	}

	// Save the original timestamps
	createdAt := lo.Metadata.CreatedAt
	updatedAt := lo.Metadata.LoadedAt

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

	var editedLoadout *loadout.Loadout

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

		// Validate YAML for duplicate keys before unmarshaling
		err = loadout.ValidateLoadoutYAML(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			usersChoice := utils.PromptForAnswer("The file contains duplicate keys. Do you want to continue editing to fix the errors? Enter 'yes' to continue to edit or 'no' to abort and discard changes?")
			if !usersChoice {
				return nil
			}
			continue // Continue editing
		}

		// Load yaml file into a Loadout struct
		editedLoadout = &loadout.Loadout{}
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
	if loadout.CompareLoadouts(*lo, *editedLoadout) {
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
			return backends.WriteLoadoutWithEncryption(loadoutName, editedLoadout, true)
		}
		return backends.WriteLoadout(loadoutName, editedLoadout)
	}

	// Remove the temp file
	os.Remove(tempFilePath)
	return nil
}
