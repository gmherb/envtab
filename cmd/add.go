/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/gmherb/envtab/internal/crypto"
	"github.com/gmherb/envtab/internal/envtab"
	"github.com/gmherb/envtab/internal/tags"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add LOADOUT_NAME [-s|--sensitive] KEY=VALUE [TAG1 TAG2,TAG3 ...]",
	Short: "Add an entry to a envtab loadout",
	Long: `Add an environment variable key-pair as an entry in an envtab
loadout. By default it is cleartext, however, the 
-s|--sensitive flag can be used to encrypt the value.
Add tags to your envtab loadout by adding them after the key and value.
Multiple tags can be provided using space or comma as a separator.`,
	Example: `  envtab add myloadout MY_ENV_VAR=myvalue
  envtab add myloadout -s MY_ENV_VAR=myvalue
  envtab add myloadout MY_ENV_VAR=myvalue tag1,tag2,tag3
  envtab add myloadout -s MY_ENV_VAR=myvalue tag1,tag2,tag3
  envtab add myloadout MY_ENV_VAR=myvalue tag1 tag2 tag3
  envtab add myloadout MY_ENV_VAR=myvalue tag1,tag2, tag3 tag4,tag5
  envtab add myloadout MY_ENV_VAR=myvalue -s tag1,tag2, tag3 tag4,tag5`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.MinimumNArgs(2),
	Aliases:               []string{"a", "ad"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DEBUG: add command called")

		var (
			name    string   // envtab loadout name
			key     string   // Environment variable key
			value   string   // Environment variable value
			newTags []string // Tags to append to envtab loadout
		)

		const EncryptedPrefix = "ENC:"
		const SOPSValuePrefix = "SOPS:"
		gcpKeyName := os.Getenv("ENVTAB_GCP_KMS_KEY")
		useSOPS := os.Getenv("ENVTAB_USE_SOPS") == "true"

		sensitive, _ := cmd.Flags().GetBool("sensitive")
		sopsValue, _ := cmd.Flags().GetBool("sops-value")
		sopsFile, _ := cmd.Flags().GetBool("sops-file")

		if len(args) == 2 && !strings.Contains(args[1], "=") {
			fmt.Println("DEBUG: No value provided for your envtab entry. No equal sign detected and only 2 args provided.")
			cmd.Usage()
			os.Exit(1)
		}

		name = args[0]
		if strings.Contains(args[1], "=") {
			fmt.Println("DEBUG: Equal sign detected in second argument. Splitting into key and value.")
			key, value = strings.Split(args[1], "=")[0], strings.Split(args[1], "=")[1]
			newTags = args[2:]

		} else {
			fmt.Println("DEBUG: No equal sign detected in second argument. Assigning second argument as key.")
			key = args[1]
			value = args[2]
			newTags = args[3:]
		}

		newTags = tags.SplitTags(newTags)
		newTags = tags.RemoveEmptyTags(newTags)
		newTags = tags.RemoveDuplicateTags(newTags)

		fmt.Printf("DEBUG: Name: %s, Key: %s, Value: %s, tags: %s.\n", name, key, value, newTags)

		var err error
		var finalValue string

		// Determine encryption method
		if sopsValue {
			// Encrypt individual value with SOPS
			encryptedValue, err := crypto.SOPSEncryptValue(value)
			if err != nil {
				fmt.Printf("ERROR: Failed to encrypt value with SOPS: %s\n", err)
				os.Exit(1)
			}
			finalValue = encryptedValue
		} else if sensitive && gcpKeyName != "" {
			// Use GCP KMS encryption (existing behavior)
			ciphertext := crypto.GcpKmsEncrypt(gcpKeyName, value)
			finalValue = EncryptedPrefix + base64.StdEncoding.EncodeToString(ciphertext)
		} else if sensitive {
			// If sensitive but no GCP key, try SOPS value encryption
			encryptedValue, err := crypto.SOPSEncryptValue(value)
			if err != nil {
				fmt.Printf("ERROR: Failed to encrypt value with SOPS: %s\n", err)
				os.Exit(1)
			}
			finalValue = encryptedValue
		} else {
			finalValue = value
		}

		// Write to loadout (with file-level SOPS encryption if requested)
		if sopsFile || useSOPS {
			err = envtab.AddEntryToLoadoutWithSOPS(name, key, finalValue, newTags, true)
		} else {
			err = envtab.AddEntryToLoadout(name, key, finalValue, newTags)
		}
		if err != nil {
			fmt.Printf("ERROR: Error writing entry to file [%s]: %s\n", name, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolP("sensitive", "s", false, "Add sensitive value (encrypted based on settings)")
	addCmd.Flags().Bool("sops-value", false, "Encrypt individual value with SOPS")
	addCmd.Flags().Bool("sops-file", false, "Encrypt entire loadout file with SOPS")
}
