package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/gmherb/envtab/internal/backends"
	"github.com/gmherb/envtab/internal/env"
	"github.com/gmherb/envtab/internal/loadout"
	"github.com/gmherb/envtab/internal/sops"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var showCmd = &cobra.Command{
	Use:   "show [LOADOUT_PATTERN...]",
	Short: "Show active loadouts",
	Long: `Show each loadout with active entries (environment variables).
Optional glob patterns can be provided to filter results.
If multiple patterns are provided, loadouts matching any pattern will be shown.`,
	Args:                  cobra.ArbitraryArgs,
	SuggestFor:            []string{"status"},
	Aliases:               []string{"s", "sh", "sho"},
	DisableFlagsInUseLine: true,
	Example: `  envtab show
  envtab show aws\*
  envtab show production
  envtab show aws\* \*gcp\*`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("show called with args", "args", args)
		decrypt, _ := cmd.Flags().GetBool("decrypt")
		all, _ := cmd.Flags().GetBool("all")
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		slog.Debug("show called with args", "decrypt", decrypt, "all", all, "key", key, "value", value, "patterns", args)

		envtabSlice, err := backends.ListLoadouts()
		if err != nil {
			slog.Error("failure listing loadouts", "error", err)
			os.Exit(1)
		}

		environment := env.NewEnv()
		environment.Populate()

		waitGroup := sync.WaitGroup{}
		ch := make(chan []string, len(envtabSlice))
		for _, loadout := range envtabSlice {
			waitGroup.Add(1)
			go ShowLoadout(loadout, environment, decrypt, key, value, all, args, &waitGroup, ch)
		}
		func() {
			waitGroup.Wait()
			close(ch)
		}()
		for entryString := range ch {
			for _, entry := range entryString {
				fmt.Println(entry)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().BoolP("decrypt", "d", false, "Show sensitive values (decrypt SOPS encrypted values)")
	showCmd.Flags().BoolP("all", "a", false, "Show all envtab entries")
	showCmd.Flags().StringP("key", "k", "", "Show env var matching key")
	showCmd.Flags().StringP("value", "v", "", "Show env var matching value")
	showCmd.MarkFlagsMutuallyExclusive("all", "key", "value")
}

func ShowLoadout(lo string, environment *env.Env, decrypt bool, keyFilter string, valueFilter string, all bool, patterns []string, waitGroup *sync.WaitGroup, ch chan []string) {
	defer waitGroup.Done()

	var entries = []string{}
	// Filter by patterns if provided
	if len(patterns) > 0 {
		matched := false
		for _, pattern := range patterns {
			m, _ := filepath.Match(pattern, lo)
			if m {
				matched = true
				break
			}
		}
		if !matched {
			return
		}
	}

	loStruct, err := backends.ReadLoadout(lo)
	if err != nil {
		// Skip loadout if SOPS is not installed (for encrypted loadouts)
		errStr := err.Error()
		if strings.Contains(errStr, "SOPS_NOT_INSTALLED") {
			slog.Warn("skipping loadout - SOPS not installed", "loadout", lo)
			return
		}
		slog.Error("failure reading loadout", "loadout", lo, "error", err)
		return
	}

	activeMap := make(map[string]bool)
	keyMap := make(map[string]bool)
	valueMap := make(map[string]bool)

	for entryKey, entryValue := range loStruct.Entries {

		// Get decrypted value for filtering/comparison
		decryptedValue := sops.SOPSDisplayValue(entryValue, true)
		// Expand variables for filtering/comparison (if not encrypted)
		if !strings.HasPrefix(decryptedValue, "SOPS:") {
			decryptedValue = loadout.ExpandVariables(decryptedValue)
		}

		if keyFilter != "" {
			if entryKey == keyFilter {
				keyMap[entryKey] = true
			}
		} else if valueFilter != "" {
			if decryptedValue == valueFilter {
				valueMap[entryKey] = true
			}
		} else if environment.IsEntryActive(entryKey, entryValue) {
			activeMap[entryKey] = true
		}

		if keyMap[entryKey] || valueMap[entryKey] || activeMap[entryKey] || all {
			// Display the original value with variable references (e.g., $HOME, $PATH)
			// Don't expand variables in the display - show them as stored
			displayValue := sops.SOPSDisplayValue(entryValue, decrypt)
			entries = append(entries, entryKey+"="+displayValue)
		}

	}

	// TODO: Add different color pattern/options
	entryColor := color.New(color.FgHiWhite).SprintFunc()
	dashColor := color.New(color.FgHiBlack).SprintFunc()
	loColor := color.New(color.FgGreen).SprintFunc()
	countColor := color.New(color.FgBlue).SprintFunc()
	padding := "   "
	var termWidth int

	// Check viper first (supports ENVTAB_TERM_WIDTH env var and config file)
	if viper.IsSet("term.width") {
		termWidth = viper.GetInt("term.width")
		slog.Debug("ENVTAB_TERM_WIDTH set to", "termWidth", termWidth)
		if termWidth <= 0 {
			slog.Warn("ENVTAB_TERM_WIDTH must be positive, using default of 80", "termWidth", termWidth)
			termWidth = 80
		}
	} else {
		termWidth, _, err = term.GetSize(int(os.Stdout.Fd()))
		slog.Debug("terminal width", "termWidth", termWidth)
		if err != nil {
			slog.Warn("failure getting terminal width, using default of 80", "error", err, "termWidth", termWidth)
			termWidth = 80
		}
	}

	// If a loadout has fewer active entries than total entries, colorize the count red
	if len(loStruct.Entries) != len(activeMap) {
		countColor = color.New(color.FgRed).SprintFunc()
	}

	countLeftHandSide := len(loStruct.Entries)
	var countRightHandSide int = 0
	if len(keyMap) > 0 {
		countRightHandSide = len(keyMap)
	} else if len(valueMap) > 0 {
		countRightHandSide = len(valueMap)
	} else {
		countRightHandSide = len(activeMap)
	}

	entryString := []string{}
	if len(entries) > 0 {
		dashCount := termWidth - len(lo) -
			len(fmt.Sprint(countLeftHandSide)) -
			len(fmt.Sprint(countRightHandSide)) -
			10 // magic number

		entryString = append(entryString,
			loColor(lo)+" "+strings.Repeat(dashColor("-"), dashCount)+" [ "+countColor(countLeftHandSide)+" / "+countColor(countRightHandSide)+" ]",
		)
		for _, entry := range entries {
			entryString = append(entryString, fmt.Sprint(padding, entryColor(entry)))
		}
	}
	ch <- entryString
}
