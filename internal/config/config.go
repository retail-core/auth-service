package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	PORT           string
	DB_SOURCE      string
	JWT_SECRET_KEY string
	MODE           string
	RABBITMQ_URL   string

	// app update check
	APP_UPDATE_CHECK_PLATFORM   string
	APP_UPDATE_CHECK_LATEST_VERSION    string
	APP_UPDATE_CHECK_UPDATE     string
	APP_UPDATE_CHECK_UPDATE_URL string
}

func LoadConfig() Config {
	godotenv.Load()
	return Config{
		PORT:           getEnv("PORT", "8080"),
		DB_SOURCE:      getEnv("DB_SOURCE", "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable"),
		JWT_SECRET_KEY: getEnv("JWT_SECRET_KEY", "mysecretkey"),
		MODE:           getEnv("ENVIRONMENT", "development"),
		RABBITMQ_URL:   getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),

		// app update check
		APP_UPDATE_CHECK_PLATFORM: getEnv("APP_UPDATE_CHECK_PLATFORM", "android"),
		APP_UPDATE_CHECK_LATEST_VERSION:  getEnv("APP_UPDATE_CHECK_LATEST_VERSION", "1.1.1"),
		APP_UPDATE_CHECK_UPDATE:   getEnv("APP_UPDATE_CHECK_UPDATE", "none"), // possible values: none, optional, forced
        APP_UPDATE_CHECK_UPDATE_URL: getEnv("APP_UPDATE_CHECK_UPDATE_URL", "https://play.google.com/store/apps/details?id=com.lynx.jarabiz&pcampaignid=web_share"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
