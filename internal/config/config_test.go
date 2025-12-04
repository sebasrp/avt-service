package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_EmailConfig(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    EmailConfig
	}{
		{
			name: "loads email config with all values set",
			envVars: map[string]string{
				"EMAIL_PROVIDER":     "mailgun",
				"MAILGUN_DOMAIN":     "mg.example.com",
				"MAILGUN_API_KEY":    "key-123",
				"EMAIL_FROM_ADDRESS": "support@example.com",
				"EMAIL_FROM_NAME":    "Example Support",
				"APP_URL":            "https://app.example.com",
				"RESET_TOKEN_TTL":    "6h",
			},
			want: EmailConfig{
				Provider:      "mailgun",
				MailgunDomain: "mg.example.com",
				MailgunAPIKey: "key-123",
				FromAddress:   "support@example.com",
				FromName:      "Example Support",
				AppURL:        "https://app.example.com",
				ResetTokenTTL: 6 * time.Hour,
			},
		},
		{
			name:    "loads email config with defaults",
			envVars: map[string]string{},
			want: EmailConfig{
				Provider:      "mock",
				MailgunDomain: "",
				MailgunAPIKey: "",
				FromAddress:   "noreply@example.com",
				FromName:      "AVT Service",
				AppURL:        "http://localhost:3000",
				ResetTokenTTL: 12 * time.Hour,
			},
		},
		{
			name: "loads email config with mock provider",
			envVars: map[string]string{
				"EMAIL_PROVIDER": "mock",
			},
			want: EmailConfig{
				Provider:      "mock",
				MailgunDomain: "",
				MailgunAPIKey: "",
				FromAddress:   "noreply@example.com",
				FromName:      "AVT Service",
				AppURL:        "http://localhost:3000",
				ResetTokenTTL: 12 * time.Hour,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			cleanEmailEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Check EmailConfig
			if cfg.Email.Provider != tt.want.Provider {
				t.Errorf("Email.Provider = %q, want %q", cfg.Email.Provider, tt.want.Provider)
			}
			if cfg.Email.MailgunDomain != tt.want.MailgunDomain {
				t.Errorf("Email.MailgunDomain = %q, want %q", cfg.Email.MailgunDomain, tt.want.MailgunDomain)
			}
			if cfg.Email.MailgunAPIKey != tt.want.MailgunAPIKey {
				t.Errorf("Email.MailgunAPIKey = %q, want %q", cfg.Email.MailgunAPIKey, tt.want.MailgunAPIKey)
			}
			if cfg.Email.FromAddress != tt.want.FromAddress {
				t.Errorf("Email.FromAddress = %q, want %q", cfg.Email.FromAddress, tt.want.FromAddress)
			}
			if cfg.Email.FromName != tt.want.FromName {
				t.Errorf("Email.FromName = %q, want %q", cfg.Email.FromName, tt.want.FromName)
			}
			if cfg.Email.AppURL != tt.want.AppURL {
				t.Errorf("Email.AppURL = %q, want %q", cfg.Email.AppURL, tt.want.AppURL)
			}
			if cfg.Email.ResetTokenTTL != tt.want.ResetTokenTTL {
				t.Errorf("Email.ResetTokenTTL = %v, want %v", cfg.Email.ResetTokenTTL, tt.want.ResetTokenTTL)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with mailgun provider",
			config: Config{
				Email: EmailConfig{
					Provider:      "mailgun",
					MailgunDomain: "mg.example.com",
					MailgunAPIKey: "key-123",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with mock provider",
			config: Config{
				Email: EmailConfig{
					Provider: "mock",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid - mailgun provider without API key",
			config: Config{
				Email: EmailConfig{
					Provider:      "mailgun",
					MailgunDomain: "mg.example.com",
					MailgunAPIKey: "",
				},
			},
			wantErr: true,
			errMsg:  "MAILGUN_API_KEY is required when EMAIL_PROVIDER=mailgun",
		},
		{
			name: "invalid - mailgun provider without domain",
			config: Config{
				Email: EmailConfig{
					Provider:      "mailgun",
					MailgunDomain: "",
					MailgunAPIKey: "key-123",
				},
			},
			wantErr: true,
			errMsg:  "MAILGUN_DOMAIN is required when EMAIL_PROVIDER=mailgun",
		},
		{
			name: "invalid - mailgun provider without both",
			config: Config{
				Email: EmailConfig{
					Provider:      "mailgun",
					MailgunDomain: "",
					MailgunAPIKey: "",
				},
			},
			wantErr: true,
			errMsg:  "MAILGUN_API_KEY is required when EMAIL_PROVIDER=mailgun",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLoad_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "fails validation when mailgun provider missing API key",
			envVars: map[string]string{
				"EMAIL_PROVIDER":  "mailgun",
				"MAILGUN_DOMAIN":  "mg.example.com",
				"MAILGUN_API_KEY": "",
			},
			wantErr: true,
			errMsg:  "MAILGUN_API_KEY is required when EMAIL_PROVIDER=mailgun",
		},
		{
			name: "fails validation when mailgun provider missing domain",
			envVars: map[string]string{
				"EMAIL_PROVIDER":  "mailgun",
				"MAILGUN_API_KEY": "key-123",
				"MAILGUN_DOMAIN":  "",
			},
			wantErr: true,
			errMsg:  "MAILGUN_DOMAIN is required when EMAIL_PROVIDER=mailgun",
		},
		{
			name: "succeeds with mock provider and no mailgun credentials",
			envVars: map[string]string{
				"EMAIL_PROVIDER": "mock",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			cleanEmailEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			_, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Load() error message = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLoad_JWTSecretUsesGetSecret(t *testing.T) {
	// Clean environment
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_SECRET_FILE")

	// Test with direct env var
	os.Setenv("JWT_SECRET", "direct-secret")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Auth.JWTSecret != "direct-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.Auth.JWTSecret, "direct-secret")
	}
}

// cleanEmailEnv removes all email-related environment variables
func cleanEmailEnv() {
	envVars := []string{
		"EMAIL_PROVIDER",
		"MAILGUN_DOMAIN",
		"MAILGUN_DOMAIN_FILE",
		"MAILGUN_API_KEY",
		"MAILGUN_API_KEY_FILE",
		"EMAIL_FROM_ADDRESS",
		"EMAIL_FROM_NAME",
		"APP_URL",
		"RESET_TOKEN_TTL",
	}
	for _, key := range envVars {
		os.Unsetenv(key)
	}
}
