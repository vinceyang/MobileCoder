package config

import (
	"os"
)

type Config struct {
	Port              string
	JWTSecret         string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	SupabaseAPIKey    string
	SupabaseProjectURL string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		JWTSecret:         getEnv("JWT_SECRET", "agentapi-secret-key"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", ""),
		DBName:            getEnv("DB_NAME", "agentapi"),
		SupabaseAPIKey:    getEnv("SUPABASE_API_KEY", ""),
		SupabaseProjectURL: getEnv("SUPABASE_PROJECT_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
