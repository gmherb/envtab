package utils

import (
	"fmt"
	"time"
)

// Get the current time in RFC3339 format
func GetCurrentTime() string {
	return fmt.Sprintf("%s", time.Now().Format(time.RFC3339))
}

// Pad a string with spaces to a given length
func PadString(s string, length int) string {
	return fmt.Sprintf("%-*s", length, s)
}
