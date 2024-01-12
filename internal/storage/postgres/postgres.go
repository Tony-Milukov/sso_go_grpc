package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"log/slog"
	"sso_go_grpc/internal/config"
	"sso_go_grpc/internal/storage/postgres/role"
	"sso_go_grpc/internal/storage/postgres/user"
	_ "strconv"
)

type Storage struct {
	Db     *sql.DB
	Config *config.Config
	Log    *slog.Logger

	User *user.Storage
	Role *role.Storage
}

// MustLoad this function returns a Storage, if there is an error , it panics
func MustLoad(cfg *config.Config, log *slog.Logger) *Storage {
	op := "storage.postgres.MustLoad"
	db, err := sql.Open(cfg.DbType, cfg.DbLink)

	if err != nil {
		panic(fmt.Sprintf("%s: %w", op, err))
	}

	fmt.Printf("Database was succesfully connected\n")

	return &Storage{
		Db:   db,
		Log:  log,
		User: user.CreateStorage(db, log),
		Role: role.CreateStorage(db, log),
	}
}
