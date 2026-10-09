package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hacrex/Archon-Base/internal/api"
	"github.com/hacrex/Archon-Base/internal/auth"
)

var ErrAlreadyExists = errors.New("resource already exists")

// BootstrapUser creates an active user and membership atomically. It is
// intended for first-user local setup; duplicate emails are rejected.
func (r *Repository) BootstrapUser(ctx context.Context, email, displayName, password string, role auth.Role) (api.User, api.Membership, error) {
	user := api.User{Email: strings.TrimSpace(email), DisplayName: strings.TrimSpace(displayName), Status: "active"}
	if err := user.Validate(); err != nil {
		return api.User{}, api.Membership{}, err
	}
	if _, err := auth.ParseRole(string(role)); err != nil {
		return api.User{}, api.Membership{}, err
	}
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return api.User{}, api.Membership{}, err
	}
	membership := api.Membership{Organization: r.organizationID, Role: string(role)}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return api.User{}, api.Membership{}, fmt.Errorf("begin bootstrap user: %w", err)
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO users (email, display_name, status)
		VALUES ($1, $2, 'active')
		RETURNING id::text, email, display_name, status, created_at`, user.Email, user.DisplayName).
		Scan(&user.ID, &user.Email, &user.DisplayName, &user.Status, &user.CreatedAt); err != nil {
		if isUniqueViolation(err) {
			return api.User{}, api.Membership{}, ErrAlreadyExists
		}
		return api.User{}, api.Membership{}, fmt.Errorf("create bootstrap user: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO user_credentials (user_id, password_hash) VALUES ($1::uuid, $2)`, user.ID, passwordHash); err != nil {
		return api.User{}, api.Membership{}, fmt.Errorf("create user credential: %w", err)
	}
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO organization_memberships (organization_id, user_id, role)
		VALUES ($1::uuid, $2::uuid, $3)
		RETURNING created_at`, r.organizationID, user.ID, role).Scan(&membership.CreatedAt); err != nil {
		return api.User{}, api.Membership{}, fmt.Errorf("create organization membership: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return api.User{}, api.Membership{}, fmt.Errorf("commit bootstrap user: %w", err)
	}
	return user, membership, nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key") || strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
