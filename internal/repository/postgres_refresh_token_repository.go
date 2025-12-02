package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/models"
)

var (
	// ErrRefreshTokenNotFound is returned when a refresh token is not found
	ErrRefreshTokenNotFound = errors.New("refresh token not found")

	// ErrRefreshTokenRevoked is returned when a token has been revoked
	ErrRefreshTokenRevoked = errors.New("refresh token has been revoked")
)

// PostgresRefreshTokenRepository implements RefreshTokenRepository using PostgreSQL
type PostgresRefreshTokenRepository struct {
	db *sql.DB
}

// NewPostgresRefreshTokenRepository creates a new PostgreSQL refresh token repository
func NewPostgresRefreshTokenRepository(db *sql.DB) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

// Create stores a new refresh token
func (r *PostgresRefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (
			id, user_id, token_hash, expires_at, created_at,
			revoked_at, replaced_by, user_agent, ip_address
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
		token.RevokedAt,
		token.ReplacedBy,
		token.UserAgent,
		token.IPAddress,
	)

	if err != nil {
		return fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return nil
}

// GetByHash retrieves a refresh token by its hash
func (r *PostgresRefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	query := `
		SELECT 
			id, user_id, token_hash, expires_at, created_at,
			revoked_at, replaced_by, user_agent, ip_address
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var token models.RefreshToken
	var revokedAt sql.NullTime
	var replacedBy *uuid.UUID

	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&revokedAt,
		&replacedBy,
		&token.UserAgent,
		&token.IPAddress,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, err
	}

	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}
	token.ReplacedBy = replacedBy

	// Check if token is revoked
	if token.RevokedAt != nil {
		return nil, ErrRefreshTokenRevoked
	}

	// Check if token is expired
	if token.ExpiresAt.Before(time.Now()) {
		return nil, ErrRefreshTokenNotFound
	}

	return &token, nil
}

// Revoke marks a refresh token as revoked by its ID
func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRefreshTokenNotFound
	}

	return nil
}

// RevokeByHash marks a refresh token as revoked by its hash
func (r *PostgresRefreshTokenRepository) RevokeByHash(ctx context.Context, hash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, hash)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRefreshTokenNotFound
	}

	return nil
}

// RevokeAllForUser revokes all active refresh tokens for a specific user
func (r *PostgresRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// DeleteExpired removes all expired tokens and returns the count
func (r *PostgresRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW()
	`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
