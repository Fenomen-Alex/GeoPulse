package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the SQL connection. The embedded SQLite backend keeps the server a
// self-contained single binary (`CGO_ENABLED=0` friendly); no external service
// or credentials are required at runtime.
type DB struct {
	*sql.DB
}

// InitDB opens the embedded SQLite database at dbPath (a file path, or
// ":memory:" for an ephemeral store used by tests and throwaway dev runs).
func InitDB(dbPath string) (*DB, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("db path is empty")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %q: %w", dbPath, err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database %q: %w", dbPath, err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return &DB{db}, nil
}

// RunMigrations executes the given SQL statements.
func (db *DB) RunMigrations(migrationSQL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, stmt := range splitStatements(migrationSQL) {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

// UpsertUser creates or refreshes a user row on successful OIDC login.
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

// EnsureUser creates a placeholder user row on first quota interaction. The
// OIDC flow does not yet persist a user row, so quota tracking inserts a
// synthetic identity to satisfy the user_daily_usage foreign key. Idempotent.
func (db *DB) EnsureUser(userID string) error {
	query := `
		INSERT OR IGNORE INTO users (id, email, name)
		VALUES (?, ?, ?);
	`
	_, err := db.Exec(query, userID, userID+"@geopulse.local", userID)
	return err
}

// GetUserDailyUsage returns the user's used count and configured quota. The
// quota falls back to the default if the user row does not exist yet.
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

// IncrementUserUsage records one spatial run for the user on today's window.
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

// SaveContactRequest persists a contact form submission for audit/review.
func (db *DB) SaveContactRequest(name, email, subject, message, ip string) error {
	query := `
		INSERT INTO contact_requests (name, email, subject, message, ip_address)
		VALUES (?, ?, ?, ?, ?);
	`
	_, err := db.Exec(query, name, email, subject, message, ip)
	return err
}