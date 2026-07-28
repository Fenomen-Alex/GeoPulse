package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	AllowedOrigin      string
	JWTSecret          string
	TursoDatabaseURL   string
	TursoAuthToken     string
	GoogleClientID     string
	GoogleClientSecret string
	ORSAPIKey          string
}

func LoadConfig() (*Config, error) {
	// Attempt to load .env from root directory if running locally
	_ = godotenv.Load("../.env") // Looks up one directory from server/

	return &Config{
		Port:               getEnv("PORT", "8080"),
		AllowedOrigin:      getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
		JWTSecret:          getEnv("JWT_SECRET", "default-dev-secret"),
		TursoDatabaseURL:   getEnv("TURSO_DATABASE_URL", ""),
		TursoAuthToken:     getEnv("TURSO_AUTH_TOKEN", ""),
		GoogleClientID:     getEnv("VITE_GOOGLE_CLIENT_ID", ""), // Reuses VITE_ variable
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		ORSAPIKey:          getEnv("ORS_API_KEY", ""),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
