package auth

import (
	"context"

	"github.com/retail-core/auth-service/internal/dtos"
	"github.com/retail-core/auth-service/internal/user"
)

type Service interface {
	Register(ctx context.Context, req dtos.RegisterRequest) (userId string, err error)
	Login(ctx context.Context, email, password string) (token, refreshToken string, user user.User, err error)
	Verify(ctx context.Context, email, code string) error
	ResendOTP(ctx context.Context, email string) error
	GenerateTokens(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error)
	ResetPassword(ctx context.Context, email, newPassword string) error
}