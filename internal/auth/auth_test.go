package auth

import (
	"testing"
	"time"
)

func TestPasswordHashAndVerify(t *testing.T) {
	if _, err := HashPassword("short"); err == nil {
		t.Fatal("short password accepted")
	}
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyPassword(hash, "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	if err := VerifyPassword(hash, "wrong password"); err != ErrInvalidCredentials {
		t.Fatalf("wrong password error = %v", err)
	}
}

func TestSessionTokenIsOpaqueAndHashable(t *testing.T) {
	token, tokenHash, err := NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	if token == tokenHash || len(token) < 32 {
		t.Fatalf("token was not separated from its hash: %q %q", token, tokenHash)
	}
	if !CompareSessionToken(tokenHash, token) {
		t.Fatal("token did not match its hash")
	}
	if CompareSessionToken(tokenHash, "not-the-token") {
		t.Fatal("wrong token matched")
	}
}

func TestSessionValidity(t *testing.T) {
	now := time.Now()
	if err := (Session{ExpiresAt: now.Add(time.Minute)}).Valid(now); err != nil {
		t.Fatal(err)
	}
	if err := (Session{ExpiresAt: now.Add(-time.Minute)}).Valid(now); err != ErrSessionExpired {
		t.Fatalf("expired session error = %v", err)
	}
	revoked := now.Add(-time.Second)
	if err := (Session{ExpiresAt: now.Add(time.Minute), RevokedAt: &revoked}).Valid(now); err != ErrSessionRevoked {
		t.Fatalf("revoked session error = %v", err)
	}
}

func TestRoleAuthorization(t *testing.T) {
	if err := Authorize(RoleAdmin, RoleOperator); err != nil {
		t.Fatal(err)
	}
	if err := Authorize(RoleDeveloper, RoleAdmin); err == nil {
		t.Fatal("developer authorized as admin")
	}
	if _, err := ParseRole("unknown"); err == nil {
		t.Fatal("unknown role accepted")
	}
}
