/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"

	"github.com/gmherb/envtab/internal/backends"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var catOutputPath string

var catCmd = &cobra.Command{
	Use:   "cat LOADOUT_NAME [LOADOUT_NAME ...]",
	Short: "Concatenate envtab loadouts to stdout",
	Long:  `Concatenate envtab loadouts to stdout.`,
	Example: `  envtab cat myloadout
  envtab cat myloadout1 myloadout2 myloadout3`,
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"c", "ca", "print"},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("cat called")

		// If --output is set, enforce exactly one loadout and write to file
		if catOutputPath != "" {
			if len(args) != 1 {
				fmt.Fprintf(os.Stderr, "ERROR: when using --output, provide exactly one LOADOUT_NAME\n")
				os.Exit(1)
			}
			loadout, err := backends.ReadLoadout(args[0])
			if err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "SOPS_NOT_INSTALLED") {
					fmt.Fprintf(os.Stderr, "ERROR: Cannot read encrypted loadout without SOPS installed: %s\n", args[0])
					os.Exit(1)
				}
				if os.IsNotExist(err) {
					fmt.Fprintf(os.Stderr, "ERROR: Loadout %s does not exist\n", args[0])
					os.Exit(1)
				}
				fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
				os.Exit(1)
			}
			data, err := yaml.Marshal(loadout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: Failed to marshal loadout: %s\n", err)
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
			fmt.Printf("Wrote loadout [%s] to %s\n", args[0], catOutputPath)
			return
		}

		for _, arg := range args {

			loadout, err := backends.ReadLoadout(arg)
			if err != nil {
				// Skip loadout if SOPS is not installed (for encrypted loadouts)
				errStr := err.Error()
				if strings.Contains(errStr, "SOPS_NOT_INSTALLED") {
					fmt.Fprintf(os.Stderr, "WARNING: Skipping loadout %s - SOPS is not installed. Install SOPS to read encrypted loadouts: https://github.com/getsops/sops\n", arg)
					continue
				}
				if os.IsNotExist(err) {
					fmt.Printf("ERROR: Loadout %s does not exist\n", arg)
					continue
				}
				fmt.Printf("Error reading loadout %s: %s\n", arg, err)
				os.Exit(1)
			}

			loadout.PrintLoadout()
		}
	},
}

func init() {
	rootCmd.AddCommand(catCmd)
	catCmd.Flags().StringVarP(&catOutputPath, "output", "o", "", "Write loadout YAML to file instead of stdout (only for single loadout)")
}
