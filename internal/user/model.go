package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           string
	Username     string
	Email        string
	Password string
	OtpCode      *string
	OtpExpiresAt *time.Time
	IsVerified   bool
	Role         string
	TenantID     *string
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}