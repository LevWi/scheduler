package test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func InitSqliteDB(path string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := applyMigrations(db.DB); err != nil {
		return nil, err
	}
	return db, nil
}

func InitTmpDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	if err := applyMigrations(db.DB); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}
	return db
}
