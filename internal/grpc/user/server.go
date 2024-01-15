package userServer

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	userService "sso_go_grpc/internal/services/user"
	"sso_go_grpc/internal/storage"
	sso "sso_go_grpc/proto/gen"
)

type serverApi struct {
	userService *userService.UserService
	sso.UnimplementedUserApiServer
}

func RegisterServer(Grpc *grpc.Server, userService *userService.UserService) {
	sso.RegisterUserApiServer(Grpc, &serverApi{userService: userService})
}
func (s *serverApi) Register(ctx context.Context, req *sso.RegisterRequest) (res *sso.RegisterResponse, err error) {

	if req.GetEmail() == "" || req.GetPassword() == "" || req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Arguments, expected: email, password, username")
	}

	token, userId, err := s.userService.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetUsername())

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

	token, userId, err := s.userService.Login(ctx, req.GetEmail(), req.GetPassword())

	if err != nil {
		if errors.Is(err, storage.ErrAuth) {
			return nil, status.Error(codes.Internal, storage.ErrAuth.Error())
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.LoginResponse{Token: token, UserId: userId}, nil
}

func (s *serverApi) GetUserById(ctx context.Context, req *sso.GetUserByIdRequest) (res *sso.GetUserByIdResponse, err error) {
	user, err := s.userService.GetUserById(ctx, req.GetUserId())

	if err != nil {
		if errors.Is(storage.ErrUserNotExists, err) {
			return nil, status.Error(codes.NotFound, storage.ErrUserNotExists.Error())
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.GetUserByIdResponse{User: user}, nil

}

func (s *serverApi) GetUserByEmail(ctx context.Context, req *sso.GetUserEmailRequest) (res *sso.GetUserEmailResponse, err error) {
	user, err := s.userService.GetUserByEmail(ctx, req.GetEmail())

	if err != nil {
		if errors.Is(storage.ErrUserNotExists, err) {
			return nil, status.Error(codes.NotFound, storage.ErrUserNotExists.Error())
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	return &sso.GetUserEmailResponse{User: user}, nil

}
