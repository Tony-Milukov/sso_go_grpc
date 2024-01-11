package authServer

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso_go_grpc/internal/storage"
	sso "sso_go_grpc/proto/gen"
)

// Service is an interface for the Auth Service
type Service interface {
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

type serverApi struct {
	authService Service
	sso.UnimplementedAuthServer
}

func RegisterServer(Grpc *grpc.Server, authService Service) {
	sso.RegisterAuthServer(Grpc, &serverApi{authService: authService})
}

func (s *serverApi) Register(ctx context.Context, req *sso.RegisterRequest) (res *sso.RegisterResponse, err error) {

	if req.GetEmail() == "" || req.GetPassword() == "" || req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Arguments, expected: email, password, username")
	}

	token, userId, err := s.authService.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetUsername())

	if err != nil {
		return nil, err
		//	return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.RegisterResponse{Token: token, UserId: userId}, nil
}

func (s *serverApi) Login(ctx context.Context, req *sso.LoginRequest) (res *sso.LoginResponse, err error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Arguments, expected: email, password")
	}

	token, userId, err := s.authService.Login(ctx, req.GetEmail(), req.GetPassword())

	if err != nil {
		if errors.Is(err, storage.ErrAuth) {
			return nil, status.Error(codes.Internal, storage.ErrAuth.Error())
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.LoginResponse{Token: token, UserId: userId}, nil
}

func (s *serverApi) GetUserById(ctx context.Context, req *sso.GetUserByIdRequest) (res *sso.GetUserByIdResponse, err error) {
	user, err := s.authService.GetUserById(ctx, req.GetUserId())

	if err != nil {
		if errors.Is(storage.ErrUserNotExists, err) {
			return nil, status.Error(codes.NotFound, storage.ErrUserNotExists.Error())
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.GetUserByIdResponse{User: user}, nil

}

func (s *serverApi) GetUserByEmail(ctx context.Context, req *sso.GetUserEmailRequest) (res *sso.GetUserEmailResponse, err error) {
	user, err := s.authService.GetUserByEmail(ctx, req.GetEmail())

	if err != nil {
		if errors.Is(storage.ErrUserNotExists, err) {
			return nil, status.Error(codes.NotFound, storage.ErrUserNotExists.Error())
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.GetUserEmailResponse{User: user}, nil

}

func (s *serverApi) CreateRole(ctx context.Context, req *sso.CreateRoleRequest) (res *sso.CreateRoleResponse, err error) {
	role, err := s.authService.CreateRole(ctx, req.GetToken(), req.GetName(), req.GetDescription())

	if err != nil {
		if errors.Is(storage.ErrRoleExists, err) {
			return nil, err
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.CreateRoleResponse{Role: &sso.Role{RoleId: role.RoleId, Description: role.Description, Name: role.Name}}, nil
}

func (s *serverApi) UpdateRole(ctx context.Context, req *sso.UpdateRoleRequest) (res *sso.UpdateRoleResponse, err error) {
	role, err := s.authService.UpdateRole(ctx, req.GetToken(), req.GetRoleId(), req.GetName(), req.GetDescription())
	if err != nil {
		if errors.Is(storage.ErrRoleNotExists, err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
	}

	return &sso.UpdateRoleResponse{Role: role}, nil
}

func (s *serverApi) DeleteRole(ctx context.Context, req *sso.DeleteRoleRequest) (*sso.DeleteRoleResponse, error) {
	err := s.authService.DeleteRole(ctx, req.GetToken(), req.GetRoleId())

	if err != nil {
		if errors.Is(storage.ErrRoleNotExists, err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.DeleteRoleResponse{Message: "Successfully Deleted the Role"}, nil
}

func (s *serverApi) AddUserRole(ctx context.Context, req *sso.AddUserRoleRequest) (res *sso.AddUserRoleResponse, err error) {
	user, err := s.authService.AddUserRole(ctx, req.GetToken(), req.GetRoleId(), req.GetUserId())

	if err != nil {
		if errors.Is(storage.ErrUserAndRoleIvalid, err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.AddUserRoleResponse{User: user}, nil

}

func (s *serverApi) RemoveUserRole(ctx context.Context, req *sso.RemoveUserRoleRequest) (res *sso.RemoveUserRoleResponse, err error) {
	user, err := s.authService.RemoveUserRole(ctx, req.GetToken(), req.GetRoleId(), req.GetUserId())

	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.RemoveUserRoleResponse{User: user}, nil
}

func (s *serverApi) VerifyUserRoles(ctx context.Context, req *sso.VerifyUserRolesRequest) (res *sso.VerifyUserRolesResponse, err error) {
	verified, err := s.authService.VerifyUserRoles(ctx, req.GetRoleIds(), req.GetUserId())

	if err != nil {
		if errors.Is(storage.ErrRoleNotExists, err) || errors.Is(storage.ErrUserNotExists, err) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.VerifyUserRolesResponse{Verified: verified}, nil
}
