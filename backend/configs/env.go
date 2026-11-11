package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {

	// Try to load .env (only needed for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found, using environment variables")
	}

	githubAppPrivateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	if privateKeyFile := os.Getenv("GITHUB_APP_PRIVATE_KEY_FILE"); privateKeyFile != "" {
		contents, err := os.ReadFile(privateKeyFile)
		if err != nil {
			log.Printf("GitHub App private key file could not be read")
		} else {
			githubAppPrivateKey = string(contents)
		}
	}

	AppConfig = Config{
		AppName: os.Getenv("APP_NAME"),
		Port:    os.Getenv("PORT"),
		Env:     os.Getenv("ENV"),

		DBHost:                   os.Getenv("DB_HOST"),
		DBPort:                   os.Getenv("DB_PORT"),
		DBUser:                   os.Getenv("DB_USER"),
		DBPassword:               os.Getenv("DB_PASSWORD"),
		DBName:                   os.Getenv("DB_NAME"),
		DBSSLMode:                os.Getenv("DB_SSLMODE"),
		JWTSecret:                os.Getenv("JWT_SECRET"),
		GitHubAppSlug:            os.Getenv("GITHUB_APP_SLUG"),
		GitHubAppID:              os.Getenv("GITHUB_APP_ID"),
		GitHubAppPrivateKey:      githubAppPrivateKey,
		GitHubClientID:           os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret:       os.Getenv("GITHUB_CLIENT_SECRET"),
		PublicAppURL:             os.Getenv("PUBLIC_APP_URL"),
		EnvironmentEncryptionKey: os.Getenv("ENVIRONMENT_ENCRYPTION_KEY"),
	}

	log.Println("✅ Environment Loaded")
}
