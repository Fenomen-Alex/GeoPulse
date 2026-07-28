package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/tursodatabase/go-libsql"
)

type DB struct {
	*sql.DB
}

func InitDB() (*DB, error) {
	dbURL := os.Getenv("TURSO_DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	if dbURL == "" {
		return nil, fmt.Errorf("TURSO_DATABASE_URL is not set")
	}

	connStr := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)
	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Turso DB: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping Turso DB: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) RunMigrations(migrationSQL string) error {
	_, err := db.Exec(migrationSQL)
	return err
}

func (db *DB) UpsertUser(id, email, name, avatarURL string) error {
	query := `
		INSERT INTO users (id, email, name, avatar_url, last_login)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			avatar_url = excluded.avatar_url,
			last_login = CURRENT_TIMESTAMP;
	`
	_, err := db.Exec(query, id, email, name, avatarURL)
	return err
}

func (db *DB) GetUserDailyUsage(userID string) (int, int, error) {
	var dailyQuota int
	var usedCount int

	err := db.QueryRow("SELECT daily_quota FROM users WHERE id = ?", userID).Scan(&dailyQuota)
	if err != nil {
		dailyQuota = 15
	}

	today := time.Now().Format("2006-01-02")
	_ = db.QueryRow("SELECT used_count FROM user_daily_usage WHERE user_id = ? AND usage_date = ?", userID, today).Scan(&usedCount)

	return usedCount, dailyQuota, nil
}

func (db *DB) IncrementUserUsage(userID string) error {
	today := time.Now().Format("2006-01-02")
	query := `
		INSERT INTO user_daily_usage (user_id, usage_date, used_count)
		VALUES (?, ?, 1)
		ON CONFLICT(user_id, usage_date) DO UPDATE SET
			used_count = used_count + 1;
	`
	_, err := db.Exec(query, userID, today)
	return err
}

func (db *DB) SaveContactRequest(name, email, subject, message, ip string) error {
	query := `
		INSERT INTO contact_requests (name, email, subject, message, ip_address)
		VALUES (?, ?, ?, ?, ?);
	`
	_, err := db.Exec(query, name, email, subject, message, ip)
	return err
}