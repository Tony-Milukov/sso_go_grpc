package roleServer

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	roleService "sso_go_grpc/internal/services/role"
	"sso_go_grpc/internal/storage"
	sso "sso_go_grpc/proto/gen"
)

type serverApi struct {
	roleService *roleService.RoleService
	sso.UnimplementedRoleApiServer
}

func RegisterServer(Grpc *grpc.Server, roleService *roleService.RoleService) {
	sso.RegisterRoleApiServer(Grpc, &serverApi{roleService: roleService})
}

func (s *serverApi) CreateRole(ctx context.Context, req *sso.CreateRoleRequest) (res *sso.CreateRoleResponse, err error) {
	role, err := s.roleService.CreateRole(ctx, req.GetToken(), req.GetName(), req.GetDescription())

	if err != nil {
		if errors.Is(storage.ErrRoleExists, err) {
			return nil, err
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.CreateRoleResponse{Role: &sso.Role{RoleId: role.RoleId, Description: role.Description, Name: role.Name}}, nil
}

func (s *serverApi) UpdateRole(ctx context.Context, req *sso.UpdateRoleRequest) (res *sso.UpdateRoleResponse, err error) {
	role, err := s.roleService.UpdateRole(ctx, req.GetToken(), req.GetRoleId(), req.GetName(), req.GetDescription())
	if err != nil {
		if errors.Is(storage.ErrRoleNotExists, err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
	}

	return &sso.UpdateRoleResponse{Role: role}, nil
}

func (s *serverApi) DeleteRole(ctx context.Context, req *sso.DeleteRoleRequest) (*sso.DeleteRoleResponse, error) {
	err := s.roleService.DeleteRole(ctx, req.GetToken(), req.GetRoleId())

	if err != nil {
		if errors.Is(storage.ErrRoleNotExists, err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.DeleteRoleResponse{Message: "Successfully Deleted the Role"}, nil
}

func (s *serverApi) AddUserRole(ctx context.Context, req *sso.AddUserRoleRequest) (res *sso.AddUserRoleResponse, err error) {
	user, err := s.roleService.AddUserRole(ctx, req.GetToken(), req.GetRoleId(), req.GetUserId())

	if err != nil {
		if errors.Is(storage.ErrUserAndRoleIvalid, err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(storage.ErrUserAlreadyHasTHeRole, err) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.AddUserRoleResponse{User: user}, nil

}

func (s *serverApi) RemoveUserRole(ctx context.Context, req *sso.RemoveUserRoleRequest) (res *sso.RemoveUserRoleResponse, err error) {
	user, err := s.roleService.RemoveUserRole(ctx, req.GetToken(), req.GetRoleId(), req.GetUserId())

	if err != nil {
		if errors.Is(storage.ErrUserAndRoleIvalid, err) || errors.Is(storage.ErrUserDontHaveTheRole, err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.RemoveUserRoleResponse{User: user}, nil
}

func (s *serverApi) VerifyUserRoles(ctx context.Context, req *sso.VerifyUserRolesRequest) (res *sso.VerifyUserRolesResponse, err error) {
	verified, err := s.roleService.VerifyUserRoles(ctx, req.GetRoleIds(), req.GetUserId())

	if err != nil {
		if errors.Is(storage.ErrRoleNotExists, err) || errors.Is(storage.ErrUserNotExists, err) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.VerifyUserRolesResponse{Verified: verified}, nil
}
