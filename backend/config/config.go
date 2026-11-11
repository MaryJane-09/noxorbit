package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	ServerPort       string
	GmailAddress     string
	GmailAppPassword string
	DatabaseURL      string
}

var AppConfig = Config{
	ServerPort:       ":8080",
	GmailAddress:     os.Getenv("GMAIL_ADDRESS"),
	GmailAppPassword: os.Getenv("GMAIL_APP_PASSWORD"),
	DatabaseURL:      os.Getenv("DATABASE_URL"),
}
