/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/crypto"
	"github.com/gmherb/envtab/internal/tags"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add LOADOUT_NAME [-v|--encrypt-value] KEY=VALUE [TAG1 TAG2,TAG3 ...]",
	Short: "Add an entry to a envtab loadout",
	Long: `Add an environment variable key-pair as an entry in an envtab
loadout. By default it is cleartext, however, the 
-v|--encrypt-value flag can be used to encrypt the value.
Add tags to your envtab loadout by adding them after the key and value.
Multiple tags can be provided using space or comma as a separator.`,
	Example: `  envtab add myloadout MY_ENV_VAR=myvalue
  envtab add myloadout -v MY_ENV_VAR=myvalue
  envtab add myloadout MY_ENV_VAR=myvalue tag1,tag2,tag3
  envtab add myloadout -v MY_ENV_VAR=myvalue tag1,tag2,tag3
  envtab add myloadout MY_ENV_VAR=myvalue tag1 tag2 tag3
  envtab add myloadout MY_ENV_VAR=myvalue tag1,tag2, tag3 tag4,tag5
  envtab add myloadout MY_ENV_VAR=myvalue -v tag1,tag2, tag3 tag4,tag5`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MinimumNArgs(2),
	Aliases:               []string{"a", "ad"},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("add command called")

		var (
			name    string   // envtab loadout name
			key     string   // Environment variable key
			value   string   // Environment variable value
			newTags []string // Tags to append to envtab loadout
		)

		encryptValue, _ := cmd.Flags().GetBool("encrypt-value")
		encryptFile, _ := cmd.Flags().GetBool("encrypt-file")

		if len(args) == 2 && !strings.Contains(args[1], "=") {
			logger.Debug("No value provided for your envtab entry. No equal sign detected and only 2 args provided.")
			cmd.Usage()
			os.Exit(1)
		}

		name = args[0]
		if strings.Contains(args[1], "=") {
			logger.Debug("Equal sign detected in second argument. Splitting into key and value.")
			key, value = strings.Split(args[1], "=")[0], strings.Split(args[1], "=")[1]
			newTags = args[2:]

		} else {
			logger.Debug("No equal sign detected in second argument. Assigning second argument as key.")
			key = args[1]
			value = args[2]
			newTags = args[3:]
		}

		newTags = tags.SplitTags(newTags)
		newTags = tags.RemoveEmptyTags(newTags)
		newTags = tags.RemoveDuplicateTags(newTags)

		logger.Debug("parsing arguments", "name", name, "key", key, "value", "[REDACTED]", "tags", newTags)

		// Check if loadout exists and determine encryption type
		isFileEncrypted := backends.IsLoadoutFileEncrypted(name)
		hasValueEncrypted := false

		// Try to read existing loadout to check for value-encrypted entries
		lo, readErr := backends.ReadLoadout(name)
		if readErr == nil {
			hasValueEncrypted = backends.HasValueEncryptedEntries(lo)
		}

		// Handle encryption type conflicts
		if isFileEncrypted {
			// Loadout is file-encrypted, must use file encryption
			if !encryptFile {
				// Auto-enable file encryption to preserve existing encryption
				encryptFile = true
			}
			if encryptValue {
				// User tried to use value encryption on file-encrypted loadout
				// Warn that it will be file-encrypted instead
				fmt.Fprintf(os.Stderr, "WARNING: Loadout '%s' is file-encrypted. Value will be stored in file-encrypted format.\n", name)
			}
		} else if hasValueEncrypted && encryptFile {
			// Loadout has value-encrypted entries, but user wants file encryption
			// This is allowed - will convert to file encryption
			fmt.Fprintf(os.Stderr, "WARNING: Converting loadout '%s' from value-encrypted to file-encrypted format.\n", name)
		} else if hasValueEncrypted && encryptValue {
			// Loadout has value-encrypted entries, user wants value encryption - OK
			// No action needed
		}

		var err error
		var finalValue string

		// Determine encryption method for the value
		// Note: If file is file-encrypted, the entire file will be encrypted regardless
		if encryptValue && !isFileEncrypted {
			// Encrypt individual value with SOPS (only if not file-encrypted)
			encrypted, err := crypto.SOPSEncryptValue(value)
			if err != nil {
				fmt.Printf("ERROR: Failed to encrypt value with SOPS: %s\n", err)
				os.Exit(1)
			}
			finalValue = encrypted
		} else {
			finalValue = value
		}

		// Write to loadout (with file-level SOPS encryption if requested or required)
		if encryptFile {
			err = backends.AddEntryToLoadoutWithSOPS(name, key, finalValue, newTags, true)
		} else {
			err = backends.AddEntryToLoadout(name, key, finalValue, newTags)
		}
		if err != nil {
			fmt.Printf("ERROR: Error writing entry to file [%s]: %s\n", name, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolP("encrypt-value", "v", false, "Encrypt individual value with SOPS")
	addCmd.Flags().BoolP("encrypt-file", "f", false, "Encrypt entire loadout file with SOPS")
}
