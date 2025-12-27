package dtos

import "github.com/google/uuid"

const (
	BusinessCreatedRoutingKey = "business_owner.created"
	StaffCreatedRoutingKey    = "staff.created"
	StaffVerifiedRoutingKey   = "staff.verified"
)

type Role string

const (
	RoleBusinessOwner Role = "business_owner"
	RoleStaff         Role = "staff"
	RoleAuditor       Role = "auditor"
	RoleAdmin         Role = "admin"
)

type RegisterRequest struct {
	Username  string  `json:"username" validate:"required,min=3,max=50"`
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=8"`
	Role      string  `json:"role" validate:"required,oneof=business_owner staff auditor admin"`
	TenantID  string  `json:"tenant_id,omitempty"`
	StaffRole *string `json:"staff_role,omitempty" validate:"omitempty"` // remote one of
	AdminID   *uuid.UUID `json:"admin_id,omitempty"`                     // for staff created by admin
	StoreID   *uuid.UUID `json:"store_id,omitempty"`                     // for staff created by business owner
}

type RegisterResponse struct {
	Message string `json:"message"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         LUser  `json:"user"`
}

type LUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type VerifyRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

type VerifyResponse struct {
	Success bool `json:"success"`
}

type ResendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResendOTPResponse struct {
	Success bool `json:"success"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}