package auth

import (
	"context"
	"errors"
	"sso/internal/services/auth"

	authv1 "github.com/XaCaMaCa/protos/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appId int64,
	) (string, error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (int64, error)
	IsAdmin(
		ctx context.Context,
		userId int64,
	) (bool, error)
}

// serverAPI реализует gRPC сервер авторизации
type serverAPI struct {
	authv1.UnimplementedAuthServiceServer
	auth Auth
}

func Register(grpc *grpc.Server, auth Auth) {
	authv1.RegisterAuthServiceServer(grpc, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {
	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int64(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Errorf(codes.Internal, "failed to login")
	}
	return &authv1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {
	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}
	userId, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Errorf(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to register")
	}
	return &authv1.RegisterResponse{UserId: userId}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *authv1.IsAdminRequest,
) (*authv1.IsAdminResponse, error) {
	if err := validateIsAdminRequest(req); err != nil {
		return nil, err
	}
	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to check admin status")
	}
	return &authv1.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func validateLoginRequest(req *authv1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Errorf(codes.InvalidArgument, "email is required")
	}
	if req.GetPassword() == "" {
		return status.Errorf(codes.InvalidArgument, "password is required")
	}
	if req.GetAppId() == 0 {
		return status.Errorf(codes.InvalidArgument, "app_id is required")
	}
	return nil
}
func validateRegisterRequest(req *authv1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Errorf(codes.InvalidArgument, "email is required")
	}
	if req.GetPassword() == "" {
		return status.Errorf(codes.InvalidArgument, "password is required")
	}
	return nil
}
func validateIsAdminRequest(req *authv1.IsAdminRequest) error {
	if req.GetUserId() == 0 {
		return status.Errorf(codes.InvalidArgument, "user_id is required")
	}
	return nil
}
