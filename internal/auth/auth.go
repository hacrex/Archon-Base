// Package auth contains authentication and authorization primitives for the control plane.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionExpired     = errors.New("session expired")
	ErrSessionRevoked     = errors.New("session revoked")
	ErrForbidden          = errors.New("forbidden")
)

type Role string

const (
	RoleOwner     Role = "owner"
	RoleAdmin     Role = "admin"
	RoleOperator  Role = "operator"
	RoleDeveloper Role = "developer"
	RoleViewer    Role = "viewer"
)

type Principal struct {
	UserID       string
	Email        string
	Organization string
	Role         Role
}

func HashPassword(password string) (string, error) {
	if len(password) < 12 {
		return "", errors.New("password must contain at least 12 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func VerifyPassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

// NewSessionToken returns the plaintext token once and its database-safe hash.
// Only the hash should be persisted.
func NewSessionToken() (token, tokenHash string, err error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", "", fmt.Errorf("generate session token: %w", err)
	}
	token = hex.EncodeToString(raw[:])
	hash := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(hash[:]), nil
}

func HashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func CompareSessionToken(expectedHash, token string) bool {
	actual := HashSessionToken(token)
	return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(actual)) == 1
}

func Authorize(actual, required Role) error {
	if roleRank(actual) < roleRank(required) {
		return fmt.Errorf("%w: role %s requires %s", ErrForbidden, actual, required)
	}
	return nil
}

func ParseRole(value string) (Role, error) {
	role := Role(strings.TrimSpace(value))
	if roleRank(role) == 0 {
		return "", fmt.Errorf("invalid role %q", value)
	}
	return role, nil
}

func roleRank(role Role) int {
	switch role {
	case RoleViewer:
		return 1
	case RoleDeveloper:
		return 2
	case RoleOperator:
		return 3
	case RoleAdmin:
		return 4
	case RoleOwner:
		return 5
	default:
		return 0
	}
}

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func (s Session) Valid(now time.Time) error {
	if s.RevokedAt != nil {
		return ErrSessionRevoked
	}
	if !now.Before(s.ExpiresAt) {
		return ErrSessionExpired
	}
	return nil
}
