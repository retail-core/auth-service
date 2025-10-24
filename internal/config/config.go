package config

import (
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    PORT         string 
    DB_SOURCE    string
    JWT_SECRET_KEY string
    MODE         string
}

func LoadConfig() Config {
    godotenv.Load()
    return Config{
        PORT:         getEnv("PORT", "8080"),
        DB_SOURCE:     getEnv("DB_SOURCE", "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable"),
        JWT_SECRET_KEY: getEnv("JWT_SECRET_KEY", "dXAGHVVprhsHaT10d+sdoMbxAa3i4+viSfSSBbKo3pDX0fHUQuc6NfyaSZt+oCxYY7OjpyVb2X+SCY1izcQ8aQQ=="),
        MODE:  getEnv("ENVIRONMENT", "development"),
    }
}

func getEnv(key, defaultVal string) string {
    if val, ok := os.LookupEnv(key); ok {
        return val
    }
    return defaultVal
}
