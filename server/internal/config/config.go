package config

import (
	"os"

	"github.com/joho/godotenv"
	"fmt"
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
	ResendAPIKey       string
	ContactEmailTo     string
	ContactEmailFrom   string
	TestMode           bool
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load("../.env")

	testMode := os.Getenv("VITE_TEST_MODE") == "true"

	requiredVars := map[string]string{
		"PORT":               "",
		"ALLOWED_ORIGIN":    "",
		"JWT_SECRET":        "",
		"TURSO_DATABASE_URL": "",
		"TURSO_AUTH_TOKEN":   "",
	}

	if !testMode {
		requiredVars["VITE_GOOGLE_CLIENT_ID"] = ""
		requiredVars["GOOGLE_CLIENT_SECRET"] = ""
		requiredVars["ORS_API_KEY"] = ""
		requiredVars["RESEND_API_KEY"] = ""
		requiredVars["CONTACT_EMAIL_TO"] = ""
		requiredVars["CONTACT_EMAIL_FROM"] = ""
	}

	if err := validateEnvVars(requiredVars); err != nil {
		return nil, fmt.Errorf("environment validation failed: %w", err)
	}

	port, _ := getEnv("PORT")
	allowedOrigin, _ := getEnv("ALLOWED_ORIGIN")
	jwtSecret, _ := getEnv("JWT_SECRET")
	tursoDatabaseURL, _ := getEnv("TURSO_DATABASE_URL")
	tursoAuthToken, _ := getEnv("TURSO_AUTH_TOKEN")
	googleClientID, _ := getEnv("VITE_GOOGLE_CLIENT_ID")
	googleClientSecret, _ := getEnv("GOOGLE_CLIENT_SECRET")
	orsAPIKey, _ := getEnv("ORS_API_KEY")
	resendAPIKey, _ := getEnv("RESEND_API_KEY")
	contactEmailTo, _ := getEnv("CONTACT_EMAIL_TO")
	contactEmailFrom, _ := getEnv("CONTACT_EMAIL_FROM")

	return &Config{
		Port:               port,
		AllowedOrigin:      allowedOrigin,
		JWTSecret:          jwtSecret,
		TursoDatabaseURL:   tursoDatabaseURL,
		TursoAuthToken:     tursoAuthToken,
		GoogleClientID:     googleClientID,
		GoogleClientSecret: googleClientSecret,
		ORSAPIKey:          orsAPIKey,
		ResendAPIKey:       resendAPIKey,
		ContactEmailTo:     contactEmailTo,
		ContactEmailFrom:   contactEmailFrom,
		TestMode:           testMode,
	}, nil
}

func getEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("environment variable %s not found", key)
	}
	return value, nil
}

func validateEnvVars(requiredVars map[string]string) error {
	var missingVars []string
	for key := range requiredVars {
		_, exists := os.LookupEnv(key)
		if !exists {
			missingVars = append(missingVars, key)
		}
	}
	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missingVars)
	}
	return nil
}