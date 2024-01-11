package app

import (
	"log/slog"
	grpcApp "sso_go_grpc/internal/app/grpc"
	"sso_go_grpc/internal/config"
	authService "sso_go_grpc/internal/services/auth"
	"sso_go_grpc/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcApp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {

	//setting up storage
	storage := postgres.MustLoad(cfg, log)

	autService := authService.New(log, storage, cfg)

	app := grpcApp.New(log, autService, cfg.GRPC.Port)

	return &App{
		GRPCServer: app,
	}
}
