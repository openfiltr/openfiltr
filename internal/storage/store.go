package storage

import "database/sql"

// Store is the application storage seam.
// PostgreSQL backs it today; later backends can swap in the same boundary.
type Store interface {
	Close() error
	Ping() error
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Begin() (*sql.Tx, error)
}

var _ Store = (*sql.DB)(nil)
