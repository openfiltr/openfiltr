package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func Open(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database url is required")
	}
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Hour)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging db: %w", err)
	}
	return db, nil
}
