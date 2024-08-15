package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tables := []struct {
		name           string
		envFile        string
		envVars        map[string]string
		expectedResult *Config
	}{
		{
			name:    "All Variables Provided",
			envFile: ".env.test",
			envVars: map[string]string{
				"CLOUDFLARE_API_KEY":    "TestKey",
				"CLOUDFLARE_ACCOUNT_ID": "TestID",
				"IMAGE_TO_TEXT_MODEL":   "TestImageModel",
				"LARGE_LANGUAGE_MODEL":  "TestLanguageModel",
			},
			expectedResult: &Config{
				CloudflareAPIKey:    "TestKey",
				CloudflareAccountID: "TestID",
				ImageToTextModel:    "TestImageModel",
				LargeLanguageModel:  "TestLanguageModel",
			},
		},
		{
			name:    "Variables Not Provided",
			envFile: ".env.empty",
			envVars: map[string]string{},
			expectedResult: &Config{
				CloudflareAPIKey:    "",
				CloudflareAccountID: "",
				ImageToTextModel:    "",
				LargeLanguageModel:  "",
			},
		},
	}

	for _, table := range tables {
		t.Run(
			table.name, func(t *testing.T) {
				// setup
				os.Setenv("ENV_PATH", table.envFile)
				for k, v := range table.envVars {
					os.Setenv(k, v)
				}

				config := LoadConfig(os.Getenv("ENV_PATH"))

				assert.Equal(t, table.expectedResult, config)

				// teardown
				os.Setenv("ENV_PATH", "")
				for k := range table.envVars {
					os.Unsetenv(k)
				}
			},
		)
	}
}
