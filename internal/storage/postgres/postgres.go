package postgres

import (
	"database/sql"
	"fmt"
)

type Storage struct {
	db *sql.DB
}

// MustLoad this function returns an Storage, if there is an error , it panics
func MustLoad(dbLink, dbType string) *Storage {
	op := "storage.postgres.MustLoad"
	db, err := sql.Open(dbType, dbLink)

	if err != nil {
		panic(fmt.Sprintf("%s: %w", op, err))
	}

	return &Storage{db}
}
