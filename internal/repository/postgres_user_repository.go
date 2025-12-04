package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/database"
	"github.com/sebasr/avt-service/internal/models"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserExists is returned when a user with the same email already exists
	ErrUserExists = errors.New("user with this email already exists")
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db *database.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *database.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, email_verified,
			verification_token, verification_token_expires_at,
			reset_token, reset_token_expires_at,
			created_at, updated_at, last_login_at, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	// Generate UUID if not provided
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Set timestamps if not provided
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.EmailVerified,
		user.VerificationToken, user.VerificationTokenExpiresAt,
		user.ResetToken, user.ResetTokenExpiresAt,
		user.CreatedAt, user.UpdatedAt, user.LastLoginAt, user.IsActive,
	)

	if err != nil {
		// Check for unique constraint violation (duplicate email)
		if database.IsUniqueViolation(err) {
			return ErrUserExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT 
			id, email, password_hash, email_verified,
			verification_token, verification_token_expires_at,
			reset_token, reset_token_expires_at,
			created_at, updated_at, last_login_at, is_active
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	var verificationToken, resetToken sql.NullString
	var verificationTokenExpiresAt, resetTokenExpiresAt, lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified,
		&verificationToken, &verificationTokenExpiresAt,
		&resetToken, &resetTokenExpiresAt,
		&user.CreatedAt, &user.UpdatedAt, &lastLoginAt, &user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Handle nullable fields
	if verificationToken.Valid {
		user.VerificationToken = &verificationToken.String
	}
	if verificationTokenExpiresAt.Valid {
		user.VerificationTokenExpiresAt = &verificationTokenExpiresAt.Time
	}
	if resetToken.Valid {
		user.ResetToken = &resetToken.String
	}
	if resetTokenExpiresAt.Valid {
		user.ResetTokenExpiresAt = &resetTokenExpiresAt.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT 
			id, email, password_hash, email_verified,
			verification_token, verification_token_expires_at,
			reset_token, reset_token_expires_at,
			created_at, updated_at, last_login_at, is_active
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	var verificationToken, resetToken sql.NullString
	var verificationTokenExpiresAt, resetTokenExpiresAt, lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified,
		&verificationToken, &verificationTokenExpiresAt,
		&resetToken, &resetTokenExpiresAt,
		&user.CreatedAt, &user.UpdatedAt, &lastLoginAt, &user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Handle nullable fields
	if verificationToken.Valid {
		user.VerificationToken = &verificationToken.String
	}
	if verificationTokenExpiresAt.Valid {
		user.VerificationTokenExpiresAt = &verificationTokenExpiresAt.Time
	}
	if resetToken.Valid {
		user.ResetToken = &resetToken.String
	}
	if resetTokenExpiresAt.Valid {
		user.ResetTokenExpiresAt = &resetTokenExpiresAt.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// Update updates an existing user's information
func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET 
			email = $2,
			password_hash = $3,
			email_verified = $4,
			verification_token = $5,
			verification_token_expires_at = $6,
			reset_token = $7,
			reset_token_expires_at = $8,
			updated_at = $9,
			last_login_at = $10,
			is_active = $11
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.EmailVerified,
		user.VerificationToken, user.VerificationTokenExpiresAt,
		user.ResetToken, user.ResetTokenExpiresAt,
		user.UpdatedAt, user.LastLoginAt, user.IsActive,
	)

	if err != nil {
		if database.IsUniqueViolation(err) {
			return ErrUserExists
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdatePassword updates a user's password hash
func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, passwordHash, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateEmailVerification updates email verification status and clears verification token
func (r *PostgresUserRepository) UpdateEmailVerification(ctx context.Context, id uuid.UUID, verified bool) error {
	query := `
		UPDATE users
		SET 
			email_verified = $2,
			verification_token = NULL,
			verification_token_expires_at = NULL,
			updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, verified, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update email verification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// SetVerificationToken sets the email verification token and expiry
func (r *PostgresUserRepository) SetVerificationToken(ctx context.Context, id uuid.UUID, token string, expiresAt *time.Time) error {
	query := `
		UPDATE users
		SET 
			verification_token = $2,
			verification_token_expires_at = $3,
			updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, token, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set verification token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// SetResetToken sets the password reset token and expiry
func (r *PostgresUserRepository) SetResetToken(ctx context.Context, id uuid.UUID, token string, expiresAt *time.Time) error {
	query := `
		UPDATE users
		SET 
			reset_token = $2,
			reset_token_expires_at = $3,
			updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, token, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set reset token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetByResetToken retrieves a user by their password reset token
func (r *PostgresUserRepository) GetByResetToken(ctx context.Context, token string) (*models.User, error) {
	query := `
		SELECT
			id, email, password_hash, email_verified,
			verification_token, verification_token_expires_at,
			reset_token, reset_token_expires_at,
			created_at, updated_at, last_login_at, is_active
		FROM users
		WHERE reset_token = $1
	`

	user := &models.User{}
	var verificationToken, resetToken sql.NullString
	var verificationTokenExpiresAt, resetTokenExpiresAt, lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified,
		&verificationToken, &verificationTokenExpiresAt,
		&resetToken, &resetTokenExpiresAt,
		&user.CreatedAt, &user.UpdatedAt, &lastLoginAt, &user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by reset token: %w", err)
	}

	// Handle nullable fields
	if verificationToken.Valid {
		user.VerificationToken = &verificationToken.String
	}
	if verificationTokenExpiresAt.Valid {
		user.VerificationTokenExpiresAt = &verificationTokenExpiresAt.Time
	}
	if resetToken.Valid {
		user.ResetToken = &resetToken.String
	}
	if resetTokenExpiresAt.Valid {
		user.ResetTokenExpiresAt = &resetTokenExpiresAt.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// ClearResetToken clears the password reset token and expiry
func (r *PostgresUserRepository) ClearResetToken(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET
			reset_token = NULL,
			reset_token_expires_at = NULL,
			updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to clear reset token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login_at = $2, updated_at = $3
		WHERE id = $1
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, now, now)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
