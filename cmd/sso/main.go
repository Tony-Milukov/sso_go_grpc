package main

import (
	"log/slog"
	"os"
	app "sso_go_grpc/internal/app"
	"sso_go_grpc/internal/config"
)

func main() {

	//setting up config
	cfg := config.MustLoad()

	//setting up logger
	log := setupLogger(cfg.Env)

	application := app.New(log, cfg)

	//running GRPC Server
	//if error -> panic
	application.GRPCServer.MustRun()

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
