package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load .env")
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
	}

	log.Println("✅ Environment Loaded")
}
