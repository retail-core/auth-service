package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/retail-core/auth-service/internal/common"
	"github.com/retail-core/auth-service/internal/logger"
	"github.com/retail-core/auth-service/internal/user"
	"go.uber.org/zap"
)

type service struct {
	repo      user.Repository
	secretKey string
}

func NewService(jwtSecret string, repo user.Repository) Service {
	return &service{
		secretKey: jwtSecret,
		repo:      repo,
	}
}

func (s *service) Register(ctx context.Context, username, email, password, role, tenantID string) (string, error) {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("db error: %w", err)
	}
	if existing != nil {
		return "", errors.New("user already exists")
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return "", errors.New("failed to hash password")
	}

	user := &user.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Role:     role,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return "", fmt.Errorf("failed to save user: %w", err)
	}

	otp, expiry, err := GenerateOTP()
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err := s.repo.UpdateOtp(ctx, email, otp, expiry); err != nil {
		return "", fmt.Errorf("failed to save OTP: %w", err)
	}

	// TODO: send OTP via notification service RabbitMQ
	logger.L().Info("OTP generated for user registration", zap.String("email", email), zap.String("otp", otp))
	return "User created successfully", nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("db error: %w", err)
	}
	if user == nil {
		return "", common.ErrInvalidLoginCredentials
	}

	if !user.IsVerified {
		return "", common.ErrUserNotVerified
	}

	if err := CheckPassword(user.Password, password); err != nil {
		return "", common.ErrInvalidLoginCredentials
	}

	token, err := GenerateJWT(user.ID, user.Role, user.TenantID, s.secretKey)
	if err != nil {
		return "", common.ErrInternal
	}
	return token, nil
}

func (s *service) Verify(ctx context.Context, email, otp string) error {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	if user == nil {
		return common.ErrInvalidVerificationCredential
	}

	if err := verifyOtp(user, otp); err != nil {
		return err
	}

	if err := s.repo.VerifyUser(ctx, email); err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}
	return nil
}

func verifyOtp(u *user.User, otp string) error {
	now := time.Now()
	if (u.OtpCode != nil && *u.OtpCode != otp) || (u.OtpExpiresAt != nil && u.OtpExpiresAt.Before(now)) {
		return common.ErrInvalidVerificationCredential
	}
	return nil
}
