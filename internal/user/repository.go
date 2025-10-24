package user

import (
	"context"
	"database/sql"
	"time"
	"github.com/retail-core/auth-service/internal/db/generated"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdateOtp(ctx context.Context, userID string, otpCode string, otpExpiresAt time.Time) error
	VerifyUser(ctx context.Context, email string) error
}


type pgRepository struct {
	q *db.Queries
	db *sql.DB
}

func NewPGRepository(q *db.Queries, db *sql.DB) Repository {
	return &pgRepository{q: q, db: db}
}

func (r *pgRepository) Create(ctx context.Context, u *User) error {
	var tenant uuid.NullUUID
	if u.TenantID != nil && *u.TenantID != "" {
		parsed, err := uuid.Parse(*u.TenantID)
		if err == nil {
			tenant = uuid.NullUUID{UUID: parsed, Valid: true}
		}
	}

	return r.q.CreateUser(ctx, db.CreateUserParams{
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
		Role:     u.Role,
		TenantID: tenant,
	})
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
		ID:          dbUser.ID.String(),
		Email:       dbUser.Email,
		Password:    dbUser.Password,
		Role:        dbUser.Role,
		TenantID:    tenantID,
		OtpCode: 	 &dbUser.OtpCode.String,
		OtpExpiresAt: &dbUser.OtpExpiresAt.Time,
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