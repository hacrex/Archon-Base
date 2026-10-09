package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hacrex/Archon-Base/internal/auth"
)

// AuthRepository is the authentication and organization-membership contract
// consumed by the HTTP authentication middleware.
type AuthRepository interface {
	Login(context.Context, string, string, time.Duration) (auth.Principal, auth.Session, string, error)
	LookupSession(context.Context, string) (auth.Principal, auth.Session, error)
	RevokeSession(context.Context, string) error
}

// Login verifies an active user's password and one membership in this
// repository's configured organization, then creates a revocable session.
func (r *Repository) Login(ctx context.Context, email, password string, lifetime time.Duration) (auth.Principal, auth.Session, string, error) {
	if strings.TrimSpace(email) == "" || password == "" {
		return auth.Principal{}, auth.Session{}, "", auth.ErrInvalidCredentials
	}
	if lifetime <= 0 || lifetime > 30*24*time.Hour {
		lifetime = 24 * time.Hour
	}
	var principal auth.Principal
	var passwordHash string
	var status string
	if err := r.db.QueryRowContext(ctx, `
		SELECT u.id::text, u.email, u.status, c.password_hash, m.organization_id::text, m.role
		FROM users u
		JOIN user_credentials c ON c.user_id = u.id
		JOIN organization_memberships m ON m.user_id = u.id
		WHERE lower(u.email) = lower($1) AND m.organization_id = $2::uuid`, email, r.organizationID).
		Scan(&principal.UserID, &principal.Email, &status, &passwordHash, &principal.Organization, &principal.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.Principal{}, auth.Session{}, "", auth.ErrInvalidCredentials
		}
		return auth.Principal{}, auth.Session{}, "", fmt.Errorf("find login identity: %w", err)
	}
	if status != "active" {
		return auth.Principal{}, auth.Session{}, "", auth.ErrInvalidCredentials
	}
	if err := auth.VerifyPassword(passwordHash, password); err != nil {
		return auth.Principal{}, auth.Session{}, "", auth.ErrInvalidCredentials
	}
	token, tokenHash, err := auth.NewSessionToken()
	if err != nil {
		return auth.Principal{}, auth.Session{}, "", err
	}
	session := auth.Session{UserID: principal.UserID, ExpiresAt: time.Now().UTC().Add(lifetime)}
	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1::uuid, $2, $3)
		RETURNING id::text, expires_at`, principal.UserID, tokenHash, session.ExpiresAt).
		Scan(&session.ID, &session.ExpiresAt); err != nil {
		return auth.Principal{}, auth.Session{}, "", fmt.Errorf("create session: %w", err)
	}
	return principal, session, token, nil
}

func (r *Repository) LookupSession(ctx context.Context, token string) (auth.Principal, auth.Session, error) {
	if strings.TrimSpace(token) == "" {
		return auth.Principal{}, auth.Session{}, auth.ErrInvalidCredentials
	}
	var principal auth.Principal
	var session auth.Session
	var revokedAt sql.NullTime
	var status string
	if err := r.db.QueryRowContext(ctx, `
		SELECT s.id::text, s.user_id::text, s.expires_at, s.revoked_at,
		       u.email, u.status, m.organization_id::text, m.role
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		JOIN organization_memberships m ON m.user_id = s.user_id
		WHERE s.token_hash = $1 AND m.organization_id = $2::uuid`, auth.HashSessionToken(token), r.organizationID).
		Scan(&session.ID, &session.UserID, &session.ExpiresAt, &revokedAt, &principal.Email, &status, &principal.Organization, &principal.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.Principal{}, auth.Session{}, auth.ErrInvalidCredentials
		}
		return auth.Principal{}, auth.Session{}, fmt.Errorf("lookup session: %w", err)
	}
	principal.UserID = session.UserID
	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}
	if status != "active" {
		return auth.Principal{}, auth.Session{}, auth.ErrInvalidCredentials
	}
	if err := session.Valid(time.Now().UTC()); err != nil {
		return auth.Principal{}, auth.Session{}, err
	}
	if _, err := r.db.ExecContext(ctx, `UPDATE sessions SET last_seen_at = now() WHERE id = $1::uuid`, session.ID); err != nil {
		return auth.Principal{}, auth.Session{}, fmt.Errorf("touch session: %w", err)
	}
	return principal, session, nil
}

func (r *Repository) RevokeSession(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return auth.ErrInvalidCredentials
	}
	result, err := r.db.ExecContext(ctx, `UPDATE sessions SET revoked_at = COALESCE(revoked_at, now()) WHERE token_hash = $1`, auth.HashSessionToken(token))
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return auth.ErrInvalidCredentials
	}
	return nil
}
