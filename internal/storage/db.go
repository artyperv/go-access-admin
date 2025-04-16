package storage

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := createSchema(db); err != nil {
		return nil, err
	}

	return &DB{conn: db}, nil
}

func createSchema(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS accesses (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        password TEXT NOT NULL,
        htpasswd_path TEXT NOT NULL,
        expires_at DATETIME NOT NULL,
        is_admin BOOLEAN DEFAULT 0
    );`

	_, err := db.Exec(query)
	return err
}

func (d *DB) Close() error {
	return d.conn.Close()
}
