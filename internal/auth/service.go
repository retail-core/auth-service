package auth

import (
	"context"
)

type Service interface {
	Register(ctx context.Context, username, email, password, role, tenantID string) (userId string, err error)
	Login(ctx context.Context, email, password string) (token string, err error)
	Verify(ctx context.Context, email, code string) error
}