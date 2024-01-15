package services

import (
	"log/slog"
	"sso_go_grpc/internal/config"
	roleService "sso_go_grpc/internal/services/role"
	userService "sso_go_grpc/internal/services/user"
	"sso_go_grpc/internal/storage/postgres"
	roleStorage "sso_go_grpc/internal/storage/postgres/role"
	"sso_go_grpc/internal/storage/postgres/user"
)

type Services struct {
	Log *slog.Logger
	Cfg *config.Config
	Providers
	UserService *userService.UserService
	RoleService *roleService.RoleService
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

	user := userService.New(providers.UserProvider, log, config)

	role := roleService.New(user, config, log, providers.RoleProvider)

	return &Services{
		Providers:   providers,
		Cfg:         config,
		Log:         log,
		RoleService: role,
		UserService: user,
	}
}
