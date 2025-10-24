package auth

import (
	"context"
	"errors"
	"fmt"
	"time"
	"github.com/retail-core/auth-service/internal/user"
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
		Username:     username,
		Email:        email,
		Password: hashedPassword,
		Role:         role,
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
	fmt.Println("otp for user generated:", otp)
	return "User created successfully", nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("db error: %w", err)
	}
	if user == nil {
		return "", fmt.Errorf("invalid email or password: %w", err)
	}

	if err := CheckPassword(user.Password, password); err != nil {
		return "", errors.New("invalid email or password")
	}

	token, err := GenerateJWT(user.ID, user.Role, user.TenantID, s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}

func (s *service) Verify(ctx context.Context, email, otp string) error {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	if user == nil {
		return errors.New("invalid email or OTP")
	}

	if err := verifyOtp(user, otp); err != nil {
		return err
	}

	if err := s.repo.VerifyUser(ctx, email); err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}
	return nil
}

func verifyOtp(user *user.User, otp string) error {
	if user.OtpCode != nil && *user.OtpCode != otp {
		return errors.New("invalid email or OTP")
	}
	if user.OtpExpiresAt != nil && user.OtpExpiresAt.Before(time.Now()) {
		return errors.New("invalid OTP code")
	}
	return nil
}
