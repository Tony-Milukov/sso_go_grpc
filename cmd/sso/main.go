package main

import (
	"log/slog"
	"os"
)

func main() {
	//cfg := config.MustLoad()

	//log := setupLogger(cfg.Env)

	//storage := postgres.MustLoad(cfg.DB_link, cfg.DB_Type)

	//TODO: SET UP APPLICATION

	//TODO: RUN APPLICATION
}

// setupLogger returns logger depending on env | default = LevelDebug; dev = LevelInfo
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	//getting logger depending on env
	switch env {
	case "dev":
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	return log
}
