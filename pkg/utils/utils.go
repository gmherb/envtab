package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Get the current time in RFC3339 format
func GetCurrentTime() string {
	return fmt.Sprintf("%s", time.Now().Format(time.RFC3339))
}

// Pad a string on Right
func PadString(s string, length int) string {
	return fmt.Sprintf("%-*s", length, s)
}

func PromptForAnswer(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
