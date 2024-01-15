package roleService

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"sso_go_grpc/internal/config"
	userService "sso_go_grpc/internal/services/user"
	"sso_go_grpc/internal/storage"
	"sso_go_grpc/internal/storage/postgres/role"
	sso "sso_go_grpc/proto/gen"
)

type roleServiceInterface interface {
	GetUserRoles(
		ctx context.Context,
		userId uint64,
	) (roles []*sso.Role,
		err error)

	UpdateRole(
		ctx context.Context,
		token string,
		roleId uint64,
		name,
		description string,

	) (role *sso.Role,
		err error)

	DeleteRole(
		ctx context.Context,
		token string,
		roleId uint64,
	) (
		err error)

	AddUserRole(
		ctx context.Context,
		token string,
		roleId,
		userId uint64,
	) (*sso.User, error)

	RemoveUserRole(
		ctx context.Context,
		token string,
		roleId,
		userId uint64,

	) (user *sso.User, err error)

	VerifyUserRoles(
		ctx context.Context,
		roleIds []uint64,
		userId uint64,
	) (verified bool, err error)
}

type RoleService struct {
	userService  *userService.UserService
	cfg          *config.Config
	log          *slog.Logger
	roleProvider *role.Storage
}

func New(userService *userService.UserService, cfg *config.Config, log *slog.Logger, roleProvider *role.Storage) *RoleService {
	return &RoleService{log: log, cfg: cfg, roleProvider: roleProvider, userService: userService}
}

func (s *RoleService) CreateRole(
	ctx context.Context,
	token string,
	name,
	description string,
) (
	*sso.Role,
	error,
) {
	role, err := s.roleProvider.CreateRole(ctx, name, description)

	if err != nil {
		if errors.Is(storage.ErrRoleExists, err) {
			return nil, storage.ErrRoleExists
		}
		return nil, err
	}

	return &sso.Role{RoleId: role.Id, Name: role.Name, Description: role.Description}, nil
}

func (s *RoleService) GetUserRoles(
	ctx context.Context,
	userId uint64,
) (roles []*sso.Role,
	err error) {
	return nil, status.Error(codes.Internal, "Not implemented")
}

func (s *RoleService) DeleteRole(
	ctx context.Context,
	token string,
	roleId uint64,
) error {
	err := s.roleProvider.DeleteRole(ctx, roleId)

	if err != nil {
		if errors.Is(storage.ErrRoleNotExists, err) {
			return storage.ErrRoleNotExists
		}
		return err
	}
	return nil
}

func (s *RoleService) UpdateRole(
	ctx context.Context,
	token string,
	roleId uint64,
	name,
	description string,

) (*sso.Role,
	error,
) {
	role, err := s.roleProvider.UpdateRole(ctx, name, description, roleId)

	if err != nil {
		return nil, err
	}

	return &sso.Role{
		RoleId:      role.Id,
		Name:        role.Name,
		Description: role.Description,
	}, nil
}

func (s *RoleService) AddUserRole(
	ctx context.Context,
	token string,
	roleId,
	userId uint64,
) (*sso.User, error) {
	op := "service.s.AddUserRole"
	logger := s.log.With("op", op)

	// check if userId and roleId are valid
	err := s.CheckUserAndRoleExists(ctx, userId, roleId)

	if err != nil {
		logger.Debug("Error on checking user and role ids", err)
		return nil, storage.ErrUserAndRoleIvalid
	}

	hasTheRole, err := s.roleProvider.VerifyUserRole(ctx, roleId, userId)

	// check if the user already has the role
	if hasTheRole {
		return nil, storage.ErrUserAlreadyHasTHeRole
	}

	// add role
	err = s.roleProvider.AddUserRole(ctx, roleId, userId)

	if err != nil {
		return nil, err
	}

	return s.userService.GetUserById(ctx, userId)
}

func (s *RoleService) RemoveUserRole(
	ctx context.Context,
	token string,
	roleId,
	userId uint64,

) (*sso.User, error) {
	op := "service.s.RemoveUserRole"
	logger := s.log.With("op", op)

	// check if userId and roleId are valid
	err := s.CheckUserAndRoleExists(ctx, userId, roleId)
	if err != nil {
		logger.Debug("Error on checking user and role ids", err)
		return nil, storage.ErrUserAndRoleIvalid
	}

	hasTheRole, err := s.roleProvider.VerifyUserRole(ctx, roleId, userId)
	if err != nil {
		logger.Debug("Error on checking user role", err)
		return nil, err
	}

	// check if the user already has the role
	if !hasTheRole {
		return nil, storage.ErrUserDontHaveTheRole
	}

	err = s.roleProvider.RemoveUserRole(ctx, roleId, userId)

	if err != nil {
		return nil, err
	}

	//get new user with updated roles
	user, err := s.userService.GetUserById(ctx, userId)

	if err != nil {
		return nil, err
	}

	var roles []*sso.Role

	for _, role := range user.Roles {
		roles = append(roles, &sso.Role{RoleId: role.RoleId, Description: role.Description, Name: role.Name})
	}

	return &sso.User{UserId: userId, Roles: roles, Username: user.Username, Email: user.Email}, nil
}

func (s *RoleService) VerifyUserRoles(
	ctx context.Context,
	roleIds []uint64,
	userId uint64,
) (verified bool, err error) {

	//op := "service.s.VerifyUserRole"
	//logger := s.log.With("op", op)

	//check user exists
	_, err = s.userService.GetUserById(ctx, userId)

	// handle errors for user getting
	if err != nil {
		if errors.Is(storage.ErrUserNotExists, err) {
			return false, storage.ErrUserNotExists
		}
		return false, err
	}

	for _, roleId := range roleIds {

		//check if role exits
		_, err = s.roleProvider.GetRoleById(ctx, roleId)

		// handle errors for role getting
		if err != nil {
			if errors.Is(storage.ErrRoleNotExists, err) {
				return false, storage.ErrRoleNotExists
			}

			return false, err
		}

		verified, err = s.roleProvider.VerifyUserRole(ctx, uint64(roleId), userId)
	}

	return verified, nil
}

// CheckUserAndRoleExists returns an error if
// user with that id or;
// role with that role id do not exist;
func (s *RoleService) CheckUserAndRoleExists(ctx context.Context, userId, roleId uint64) error {
	//check user exists
	_, err := s.userService.GetUserById(ctx, userId)

	// handle errors for user getting
	if err != nil {
		if errors.Is(storage.ErrUserNotExists, err) {
			return storage.ErrUserNotExists
		}
		return err
	}

	//check if role exits
	_, err = s.roleProvider.GetRoleById(ctx, roleId)

	// handle errors for role getting
	if err != nil {
		if errors.Is(storage.ErrRoleNotExists, err) {
			return storage.ErrRoleNotExists
		}

		return err
	}

	return nil
}
