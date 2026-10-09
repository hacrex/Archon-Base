package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hacrex/Archon-Base/internal/auth"
)

type fakeAuthStore struct {
	fakeStore
	principal auth.Principal
	session   auth.Session
	token     string
	revoked   string
}

func (f *fakeAuthStore) Login(_ context.Context, email, password string, _ time.Duration) (auth.Principal, auth.Session, string, error) {
	if email != "user@example.com" || password != "correct password" {
		return auth.Principal{}, auth.Session{}, "", auth.ErrInvalidCredentials
	}
	return f.principal, f.session, f.token, nil
}

func (f *fakeAuthStore) LookupSession(_ context.Context, token string) (auth.Principal, auth.Session, error) {
	if token != f.token {
		return auth.Principal{}, auth.Session{}, auth.ErrInvalidCredentials
	}
	return f.principal, f.session, nil
}

func (f *fakeAuthStore) RevokeSession(_ context.Context, token string) error {
	if token != f.token {
		return auth.ErrInvalidCredentials
	}
	f.revoked = token
	return nil
}

func TestLoginLogoutAndMe(t *testing.T) {
	store := &fakeAuthStore{
		principal: auth.Principal{UserID: "user-1", Email: "user@example.com", Organization: "org-1", Role: auth.RoleDeveloper},
		session:   auth.Session{ID: "session-1", UserID: "user-1", ExpiresAt: time.Now().Add(time.Hour)},
		token:     "session-token",
	}
	srv := New(store)

	login := httptest.NewRecorder()
	srv.Handler().ServeHTTP(login, httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{"email":"user@example.com","password":"correct password"}`)))
	if login.Code != http.StatusOK || !strings.Contains(login.Body.String(), `"token":"session-token"`) {
		t.Fatalf("login status = %d, body = %s", login.Code, login.Body.String())
	}

	me := httptest.NewRecorder()
	meRequest := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	meRequest.Header.Set("Authorization", "Bearer session-token")
	srv.Handler().ServeHTTP(me, meRequest)
	if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), "user@example.com") {
		t.Fatalf("me status = %d, body = %s", me.Code, me.Body.String())
	}

	logout := httptest.NewRecorder()
	logoutRequest := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	logoutRequest.Header.Set("Authorization", "Bearer session-token")
	srv.Handler().ServeHTTP(logout, logoutRequest)
	if logout.Code != http.StatusOK || store.revoked != "session-token" {
		t.Fatalf("logout status = %d, revoked = %q, body = %s", logout.Code, store.revoked, logout.Body.String())
	}
}

func TestBearerMiddlewareProtectsResources(t *testing.T) {
	store := &fakeAuthStore{token: "valid", session: auth.Session{ExpiresAt: time.Now().Add(time.Hour)}}
	srv := New(store)

	unauthorized := httptest.NewRecorder()
	srv.Handler().ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/v1/projects/demo", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, body = %s", unauthorized.Code, unauthorized.Body.String())
	}

	authorized := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/projects/demo", nil)
	request.Header.Set("Authorization", "Bearer valid")
	srv.Handler().ServeHTTP(authorized, request)
	if authorized.Code == http.StatusUnauthorized {
		t.Fatalf("valid token rejected: %s", authorized.Body.String())
	}
}

var _ ResourceStore = (*fakeAuthStore)(nil)
var _ interface {
	Login(context.Context, string, string, time.Duration) (auth.Principal, auth.Session, string, error)
	LookupSession(context.Context, string) (auth.Principal, auth.Session, error)
	RevokeSession(context.Context, string) error
} = (*fakeAuthStore)(nil)
