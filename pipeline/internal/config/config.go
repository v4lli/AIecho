package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

type Config struct {
	CloudflareAPIKey    string
	CloudflareAccountID string
	ImageToTextModel    string
	LargeLanguageModel  string
}

func LoadConfig(filename string) *Config {
	err := godotenv.Load(filename)
	if err != nil {
		log.Printf("Error loading .env file %v", err)
	}
	cloudflareAPIKey := loadEnvVar("CLOUDFLARE_API_KEY")
	cloudflareAccountID := loadEnvVar("CLOUDFLARE_ACCOUNT_ID")
	imageToTextModel := loadEnvVar("IMAGE_TO_TEXT_MODEL")
	largeLanguageModel := loadEnvVar("LARGE_LANGUAGE_MODEL")
	if cloudflareAPIKey == "" || cloudflareAccountID == "" || imageToTextModel == "" || largeLanguageModel == "" {
		log.Fatalf("Empty environment value for required field")
	}
	return &Config{
		CloudflareAPIKey:    cloudflareAPIKey,
		CloudflareAccountID: cloudflareAccountID,
		ImageToTextModel:    imageToTextModel,
		LargeLanguageModel:  largeLanguageModel,
	}
}

func (c *Config) GenerateHeader() http.Header {
	header := make(http.Header)
	header.Set("Authorization", "Bearer "+c.CloudflareAPIKey)
	header.Set("Content-Type", "application/json")
	return header
}

func (c *Config) GenerateI2TURL() string {
	return fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s", c.CloudflareAccountID, c.ImageToTextModel,
	)
}

func (c *Config) GenerateLLMURL() string {
	return fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s", c.CloudflareAccountID, c.LargeLanguageModel,
	)
}

func loadEnvVar(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("environment variable %s not set", key)
	}
	return value
}
