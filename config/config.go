package config

import "os"

type Config struct {
	PostgresUsername string
	PostgresPassword string
	PostgresHost     string
	PostgresPort     string
	PostgresDatabase string

	TgBotApiToken string
	TgWebhookUrl string
}

func InitConfig() *Config {
	return &Config{
		PostgresUsername: os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
		PostgresDatabase: os.Getenv("POSTGRES_DB"),

		TgBotApiToken: os.Getenv("TG_BOT_API_TOKEN"),
		TgWebhookUrl: os.Getenv("TG_WEBHOOK_URL"),
	}
}
