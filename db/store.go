package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var schema = `
    CREATE TABLE IF NOT EXISTS students (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        uid TEXT NOT NULL UNIQUE,
        name TEXT NOT NULL,
        status TEXT CHECK( status IN ('IN','OUT') ) NOT NULL DEFAULT 'OUT',
        start_date TEXT NOT NULL,
        end_date TEXT NOT NULL,
        schedule TEXT NOT NULL,
        done_seconds INTEGER NOT NULL DEFAULT 0,
        excluded_days TEXT DEFAULT '[]',
        break_time INTEGER NOT NULL DEFAULT 0,
        created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS scans (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        uid TEXT NOT NULL,
        timestamp TEXT NOT NULL, 
        student_id INTEGER NOT NULL
    );

	CREATE TABLE IF NOT EXISTS devices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		secret_key TEXT NOT NULL
	);
`

// TODO: foreing keys

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func Connect() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	// TODO: Remove when moving away from memory
	db.SetMaxOpenConns(1)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	if _, err = db.Exec(schema); err != nil {
		return nil, err
	}

	return db, nil
}
