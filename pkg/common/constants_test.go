package common

import (
	"testing"
	"time"
)

func TestConstants(t *testing.T) {
	// Check that DefaultTimeout is properly set
	if DefaultTimeout != 30*time.Second {
		t.Errorf("DefaultTimeout should be 30s, got %v", DefaultTimeout)
	}

	// Check that UserAgent is properly set
	expectedUserAgent := "lsweb/1.0"
	if UserAgent != expectedUserAgent {
		t.Errorf("UserAgent should be %s, got %s", expectedUserAgent, UserAgent)
	}

	// Check that MaxContentSize is properly set
	expectedSize := 10 * 1024 * 1024 // 10MB
	if MaxContentSize != expectedSize {
		t.Errorf("MaxContentSize should be %d, got %d", expectedSize, MaxContentSize)
	}

	// Check Version is properly set
	if Version == "" {
		t.Error("Version should not be empty")
	}
}
