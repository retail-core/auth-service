package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/retail-core/auth-service/internal/common"
	"github.com/retail-core/auth-service/internal/config"
	"github.com/retail-core/auth-service/internal/dtos"
	"github.com/retail-core/auth-service/internal/logger"
	"github.com/retail-core/auth-service/internal/queue"
	"github.com/retail-core/auth-service/internal/user"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
)

type service struct {
	repo      user.Repository
	config    config.Config
	publisher queue.Publisher
}

func NewService(config config.Config, repo user.Repository, publisher queue.Publisher) Service {
	return &service{
		config:    config,
		repo:      repo,
		publisher: publisher,
	}
}

func (s *service) Register(ctx context.Context, req dtos.RegisterRequest) (string, error) {
	existing, err := s.repo.GetByEmail(ctx, req.Email)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	if existing != nil {
		return "", common.ErrUserAlreadyExists
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return "", errors.New("failed to hash password")
	}

	user := &user.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     req.Role,
	}

	var userCreatedID uuid.UUID

	if userCreatedID, err = s.repo.Create(ctx, user); err != nil {
		return "", fmt.Errorf("failed to save user: %w", err)
	}

	otp, expiry, err := GenerateOTP()
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err := s.repo.UpdateOtp(ctx, req.Email, otp, expiry); err != nil {
		return "", fmt.Errorf("failed to save OTP: %w", err)
	}

	if req.Role == string(dtos.RoleBusinessOwner) {
		s.__publishSendOtpEvent(req.Email, req.Username, otp)
	} else {
		s.__publishSendInvitationEmailEvent(req.Email, req.Username, otp, req.StoreName)
	}

	if req.Role == string(dtos.RoleStaff) {
		event := map[string]any{
			"owner_id":  req.AdminID,
			"role":      req.StaffRole,
			"user_id":   userCreatedID.String(),
			"email":     req.Email,
			"user_name": req.Username,
			"store_id":  req.StoreID,
		}
		if err := s.publisher.PublishDomainEvent(ctx, dtos.StaffCreatedRoutingKey, event); err != nil {
			return "", fmt.Errorf("failed to publish staff-created domain event: %w", err)
		}

		logger.L().Info("Published staff.created event", zap.String("userId", userCreatedID.String()))
	}

	logger.L().Info("OTP generated", zap.String("email", req.Email), zap.String("otp", otp))
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

	if dbUser.IsVerified == false {
		var otp string = *dbUser.OtpCode
		var err error
		var expires time.Time

		if dbUser.OtpExpiresAt.Before(time.Now()) {
			otp, expires, err = GenerateOTP()
			if err != nil {
				return "", "", user.User{}, common.ErrInternal
			}

			if err := s.repo.UpdateOtp(ctx, dbUser.Email, otp, expires); err != nil {
				return "", "", user.User{}, fmt.Errorf("failed to save OTP: %w", err)
			}

		}
		s.__publishSendOtpEvent(dbUser.Email, dbUser.Username, otp)
		return "", "", user.User{}, common.ErrUserNotVerified
	}

	if err := CheckPassword(dbUser.Password, password); err != nil {
		return "", "", user.User{}, common.ErrInvalidLoginCredentials
	}

	token, err := GenerateJWT(dbUser.ID, dbUser.Role, dbUser.TenantID, dbUser.IsVerified, s.config.JWT_SECRET_KEY)
	if err != nil {
		return "", "", user.User{}, common.ErrInternal
	}

	rt, expiresAt, err := GenerateRefreshToken(dbUser.ID, s.config.JWT_SECRET_KEY)
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

	alreadyExistingUser := user.IsVerified

	if err := verifyOtp(user, otp); err != nil {
		return err
	}

	if err := s.repo.VerifyUser(ctx, email); err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}

	event := map[string]any{
		"ownerId":   user.ID,
		"ownerName": user.Username,
	}

	if user.Role == "business_owner" && !alreadyExistingUser {
		err = s.publisher.PublishDomainEvent(ctx, "business_owner.created", event)
		if err != nil {
			return fmt.Errorf("failed to publish domain event: %w", err)
		}
		logger.L().Info("Published business_owner.created event", zap.String("ownerId", user.ID))
	}

	if user.Role == "staff" && !alreadyExistingUser {
		event := map[string]any{
			"user_id": user.ID,
		}

		err = s.publisher.PublishDomainEvent(ctx, "staff.verified", event)
		if err != nil {
			return fmt.Errorf("failed to publish domain event: %w", err)
		}

		logger.L().Info("Published staff.verified event", zap.String("userId", user.ID))
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

	// we uncommented this line cos a user might be verified but wants to reset password
	// in which case we still need to send OTP
	// if user.IsVerified {
	// 	return common.ErrUserAlreadyVerified
	// }

	otp, expiry, err := GenerateOTP()

	if err != nil {
		return common.ErrInternal
	}

	if err := s.repo.UpdateOtp(ctx, email, otp, expiry); err != nil {
		return common.ErrInternal
	}

	s.__publishSendOtpEvent(user.Email, user.Username, otp)
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
	newAccessToken, err := GenerateJWT(user.ID, user.Role, user.TenantID, user.IsVerified, s.config.JWT_SECRET_KEY)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, expiresAt, err := GenerateRefreshToken(user.ID, s.config.JWT_SECRET_KEY)
	if err != nil {
		return "", "", err
	}

	err = s.repo.CreateRefreshToken(ctx, newRefreshToken, user.ID, expiresAt)
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

func (s *service) ResetPassword(ctx context.Context, email, newPassword string) error {
	existing, err := s.repo.GetByEmail(ctx, email)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	existing.Password = hashedPassword

	if err := s.repo.UpdatePassword(ctx, existing.ID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *service) DeleteUser(ctx context.Context, userID string) error {
	err := s.repo.DeleteUser(ctx, userID)

	if err != nil {
		logger.L().Error("Failed to delete user", zap.String("userID", userID), zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *service) __publishSendInvitationEmailEvent(email string, username string, otp string, storeName *string) error {
	var store string = ""

	if storeName != nil {
		store = *storeName
	}

	notification := map[string]any{
		"channel":  "email",
		"to":       email,
		"template": "staff-invitation",
		"data": map[string]any{
			"otp":                otp,
			"username":           username,
			"email":              email,
			"role":               "staff",
			"temporary_password": "Mypwd@1234",
			"store_name":         store,
		},
	}

	ctx := context.Background()
	if err := s.publisher.PublishNotification(ctx, "email", notification); err != nil {
		logger.L().Error("Failed to publish notification", zap.Error(err))
		return err
	}

	return nil
}

func (s *service) __publishSendOtpEvent(email, username, otp string) error {

	notification := map[string]any{
		"channel":  "email",
		"to":       email,
		"template": "otp-email",
		"data": map[string]any{
			"otp":      otp,
			"username": username,
			"email":    email,
		},
	}

	ctx := context.Background()
	if err := s.publisher.PublishNotification(ctx, "email", notification); err != nil {
		logger.L().Error("Failed to publish notification", zap.Error(err))
		return err
	}

	return nil
}

func (s *service) GetAppUpdateCheck(ctx context.Context, platform, version string) (dtos.AppUpdateCheckResponse, error) {
	v := "v" + version
	target := "v" + s.config.APP_UPDATE_CHECK_LATEST_VERSION

	if !semver.IsValid(v) || !semver.IsValid(target) {
		return dtos.AppUpdateCheckResponse{Update: dtos.UpdateNone, }, nil
	}
	
	if semver.Compare(v, target) >= 0 {
		return dtos.AppUpdateCheckResponse{Update: dtos.UpdateNone}, nil
	}

	return dtos.AppUpdateCheckResponse{
		Update:        dtos.UpdateMode(s.config.APP_UPDATE_CHECK_UPDATE),
		LatestVersion: s.config.APP_UPDATE_CHECK_LATEST_VERSION,
		UpdateURL:     s.config.APP_UPDATE_CHECK_UPDATE_URL,
	}, nil
}
