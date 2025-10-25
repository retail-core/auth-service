package auth

import (
	"context"
	"database/sql"
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

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	if existing != nil {
		return "", common.ErrUserAlreadyExists
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
	logger.L().Info("OTP generated", zap.String("email", email), zap.String("otp", otp))
	return "User created successfully", nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, string, user.User, error) {
	dbUser, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", user.User{}, fmt.Errorf("db error: %w", err)
	}
	if dbUser == nil {
		return "", "", user.User{}, common.ErrInvalidLoginCredentials
	}

	if !dbUser.IsVerified {
		return "", "", user.User{}, common.ErrUserNotVerified
	}

	if err := CheckPassword(dbUser.Password, password); err != nil {
		return "", "", user.User{}, common.ErrInvalidLoginCredentials
	}

	token, err := GenerateJWT(dbUser.ID, dbUser.Role, dbUser.TenantID, dbUser.IsVerified, s.secretKey)
	if err != nil {
		return "", "", user.User{}, common.ErrInternal
	}

	rt, expiresAt, err := GenerateRefreshToken(dbUser.ID, s.secretKey)
	if err != nil {
		return "", "", user.User{}, err
	}

	err = s.repo.CreateRefreshToken(ctx, rt, dbUser.ID, expiresAt)
	if err != nil {
		return "", "", user.User{}, err
	}

	return token, rt, *dbUser, nil
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

func (s *service) ResendOTP(ctx context.Context, email string) error {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return common.ErrInternal
	}

	if user == nil {
		return common.ErrNotFound
	}

	if user.IsVerified {
		return common.ErrUserAlreadyVerified
	}

	otp, expiry, err := GenerateOTP()

	if err != nil {
		return common.ErrInternal
	}

	if err := s.repo.UpdateOtp(ctx, email, otp, expiry); err != nil {
		return common.ErrInternal
	}
	// TODO: send OTP via notification service RabbitMQ
	logger.L().Info("OTP generated", zap.String("email", email), zap.String("otp", otp))
	return nil
}

func (s *service) GenerateTokens(ctx context.Context, refreshToken string) (string, string, error) {
	rt, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	if rt.ExpiresAt.Before(time.Now()) {
		return "", "", common.ErrResourceExpired
	}

	user, err := s.repo.GetByID(ctx, rt.UserID)
	if err != nil {
		return "", "", err
	}

	// Generate new access and refresh tokens
	newAccessToken, err := GenerateJWT(user.ID, user.Role, user.TenantID, user.IsVerified, s.secretKey)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, _, err := GenerateRefreshToken(user.ID, s.secretKey)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func verifyOtp(u *user.User, otp string) error {
	now := time.Now()
	if (u.OtpCode != nil && *u.OtpCode != otp) || (u.OtpExpiresAt != nil && u.OtpExpiresAt.Before(now)) {
		return common.ErrInvalidVerificationCredential
	}
	return nil
}
