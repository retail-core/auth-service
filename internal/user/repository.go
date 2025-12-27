package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/retail-core/auth-service/internal/common"
	"github.com/retail-core/auth-service/internal/db/generated"
)

type Repository interface {
	// return user model and error
	Create(ctx context.Context, user *User) (uuid.UUID, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	UpdateOtp(ctx context.Context, userID string, otpCode string, otpExpiresAt time.Time) error
	VerifyUser(ctx context.Context, email string) error
	UpdatePassword(ctx context.Context, userID string, hashedPassword string) error
	GetRefreshToken(ctx context.Context, token string) (RefreshToken, error)
	CreateRefreshToken(ctx context.Context, refreshToken, userID string, expiresAt time.Time) error
}

type pgRepository struct {
	q  *db.Queries
	db *sql.DB
}

func NewPGRepository(q *db.Queries, db *sql.DB) Repository {
	return &pgRepository{q: q, db: db}
}

func (r *pgRepository) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return common.ErrNotFound
	}

	return nil
}

func (r *pgRepository) Create(ctx context.Context, u *User) (uuid.UUID, error) {
	var tenant uuid.NullUUID
	if u.TenantID != nil && *u.TenantID != "" {
		parsed, err := uuid.Parse(*u.TenantID)
		if err == nil {
			tenant = uuid.NullUUID{UUID: parsed, Valid: true}
		}
	}

	id, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
		Role:     u.Role,
		TenantID: tenant,
	})

	return id, err
}

func (r *pgRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	dbUser, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var tenantID *string
	if dbUser.TenantID.Valid {
		s := dbUser.TenantID.UUID.String()
		tenantID = &s
	}

	return &User{
		ID:           dbUser.ID.String(),
		Email:        dbUser.Email,
		Password:     dbUser.Password,
		Role:         dbUser.Role,
		TenantID:     tenantID,
		OtpCode:      &dbUser.OtpCode.String,
		OtpExpiresAt: &dbUser.OtpExpiresAt.Time,
		IsVerified:   dbUser.IsVerified,
		Username:     dbUser.Username,
	}, nil
}

func (r *pgRepository) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	dbUser, err := r.q.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var tenantID *string
	if dbUser.TenantID.Valid {
		s := dbUser.TenantID.UUID.String()
		tenantID = &s
	}

	return &User{
		ID:           dbUser.ID.String(),
		Email:        dbUser.Email,
		Password:     dbUser.Password,
		Role:         dbUser.Role,
		TenantID:     tenantID,
		OtpCode:      &dbUser.OtpCode.String,
		OtpExpiresAt: &dbUser.OtpExpiresAt.Time,
		IsVerified:   dbUser.IsVerified,
	}, nil
}

func (r *pgRepository) UpdateOtp(ctx context.Context, email string, otpCode string, otpExpiresAt time.Time) error {
	query := `
		UPDATE users
		SET otp_code = $1, otp_expires_at = $2
		WHERE email = $3
	`
	_, err := r.db.ExecContext(ctx, query, otpCode, otpExpiresAt, email)
	return err
}

func (r *pgRepository) VerifyUser(ctx context.Context, email string) error {
	query := `
		UPDATE users
		SET is_verified = TRUE, otp_code = NULL, otp_expires_at = NULL
		WHERE email = $1
	`
	_, err := r.db.ExecContext(ctx, query, email)
	return err
}

func (r *pgRepository) CreateRefreshToken(ctx context.Context, refreshToken, userID string, expiresAt time.Time) error {
	err := r.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    uuid.MustParse(userID),
		ExpiresAt: expiresAt,
	})
	return err
}

func (r *pgRepository) GetRefreshToken(ctx context.Context, token string) (RefreshToken, error) {
	dbToken, err := r.q.GetRefreshToken(ctx, token)
	if err != nil {
		return RefreshToken{}, err
	}
	return RefreshToken{
		Token:     dbToken.Token,
		UserID:    dbToken.UserID,
		ExpiresAt: dbToken.ExpiresAt,
	}, nil
}
