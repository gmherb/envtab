package utils

import (
	"testing"
	"time"
)

func TestGetCurrentTime(t *testing.T) {
	currentTime := GetCurrentTime()
	_, err := time.Parse(time.RFC3339, currentTime)
	if err != nil {
		t.Errorf("GetCurrentTime provided invalid time.RFC3339 value: %s", err)
	}
}
