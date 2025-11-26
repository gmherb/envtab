package utils

import (
	"io"
	"os"
	"testing"
	"time"
)

func TestGetCurrentTime(t *testing.T) {
	currentTime := GetCurrentTime()
	_, err := time.Parse(time.RFC3339, currentTime)
	if err != nil {
		t.Errorf("GetCurrentTime provided invalid time.RFC3339 value: %s", err)
	}

	// Test that it returns different times when called with delay
	// Note: RFC3339 format has second precision, so we need to wait at least 1 second
	time1 := GetCurrentTime()
	time.Sleep(1100 * time.Millisecond)
	time2 := GetCurrentTime()
	if time1 == time2 {
		t.Error("GetCurrentTime() should return different values when called at different times")
	}
}

func TestPromptForAnswer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"yes with y", "y\n", true},
		{"yes with yes", "yes\n", true},
		{"yes with YES", "YES\n", true},
		{"yes with Yes", "Yes\n", true},
		{"no with n", "n\n", false},
		{"no with no", "no\n", false},
		{"no with NO", "NO\n", false},
		{"no with No", "No\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original stdin
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			// Create a pipe to simulate user input
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdin = r

			// Write input in a goroutine
			go func() {
				defer w.Close()
				io.WriteString(w, tt.input)
			}()

			// Call the function
			result := PromptForAnswer("Test prompt")

			// Close the read end
			r.Close()

			if result != tt.expected {
				t.Errorf("PromptForAnswer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPromptForAnswerRetry(t *testing.T) {
	// Test that it retries on invalid input
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r

	// Write invalid input followed by valid input
	go func() {
		defer w.Close()
		io.WriteString(w, "invalid\n")
		io.WriteString(w, "maybe\n")
		io.WriteString(w, "y\n")
	}()

	// This should eventually return true after retries
	result := PromptForAnswer("Test prompt")
	r.Close()

	if result != true {
		t.Errorf("PromptForAnswer() should eventually accept 'y', got %v", result)
	}
}

func TestGetCurrentTimeFormat(t *testing.T) {
	// Test that the format is consistent
	for i := 0; i < 10; i++ {
		timeStr := GetCurrentTime()
		_, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			t.Errorf("GetCurrentTime() returned invalid RFC3339 format: %s, error: %v", timeStr, err)
		}
	}
}
