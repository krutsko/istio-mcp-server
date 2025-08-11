package version

import (
	"testing"
)

// TestBinaryName tests the BinaryName variable
func TestBinaryName(t *testing.T) {
	t.Run("has correct binary name", func(t *testing.T) {
		expected := "istio-mcp-server"
		if BinaryName != expected {
			t.Fatalf("Expected binary name '%s', got '%s'", expected, BinaryName)
		}
	})

	t.Run("binary name is not empty", func(t *testing.T) {
		if BinaryName == "" {
			t.Fatal("Binary name should not be empty")
		}
	})
}

// TestVersion tests the Version variable
func TestVersion(t *testing.T) {
	t.Run("has version", func(t *testing.T) {
		if Version == "" {
			t.Fatal("Version should not be empty")
		}
	})

	t.Run("version has expected format", func(t *testing.T) {
		// Basic check that version looks like a semantic version
		if len(Version) < 5 { // At least "0.0.0"
			t.Fatalf("Version '%s' seems too short", Version)
		}

		// Check it contains dots (basic semantic version check)
		dotCount := 0
		for _, char := range Version {
			if char == '.' {
				dotCount++
			}
		}
		if dotCount < 2 {
			t.Fatalf("Version '%s' doesn't appear to be semantic version format", Version)
		}
	})
}

// TestVariables ensures the variables can be used properly
func TestVariables(t *testing.T) {
	t.Run("variables are properly set", func(t *testing.T) {
		if BinaryName == "" {
			t.Fatal("Binary name should not be empty")
		}
		if Version == "" {
			t.Fatal("Version should not be empty")
		}
		if CommitHash == "" {
			t.Fatal("Commit hash should not be empty")
		}
		if BuildTime == "" {
			t.Fatal("Build time should not be empty")
		}
	})
}
