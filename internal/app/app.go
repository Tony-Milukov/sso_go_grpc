package app

import (
	"log/slog"
	grpcApp "sso_go_grpc/internal/app/grpc"
	"sso_go_grpc/internal/config"
	"sso_go_grpc/internal/services"
	"sso_go_grpc/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcApp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {

	//setting up storage
	storage := postgres.MustLoad(cfg, log)

	service := services.New(log, storage, cfg)

	app := grpcApp.New(log, service, cfg.GRPC.Port)

	return &App{
		GRPCServer: app,
	}
}
