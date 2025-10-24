package user

import (
	"time"
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