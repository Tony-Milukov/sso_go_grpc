package authService

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"sso_go_grpc/internal/config"
	"sso_go_grpc/internal/domain/models"
	"sso_go_grpc/internal/lib/bcrypt"
	"sso_go_grpc/internal/lib/jwt"
	"sso_go_grpc/internal/storage"
	"sso_go_grpc/internal/storage/postgres"
	sso "sso_go_grpc/proto/gen"
	//	jwt "sso_go_grpc/internal/lib/jwt"
)

// UserProvider is an interface for the Auth Service
type UserProvider interface {
	CreateUser(ctx context.Context, username, email, password string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserById(ctx context.Context, userId uint64) (*models.User, error)
	GetRoleById(ctx context.Context, roleId uint64) (*models.Role, error)
	CreateRole(ctx context.Context, name, description string) (*models.Role, error)
	UpdateRole(ctx context.Context, name, description string, roleId uint64) (*models.Role, error)
	DeleteRole(ctx context.Context, roleId uint64) error
	AddUserRole(ctx context.Context, roleId, userId uint64) error
	RemoveUserRole(ctx context.Context, roleId, userId uint64) error
	VerifyUserRole(ctx context.Context, roleId, userId uint64) (bool, error)
}
type Auth struct {
	log *slog.Logger
	UserProvider
	cfg *config.Config
}

// New this function returns new AuthService with userProvider where are all the postgres methods
func New(log *slog.Logger, userProvider *postgres.Storage, config *config.Config) *Auth {
	return &Auth{log: log, UserProvider: userProvider, cfg: config}
}

func (auth *Auth) Register(
	ctx context.Context,
	email,
	password,
	username string,
) (token string, userId uint64, err error) {
	op := "service.auth"
	log := auth.log.With("op", op)
	user, err := auth.UserProvider.CreateUser(ctx, email, password, username)

	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Debug("User with that username or email already exists")
			return "", 0, status.Error(codes.AlreadyExists, "User with that email or username already exists")
		}
		log.Debug("Error", err)
		return "", 0, status.Error(codes.Internal, "Internal Error")
	}

	// generate token
	token, err = jwt.NewToken(user, auth.cfg.JwtSecret)

	if err != nil {
		log.Debug("Error on generating jwt", err)
		return "", 0, status.Error(codes.Internal, "Internal Error")
	}
	return token, user.UserId, nil
}

func (auth *Auth) Login(
	ctx context.Context,
	email,
	password string,
) (token string, userId uint64, err error) {
	op := "auth.service.Login"
	logger := auth.log.With("op", op)

	logger.Debug("Getting User from the database")
	user, err := auth.UserProvider.GetUserByEmail(ctx, email)

	// if the user is not defined or
	// if user is defined but the password is incorrect
	if err != nil ||
		bcrypt.ComparePasswords(user.Password, password) != nil {
		logger.Debug("Invalid credentials")
		return "", 0, storage.ErrAuth
	}

	// generate token
	token, err = jwt.NewToken(user, auth.cfg.JwtSecret)

	if err != nil {
		logger.Debug("Error on generating jwt", err)
		return "", 0, status.Error(codes.Internal, "Internal Error")
	}
	return token, user.UserId, nil
}

func (auth *Auth) GetUserById(
	ctx context.Context,
	userId uint64,
) (*sso.User, error) {

	op := "auth.service.GetUserById"

	logger := auth.log.With("op", op)
	user, err := auth.UserProvider.GetUserById(ctx, userId)

	if err != nil {
		logger.Debug("Error on finding user with userId", userId)
		if errors.Is(storage.ErrUserNotExists, err) {
			return nil, storage.ErrUserNotExists
		}

		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	var roles []*sso.Role

	for _, role := range user.Roles {
		protoRole := &sso.Role{RoleId: role.Id, Name: role.Name, Description: role.Description}
		roles = append(roles, protoRole)
	}

	return &sso.User{Email: user.Email, UserId: userId, Roles: roles, Username: user.Username}, nil
}

func (auth *Auth) GetUserByEmail(
	ctx context.Context,
	email string,
) (
	*sso.User,
	error,
) {
	user, err := auth.UserProvider.GetUserByEmail(ctx, email)

	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	var protoRoles []*sso.Role

	for _, role := range user.Roles {
		protoRoles = append(protoRoles, &sso.Role{RoleId: role.Id, Name: role.Name, Description: role.Description})
	}

	return &sso.User{UserId: user.UserId, Email: user.Email, Username: user.Username, Roles: protoRoles}, err
}

func (auth *Auth) CreateRole(
	ctx context.Context,
	token string,
	name,
	description string,
) (
	*sso.Role,
	error,
) {
	role, err := auth.UserProvider.CreateRole(ctx, name, description)

	if err != nil {
		if errors.Is(storage.ErrRoleExists, err) {
			return nil, storage.ErrRoleExists
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.Role{RoleId: role.Id, Name: role.Name, Description: role.Description}, nil
}

func (auth *Auth) GetUserRoles(
	ctx context.Context,
	userId uint64,
) (roles []*sso.Role,
	err error) {
	return nil, status.Error(codes.Internal, "Not implemented")
}

func (auth *Auth) DeleteRole(
	ctx context.Context,
	token string,
	roleId uint64,
) error {
	err := auth.UserProvider.DeleteRole(ctx, roleId)

	if err != nil {
		if errors.Is(storage.ErrRoleNotExists, err) {
			return storage.ErrRoleNotExists
		}
		return err
	}
	return nil
}

func (auth *Auth) UpdateRole(
	ctx context.Context,
	token string,
	roleId uint64,
	name,
	description string,

) (*sso.Role,
	error,
) {
	role, err := auth.UserProvider.UpdateRole(ctx, name, description, roleId)

	if err != nil {
		return nil, err
	}

	return &sso.Role{
		RoleId:      role.Id,
		Name:        role.Name,
		Description: role.Description,
	}, nil
}

func (auth *Auth) AddUserRole(
	ctx context.Context,
	token string,
	roleId,
	userId uint64,
) (*sso.User, error) {
	op := "service.auth.AddUserRole"
	logger := auth.log.With("op", op)

	// check if userId and roleId are valid
	err := auth.CheckUserAndRoleExists(ctx, userId, roleId)

	if err != nil {
		logger.Debug("Error on checking user and role ids", err)
		return nil, storage.ErrUserAndRoleIvalid
	}

	hasTheRole, err := auth.UserProvider.VerifyUserRole(ctx, roleId, userId)

	// check if the user already has the role
	if hasTheRole {
		return nil, storage.ErrUserAlreadyHasTHeRole
	}

	// add role
	err = auth.UserProvider.AddUserRole(ctx, roleId, userId)

	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return auth.GetUserById(ctx, userId)
}

func (auth *Auth) RemoveUserRole(
	ctx context.Context,
	token string,
	roleId,
	userId uint64,

) (*sso.User, error) {
	op := "service.auth.RemoveUserRole"
	logger := auth.log.With("op", op)

	// check if userId and roleId are valid
	err := auth.CheckUserAndRoleExists(ctx, userId, roleId)
	if err != nil {
		logger.Debug("Error on checking user and role ids", err)
		return nil, storage.ErrUserAndRoleIvalid
	}

	hasTheRole, err := auth.UserProvider.VerifyUserRole(ctx, roleId, userId)
	if err != nil {
		logger.Debug("Error on checking user role", err)
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	// check if the user already has the role
	if !hasTheRole {
		return nil, storage.ErrUserDontHaveTheRole
	}

	err = auth.UserProvider.RemoveUserRole(ctx, roleId, userId)

	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	//get new user with updated roles
	user, err := auth.GetUserById(ctx, userId)

	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	var roles []*sso.Role

	for _, role := range user.Roles {
		roles = append(roles, &sso.Role{RoleId: role.RoleId, Description: role.Description, Name: role.Name})
	}

	return &sso.User{UserId: userId, Roles: roles, Username: user.Username, Email: user.Email}, nil
}

func (auth *Auth) VerifyUserRoles(
	ctx context.Context,
	roleIds []uint64,
	userId uint64,
) (verified bool, err error) {

	op := "service.auth.VerifyUserRole"
	logger := auth.log.With("op", op)

	//check user exists
	_, err = auth.UserProvider.GetUserById(ctx, userId)

	// handle errors for user getting
	if err != nil {
		if errors.Is(storage.ErrUserNotExists, err) {
			return false, storage.ErrUserNotExists
		}
		return false, status.Error(codes.Internal, "Internal Server Error")
	}

	if err != nil {
		logger.Debug("Error on checking user and role ids", err)
		return false, status.Error(codes.Internal, "Internal Server Error")
	}
	for _, roleId := range roleIds {

		//check if role exits
		_, err = auth.UserProvider.GetRoleById(ctx, roleId)

		// handle errors for role getting
		if err != nil {
			if errors.Is(storage.ErrRoleNotExists, err) {
				return false, storage.ErrRoleNotExists
			}

			return false, status.Error(codes.Internal, "Internal Server Error")
		}

		verified, err = auth.UserProvider.VerifyUserRole(ctx, uint64(roleId), userId)

		if err != nil {

			return false, status.Error(codes.Internal, "Internal Server Error")
		}
	}

	return verified, nil
}

// CheckUserAndRoleExists returns an error if
// user with that id or;
// role with that role id do not exist;
func (auth *Auth) CheckUserAndRoleExists(ctx context.Context, userId, roleId uint64) error {
	//check user exists
	_, err := auth.UserProvider.GetUserById(ctx, userId)

	// handle errors for user getting
	if err != nil {
		if errors.Is(storage.ErrUserNotExists, err) {
			return storage.ErrUserNotExists
		}
		return status.Error(codes.Internal, "Internal Server Error")
	}

	//check if role exits
	_, err = auth.UserProvider.GetRoleById(ctx, roleId)

	// handle errors for role getting
	if err != nil {
		if errors.Is(storage.ErrRoleNotExists, err) {
			return storage.ErrRoleNotExists
		}

		return status.Error(codes.Internal, "Internal Server Error")
	}

	return nil
}
