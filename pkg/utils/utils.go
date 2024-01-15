package utils

import (
	"fmt"
	"time"
)

// Get the current time in RFC3339 format
func GetCurrentTime() string {
	return fmt.Sprintf("%s", time.Now().Format(time.RFC3339))
}
