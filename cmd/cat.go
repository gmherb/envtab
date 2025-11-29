/*
Copyright © 2024 Greg Herbster
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/config"
	"github.com/gmherb/envtab/internal/loadout"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var catOutputPath string
var catDecrypt bool

var catCmd = &cobra.Command{
	Use:   "cat LOADOUT_NAME [LOADOUT_NAME ...]",
	Short: "Concatenate envtab loadouts to stdout",
	Long: `Concatenate envtab loadouts to stdout.
By default, shows encrypted values/files. If the --decrypt flag is provided,
then the values/files will be decrypted and shown in cleartext.`,
	Example: `  envtab cat myloadout
  envtab cat myloadout1 myloadout2 myloadout3
  envtab cat myloadout --decrypt
  envtab cat myloadout --decrypt --output decrypted.yaml`,
	Args:       cobra.MinimumNArgs(1),
	SuggestFor: []string{"print", "display"},
	Aliases:    []string{"c", "ca"},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("cat called with args", "args", args)

		// If --output is set, enforce exactly one loadout and write to file
		if catOutputPath != "" {
			if len(args) != 1 {
				fmt.Fprintf(os.Stderr, "ERROR: when using --output, provide exactly one LOADOUT_NAME\n")
				os.Exit(1)
			}

			data, isFileEncrypted, err := getLoadoutDataForFile(args[0])
			if err != nil {
				os.Exit(1)
			}

			// Ensure parent directory exists if path includes directories
			if dir := filepath.Dir(catOutputPath); dir != "." {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: Failed to create directories for %s: %s\n", catOutputPath, err)
					os.Exit(1)
				}
			}
			if err := os.WriteFile(catOutputPath, data, 0600); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: Failed to write file %s: %s\n", catOutputPath, err)
				os.Exit(1)
			}

			decryptMsg := ""
			if catDecrypt {
				decryptMsg = " (with decrypted values)"
			} else if isFileEncrypted {
				decryptMsg = " (encrypted)"
			}
			fmt.Printf("Wrote loadout [%s] to %s%s\n", args[0], catOutputPath, decryptMsg)
			return
		}

		// Output to stdout for each loadout
		for _, arg := range args {
			printLoadoutToStdout(arg)
		}
	},
}

// getLoadoutDataForFile retrieves loadout data as bytes for file output.
func getLoadoutDataForFile(loadoutName string) ([]byte, bool, error) {
	isFileEncrypted := backends.IsLoadoutFileEncrypted(loadoutName)

	// Handle file-level encrypted loadout without decryption
	if isFileEncrypted && !catDecrypt {
		filePath := filepath.Join(config.InitEnvtab(""), loadoutName+".yaml")
		data, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "ERROR: Loadout %s does not exist\n", loadoutName)
			} else {
				fmt.Fprintf(os.Stderr, "ERROR: Failed to read loadout file: %s\n", err)
			}
			return nil, false, err
		}
		return data, true, nil
	}

	// Read and potentially decrypt loadout
	loadout, err := readLoadoutWithErrorHandling(loadoutName, true)
	if err != nil {
		return nil, false, err
	}

	// Decrypt value-level encrypted entries if --decrypt is set
	if catDecrypt {
		if _, err := loadout.DecryptSOPSValues(); err != nil {
			// DecryptSOPSValues prints warnings for failed decryptions
			// Continue anyway to output what we can
		}
	}

	// Marshal to YAML
	data, err := yaml.Marshal(loadout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to marshal loadout: %s\n", err)
		return nil, false, err
	}
	return data, false, nil
}

// printLoadoutToStdout prints a loadout to stdout, handling encryption appropriately.
func printLoadoutToStdout(loadoutName string) {
	isFileEncrypted := backends.IsLoadoutFileEncrypted(loadoutName)

	// Handle file-level encrypted loadout without decryption
	if isFileEncrypted && !catDecrypt {
		filePath := filepath.Join(config.InitEnvtab(""), loadoutName+".yaml")
		data, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "ERROR: Loadout %s does not exist\n", loadoutName)
			} else {
				fmt.Fprintf(os.Stderr, "ERROR: Failed to read loadout file: %s\n", err)
			}
			return
		}
		os.Stdout.Write(data)
		return
	}

	// Read and potentially decrypt loadout
	loadout, err := readLoadoutWithErrorHandling(loadoutName, false)
	if err != nil {
		return
	}

	// Decrypt value-level encrypted entries if --decrypt is set
	if catDecrypt {
		if _, err := loadout.DecryptSOPSValues(); err != nil {
			// DecryptSOPSValues prints warnings for failed decryptions
			// Continue anyway to print what we can
		}
	}

	loadout.PrintLoadout()
}

// readLoadoutWithErrorHandling reads a loadout with consistent error handling.
// If exitOnError is true, exits on error; otherwise returns error for caller to handle.
func readLoadoutWithErrorHandling(loadoutName string, exitOnError bool) (*loadout.Loadout, error) {
	loadout, err := backends.ReadLoadout(loadoutName)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "SOPS_NOT_INSTALLED") {
			if exitOnError {
				fmt.Fprintf(os.Stderr, "ERROR: Cannot read encrypted loadout without SOPS installed: %s\n", loadoutName)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "WARNING: Skipping loadout %s - SOPS is not installed. Install SOPS to read encrypted loadouts: https://github.com/getsops/sops\n", loadoutName)
			return nil, err
		}
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "ERROR: Loadout %s does not exist\n", loadoutName)
			if exitOnError {
				os.Exit(1)
			}
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		if exitOnError {
			os.Exit(1)
		}
		return nil, err
	}
	return loadout, nil
}

func init() {
	rootCmd.AddCommand(catCmd)
	catCmd.Flags().StringVarP(&catOutputPath, "output", "o", "", "Write loadout YAML to file instead of stdout (only for single loadout)")
	catCmd.Flags().BoolVarP(&catDecrypt, "decrypt", "d", false, "Decrypt file-level and value-level encrypted entries (default: show encrypted values)")
}
