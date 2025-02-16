package utils

import (
	"database/sql"
	"github.com/pressly/goose/v3"
	"log"
	"path/filepath"
	"runtime"
)

func GooseDown(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose.SetDialect: %v", err)
	}

	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filename)
	migrationsPath := filepath.Join(basePath, "../../migrations")

	err := goose.Down(db, migrationsPath)
	return err
}
