package role

//
//import (
//	"context"
//	"errors"
//	"google.golang.org/grpc/codes"
//	"google.golang.org/grpc/status"
//	"log/slog"
//	"sso_go_grpc/internal/config"
//	authService "sso_go_grpc/internal/services/auth"
//	"sso_go_grpc/internal/storage"
//	sso "sso_go_grpc/proto/gen"
//)
//
//type RolesInterface interface {
//	CreateRole(
//		ctx context.Context,
//		token string,
//		name,
//		description string,
//	) (
//		role *sso.Role,
//		err error,
//	)
//
//	GetUserRoles(
//		ctx context.Context,
//		userId uint64,
//	) (roles []*sso.Role,
//		err error)
//
//	UpdateRole(
//		ctx context.Context,
//		token string,
//		roleId uint64,
//		name,
//		description string,
//
//	) (role *sso.Role,
//		err error)
//
//	DeleteRole(
//		ctx context.Context,
//		token string,
//		roleId uint64,
//	) (
//		err error)
//
//	AddUserRole(
//		ctx context.Context,
//		token string,
//		roleId,
//		userId uint64,
//	) (*sso.User, error)
//
//	RemoveUserRole(
//		ctx context.Context,
//		token string,
//		roleId,
//		userId uint64,
//
//	) (user *sso.User, err error)
//
//	VerifyUserRoles(
//		ctx context.Context,
//		roleIds []uint64,
//		userId uint64,
//	) (verified bool, err error)
//}
//
//type Service struct {
//	Log *slog.Logger
//	Cfg *config.Config
//}
//
//func NewRoleService(ctx context.Context, auth authService.Auth) *Service {
//	return &Service{Log: auth.Log, Cfg: auth.Cfg}
//}
//
//func (auth *Service) CreateRole(
//	ctx context.Context,
//	token string,
//	name,
//	description string,
//) (
//	*sso.Role,
//	error,
//) {
//	role, err := auth.P.CreateRole(ctx, name, description)
//
//	if err != nil {
//		if errors.Is(storage.ErrRoleExists, err) {
//			return nil, storage.ErrRoleExists
//		}
//		return nil, err
//	}
//
//	return &sso.Role{RoleId: role.Id, Name: role.Name, Description: role.Description}, nil
//}
//
//func (auth *Service) GetUserRoles(
//	ctx context.Context,
//	userId uint64,
//) (roles []*sso.Role,
//	err error) {
//	return nil, status.Error(codes.Internal, "Not implemented")
//}
//
//func (auth *Service) DeleteRole(
//	ctx context.Context,
//	token string,
//	roleId uint64,
//) error {
//	err := auth.RoleProvider.DeleteRole(ctx, roleId)
//
//	if err != nil {
//		if errors.Is(storage.ErrRoleNotExists, err) {
//			return storage.ErrRoleNotExists
//		}
//		return err
//	}
//	return nil
//}
//
//func (auth *Service) UpdateRole(
//	ctx context.Context,
//	token string,
//	roleId uint64,
//	name,
//	description string,
//
//) (*sso.Role,
//	error,
//) {
//	role, err := auth.RoleProvider.UpdateRole(ctx, name, description, roleId)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return &sso.Role{
//		RoleId:      role.Id,
//		Name:        role.Name,
//		Description: role.Description,
//	}, nil
//}
//
//func (auth *Service) AddUserRole(
//	ctx context.Context,
//	token string,
//	roleId,
//	userId uint64,
//) (*sso.User, error) {
//	op := "service.auth.AddUserRole"
//	logger := auth.Log.With("op", op)
//
//	// check if userId and roleId are valid
//	err := auth.CheckUserAndRoleExists(ctx, userId, roleId)
//
//	if err != nil {
//		logger.Debug("Error on checking user and role ids", err)
//		return nil, storage.ErrUserAndRoleIvalid
//	}
//
//	hasTheRole, err := auth.RoleProvider.VerifyUserRole(ctx, roleId, userId)
//
//	// check if the user already has the role
//	if hasTheRole {
//		return nil, storage.ErrUserAlreadyHasTHeRole
//	}
//
//	// add role
//	err = auth.RoleProvider.AddUserRole(ctx, roleId, userId)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return auth.GetUserById(ctx, userId)
//}
//
//func (auth *Service) RemoveUserRole(
//	ctx context.Context,
//	token string,
//	roleId,
//	userId uint64,
//
//) (*sso.User, error) {
//	op := "service.auth.RemoveUserRole"
//	logger := auth.Log.With("op", op)
//
//	// check if userId and roleId are valid
//	err := auth.CheckUserAndRoleExists(ctx, userId, roleId)
//	if err != nil {
//		logger.Debug("Error on checking user and role ids", err)
//		return nil, storage.ErrUserAndRoleIvalid
//	}
//
//	hasTheRole, err := auth.RoleProvider.VerifyUserRole(ctx, roleId, userId)
//	if err != nil {
//		logger.Debug("Error on checking user role", err)
//		return nil, err
//	}
//
//	// check if the user already has the role
//	if !hasTheRole {
//		return nil, storage.ErrUserDontHaveTheRole
//	}
//
//	err = auth.RoleProvider.RemoveUserRole(ctx, roleId, userId)
//
//	if err != nil {
//		return nil, err
//	}
//
//	//get new user with updated roles
//	user, err := auth.GetUserById(ctx, userId)
//
//	if err != nil {
//		return nil, err
//	}
//
//	var roles []*sso.Role
//
//	for _, role := range user.Roles {
//		roles = append(roles, &sso.Role{RoleId: role.RoleId, Description: role.Description, Name: role.Name})
//	}
//
//	return &sso.User{UserId: userId, Roles: roles, Username: user.Username, Email: user.Email}, nil
//}
//
//func (auth *Service) VerifyUserRoles(
//	ctx context.Context,
//	roleIds []uint64,
//	userId uint64,
//) (verified bool, err error) {
//
//	//op := "service.auth.VerifyUserRole"
//	//logger := auth.log.With("op", op)
//
//	//check user exists
//	_, err = auth.UserProvider.GetUserById(ctx, userId)
//
//	// handle errors for user getting
//	if err != nil {
//		if errors.Is(storage.ErrUserNotExists, err) {
//			return false, storage.ErrUserNotExists
//		}
//		return false, err
//	}
//
//	for _, roleId := range roleIds {
//
//		//check if role exits
//		_, err = auth.RoleProvider.GetRoleById(ctx, roleId)
//
//		// handle errors for role getting
//		if err != nil {
//			if errors.Is(storage.ErrRoleNotExists, err) {
//				return false, storage.ErrRoleNotExists
//			}
//
//			return false, err
//		}
//
//		verified, err = auth.RoleProvider.VerifyUserRole(ctx, uint64(roleId), userId)
//	}
//
//	return verified, nil
//}
//
//// CheckUserAndRoleExists returns an error if
//// user with that id or;
//// role with that role id do not exist;
//func (auth *Service) CheckUserAndRoleExists(ctx context.Context, userId, roleId uint64) error {
//	//check user exists
//	_, err := auth.UserProvider.GetUserById(ctx, userId)
//
//	// handle errors for user getting
//	if err != nil {
//		if errors.Is(storage.ErrUserNotExists, err) {
//			return storage.ErrUserNotExists
//		}
//		return err
//	}
//
//	//check if role exits
//	_, err = auth.RoleProvider.GetRoleById(ctx, roleId)
//
//	// handle errors for role getting
//	if err != nil {
//		if errors.Is(storage.ErrRoleNotExists, err) {
//			return storage.ErrRoleNotExists
//		}
//
//		return err
//	}
//
//	return nil
//}
