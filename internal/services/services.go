package services

import (
	"log/slog"
	"sso_go_grpc/internal/config"
	"sso_go_grpc/internal/services/role"
	"sso_go_grpc/internal/storage/postgres"
	roleStorage "sso_go_grpc/internal/storage/postgres/role"
	"sso_go_grpc/internal/storage/postgres/user"
)

type Services struct {
	Log *slog.Logger
	Cfg *config.Config
	Providers
	Role *role.Service
}

type Providers struct {
	UserProvider *user.Storage
	RoleProvider *roleStorage.Storage
}

// New this function returns new AuthService with userProvider where are all the postgres methods
func New(log *slog.Logger, storage *postgres.Storage, config *config.Config) *Services {
	providers := Providers{
		UserProvider: storage.User,
		RoleProvider: storage.Role,
	}

	return &Services{
		Providers: providers,
		Cfg:       config,
		Log:       log,
	}
}
