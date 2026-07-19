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

	AppConfig = Config{
		AppName: os.Getenv("APP_NAME"),
		Port:    os.Getenv("PORT"),
		Env:     os.Getenv("ENV"),

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
	}

	log.Println("✅ Environment Loaded")
}
