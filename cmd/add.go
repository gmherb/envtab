/*
Copyright © 2024 Greg Herbster
*/
package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/tags"
	"github.com/gmherb/envtab/pkg/sops"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add LOADOUT_NAME [-e|--encrypt-value] [-f|--encrypt-file] KEY=VALUE [TAG1 TAG2,TAG3 ...]",
	Short: "Add an entry to a envtab loadout",
	Long: `Add an environment variable key-pair as an entry in an envtab
loadout. By default it is cleartext, however, the 
-e|--encrypt-value flag can be used to encrypt the value.
Add tags to your envtab loadout by adding them after the key and value.
Multiple tags can be provided using space or comma as a separator.`,
	Example: `  envtab add myloadout MY_ENV_VAR=myvalue
  envtab add myloadout -e MY_ENV_VAR=myvalue
  envtab add myloadout MY_ENV_VAR=myvalue tag1,tag2,tag3
  envtab add myloadout -e MY_ENV_VAR=myvalue tag1,tag2,tag3
  envtab add myloadout MY_ENV_VAR=myvalue tag1 tag2 tag3
  envtab add myloadout MY_ENV_VAR=myvalue tag1,tag2, tag3 tag4,tag5
  envtab add myloadout MY_ENV_VAR=myvalue -e tag1,tag2, tag3 tag4,tag5`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MinimumNArgs(2),
	Aliases:               []string{"a", "ad"},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("add command called")

		encryptValue, _ := cmd.Flags().GetBool("encrypt-value")
		encryptFile, _ := cmd.Flags().GetBool("encrypt-file")

		if len(args) == 2 && !strings.Contains(args[1], "=") {
			slog.Debug("No value provided for envtab entry. No equal sign detected and only 2 args provided.")
			cmd.Usage()
			os.Exit(1)
		}

		// Parse arguments: name, key, value, and tags
		name := args[0]
		var key, value string
		var newTags []string

		if strings.Contains(args[1], "=") {
			slog.Debug("Equal sign detected in second argument. Splitting into key and value.")
			parts := strings.SplitN(args[1], "=", 2)
			key, value = parts[0], parts[1]
			newTags = args[2:]
		} else {
			slog.Debug("No equal sign detected in second argument. Assigning second argument as key.")
			key = args[1]
			value = args[2]
			newTags = args[3:]
		}

		// Process tags
		newTags = tags.RemoveDuplicateTags(tags.RemoveEmptyTags(tags.SplitTags(newTags)))

		slog.Debug("parsing arguments", "name", name, "key", key, "value", "[REDACTED]", "tags", newTags)

		// Check if loadout exists and determine encryption type
		isFileEncrypted := backends.IsLoadoutFileEncrypted(name)
		hasValueEncrypted := false

		if lo, readErr := backends.ReadLoadout(name); readErr == nil {
			hasValueEncrypted = backends.HasValueEncryptedEntries(lo)
		}

		// Handle encryption type conflicts
		if isFileEncrypted {
			if !encryptFile {
				encryptFile = true
			}
			if encryptValue {
				slog.Warn("loadout is file-encrypted, value will be stored in file-encrypted format", "loadout", name)
			}
		} else if hasValueEncrypted && encryptFile {
			slog.Warn("converting loadout from value-encrypted to file-encrypted format", "loadout", name)
		}

		// Encrypt value if requested (only if not file-encrypted)
		finalValue := value
		if encryptValue && !isFileEncrypted {
			encrypted, err := sops.SOPSEncryptValue(value)
			if err != nil {
				slog.Error("failure encrypting value with SOPS", "error", err)
				os.Exit(1)
			}
			finalValue = encrypted
		}

		// Write to loadout
		var err error
		if encryptFile {
			err = backends.AddEntryToLoadoutWithSOPS(name, key, finalValue, newTags, true)
		} else {
			err = backends.AddEntryToLoadout(name, key, finalValue, newTags)
		}
		if err != nil {
			slog.Error("failure writing entry to loadout", "loadout", name, "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolP("encrypt-value", "e", false, "Encrypt individual value with SOPS")
	addCmd.Flags().BoolP("encrypt-file", "f", false, "Encrypt entire loadout file with SOPS")
}
