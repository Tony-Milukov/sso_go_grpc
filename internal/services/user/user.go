package userService

import (
	"context"
	"errors"
	"log/slog"
	"sso_go_grpc/internal/config"
	"sso_go_grpc/internal/lib/bcrypt"
	"sso_go_grpc/internal/lib/jwt"
	"sso_go_grpc/internal/storage"
	"sso_go_grpc/internal/storage/postgres/user"
	sso "sso_go_grpc/proto/gen"
)

type userServiceInterface interface {
	Register(
		ctx context.Context,
		email string,
		password,
		username string,
	) (token string, userId uint64,
		err error)

	Login(
		ctx context.Context,
		email,
		password string,
	) (token string, userId uint64, err error)

	GetUserById(
		ctx context.Context,
		userId uint64,
	) (
		user *sso.User,
		err error,
	)

	GetUserByEmail(
		ctx context.Context,
		email string,
	) (
		user *sso.User,
		err error,
	)

	CreateRole(
		ctx context.Context,
		token string,
		name,
		description string,
	) (
		role *sso.Role,
		err error,
	)
}

type UserService struct {
	userProvider *user.Storage
	log          *slog.Logger
	config       *config.Config
	userServiceInterface
}

func New(userProvider *user.Storage, log *slog.Logger, cfg *config.Config) *UserService {
	return &UserService{userProvider: userProvider, log: log}
}
func (s *UserService) Register(
	ctx context.Context,
	email,
	password,
	username string,
) (token string, userId uint64, err error) {
	op := "service.auth"
	log := s.log.With("op", op)
	user, err := s.userProvider.CreateUser(ctx, email, password, username)

	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Debug("User with that username or email already exists")

			return "", 0, storage.ErrUserExists
		}
		log.Debug("Error", err)
		return "", 0, err
	}

	// generate token
	token, err = jwt.NewToken(user, s.config.JwtSecret)

	if err != nil {
		log.Debug("Error on generating jwt", err)
		return "", 0, err
	}
	return token, user.UserId, nil
}

func (s *UserService) login(
	ctx context.Context,
	email,
	password string,
) (token string, userId uint64, err error) {
	op := "auth.service.login"
	logger := s.log.With("op", op)

	logger.Debug("Getting User from the database")
	user, err := s.userProvider.GetUserByEmail(ctx, email)

	// if the user is not defined or
	// if user is defined but the password is incorrect
	if err != nil ||
		bcrypt.ComparePasswords(user.Password, password) != nil {
		logger.Debug("Invalid credentials")
		return "", 0, storage.ErrAuth
	}

	// generate token
	token, err = jwt.NewToken(user, s.config.JwtSecret)

	if err != nil {
		logger.Debug("Error on generating jwt", err)
		return "", 0, err
	}
	return token, user.UserId, nil
}

func (s *UserService) GetUserById(
	ctx context.Context,
	userId uint64,
) (*sso.User, error) {

	op := "auth.service.GetUserById"

	logger := s.log.With("op", op)
	user, err := s.userProvider.GetUserById(ctx, userId)

	if err != nil {
		logger.Debug("Error on finding user with userId", userId)
		if errors.Is(storage.ErrUserNotExists, err) {
			return nil, storage.ErrUserNotExists
		}

		return nil, err
	}

	var roles []*sso.Role

	for _, role := range user.Roles {
		protoRole := &sso.Role{RoleId: role.Id, Name: role.Name, Description: role.Description}
		roles = append(roles, protoRole)
	}

	return &sso.User{Email: user.Email, UserId: userId, Roles: roles, Username: user.Username}, nil
}

func (s *UserService) GetUserByEmail(
	ctx context.Context,
	email string,
) (
	*sso.User,
	error,
) {
	user, err := s.userProvider.GetUserByEmail(ctx, email)

	if err != nil {
		return nil, err
	}

	var protoRoles []*sso.Role

	for _, role := range user.Roles {
		protoRoles = append(protoRoles, &sso.Role{RoleId: role.Id, Name: role.Name, Description: role.Description})
	}

	return &sso.User{UserId: user.UserId, Email: user.Email, Username: user.Username, Roles: protoRoles}, err
}
