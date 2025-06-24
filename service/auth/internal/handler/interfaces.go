package handler

import (
	"context"

	"github.com/MamangRust/monolith-point-of-sale-shared/pb"
)

type AuthHandleGrpc interface {
	pb.AuthServiceServer
	LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.ApiResponseLogin, error)
	RegisterUser(ctx context.Context, req *pb.RegisterRequest) (*pb.ApiResponseRegister, error)
}
