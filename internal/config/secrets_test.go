package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetSecret(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		envValue     string
		fileEnvVar   string
		fileContent  string
		defaultValue string
		want         string
	}{
		{
			name:         "returns direct environment variable when set",
			envVar:       "TEST_SECRET",
			envValue:     "direct-value",
			defaultValue: "default",
			want:         "direct-value",
		},
		{
			name:         "returns default when no env var or file",
			envVar:       "TEST_SECRET",
			defaultValue: "default-value",
			want:         "default-value",
		},
		{
			name:         "returns empty string when default is empty",
			envVar:       "TEST_SECRET",
			defaultValue: "",
			want:         "",
		},
		{
			name:         "reads from file when _FILE env var is set",
			envVar:       "TEST_SECRET",
			fileEnvVar:   "TEST_SECRET_FILE",
			fileContent:  "file-content",
			defaultValue: "default",
			want:         "file-content",
		},
		{
			name:         "trims whitespace from file content",
			envVar:       "TEST_SECRET",
			fileEnvVar:   "TEST_SECRET_FILE",
			fileContent:  "  file-content\n\t",
			defaultValue: "default",
			want:         "file-content",
		},
		{
			name:         "prefers direct env var over file",
			envVar:       "TEST_SECRET",
			envValue:     "direct-value",
			fileEnvVar:   "TEST_SECRET_FILE",
			fileContent:  "file-content",
			defaultValue: "default",
			want:         "direct-value",
		},
		{
			name:         "returns default when file doesn't exist",
			envVar:       "TEST_SECRET",
			fileEnvVar:   "TEST_SECRET_FILE",
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv(tt.envVar)
			os.Unsetenv(tt.fileEnvVar)

			// Set up direct environment variable if specified
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}

			// Set up file-based secret if specified
			if tt.fileContent != "" {
				// Create temporary file
				tmpDir := t.TempDir()
				tmpFile := filepath.Join(tmpDir, "secret")
				if err := os.WriteFile(tmpFile, []byte(tt.fileContent), 0600); err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}

				// Set file path env var
				os.Setenv(tt.fileEnvVar, tmpFile)
				defer os.Unsetenv(tt.fileEnvVar)
			} else if tt.fileEnvVar != "" && tt.name == "returns default when file doesn't exist" {
				// Set file env var to non-existent path
				os.Setenv(tt.fileEnvVar, "/nonexistent/path/to/secret")
				defer os.Unsetenv(tt.fileEnvVar)
			}

			// Test GetSecret
			got := GetSecret(tt.envVar, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetSecret() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetSecret_RealWorldScenarios(t *testing.T) {
	t.Run("Docker secrets scenario", func(t *testing.T) {
		// Simulate Docker secrets directory
		tmpDir := t.TempDir()
		secretFile := filepath.Join(tmpDir, "mailgun_api_key")
		secretContent := "key-abc123xyz789"

		if err := os.WriteFile(secretFile, []byte(secretContent), 0600); err != nil {
			t.Fatalf("failed to create secret file: %v", err)
		}

		// Set the _FILE env var
		os.Setenv("MAILGUN_API_KEY_FILE", secretFile)
		defer os.Unsetenv("MAILGUN_API_KEY_FILE")

		got := GetSecret("MAILGUN_API_KEY", "")
		if got != secretContent {
			t.Errorf("GetSecret() = %q, want %q", got, secretContent)
		}
	})

	t.Run("Environment variable scenario", func(t *testing.T) {
		// Simulate environment variable
		secretValue := "env-var-secret-123" // #nosec G101 - test value, not a real credential
		os.Setenv("JWT_SECRET", secretValue)
		defer os.Unsetenv("JWT_SECRET")

		got := GetSecret("JWT_SECRET", "default-secret")
		if got != secretValue {
			t.Errorf("GetSecret() = %q, want %q", got, secretValue)
		}
	})

	t.Run("Fallback to default scenario", func(t *testing.T) {
		// No env var or file set
		defaultValue := "dev-mode-secret"

		got := GetSecret("SOME_SECRET", defaultValue)
		if got != defaultValue {
			t.Errorf("GetSecret() = %q, want %q", got, defaultValue)
		}
	})
}
