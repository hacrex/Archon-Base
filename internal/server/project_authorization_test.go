package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hacrex/Archon-Base/internal/api"
	"github.com/hacrex/Archon-Base/internal/auth"
	"github.com/hacrex/Archon-Base/internal/store"
)

type projectAuthorizationStore struct {
	fakeAuthStore
	allow bool
}

func (f *projectAuthorizationStore) AuthorizeProject(context.Context, string, string, string, auth.Role) error {
	if !f.allow {
		return store.ErrForbidden
	}
	return nil
}

func (f *projectAuthorizationStore) ListProjectsForPrincipal(context.Context, string, string) ([]api.Project, error) {
	return f.ListProjects(context.Background())
}

func TestProjectAuthorizationDeniesUnassignedProject(t *testing.T) {
	store := &projectAuthorizationStore{fakeAuthStore: fakeAuthStore{
		principal: auth.Principal{UserID: "user-1", Organization: "org-1", Role: auth.RoleDeveloper},
		token:     "token", session: auth.Session{ExpiresAt: time.Now().Add(time.Hour)},
	}}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/projects/private", nil)
	request.Header.Set("Authorization", "Bearer token")
	New(store).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestProjectCreationRequiresDeveloperOrganizationRole(t *testing.T) {
	store := &projectAuthorizationStore{fakeAuthStore: fakeAuthStore{
		principal: auth.Principal{UserID: "user-1", Organization: "org-1", Role: auth.RoleViewer},
		token:     "token", session: auth.Session{ExpiresAt: time.Now().Add(time.Hour)},
	}, allow: true}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/projects", nil)
	request.Header.Set("Authorization", "Bearer token")
	New(store).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
