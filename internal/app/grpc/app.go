package grpcApp

import (
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	roleServer "sso_go_grpc/internal/grpc/role"
	userServer "sso_go_grpc/internal/grpc/user"
	"sso_go_grpc/internal/services"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func (app *App) MustRun() {
	if err := app.Run(); err != nil {
		panic(err)
	}

}

// Run this method runs the server
func (app *App) Run() error {
	const op = "grpc.app.Run"

	//setup logger for this function
	log := app.log.With(slog.String("op", op))

	//starting TCP listener
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", app.port))

	if err != nil {
		log.Error(fmt.Sprintf("%s: %w", op, err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Starting Grpc Server", "port", app.port)
	//Serving the listener to the GRPC server
	//if there is an error return it
	if err := app.gRPCServer.Serve(l); err != nil {
		log.Info("Successfully started Grpc server on ", "port", app.port)

		log.Error(fmt.Sprintf("%s: %w", op, err))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("Successfully started Grpc server on ", "port", app.port)
	return nil
}

func New(log *slog.Logger, services *services.Services, port int) *App {
	//creating new Grpc Server
	grpcServer := grpc.NewServer()

	//Register the new gRPC Server with the  AUthService
	userServer.RegisterServer(grpcServer, services.UserService)
	roleServer.RegisterServer(grpcServer, services.RoleService)

	//return a structure with that params
	return &App{log: log, gRPCServer: grpcServer, port: port}
}
