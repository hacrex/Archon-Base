package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

type projectMemberStore struct {
	projectAuthorizationStore
	members []api.ProjectMembership
}

func (f *projectAuthorizationStore) AuthorizeProject(context.Context, string, string, string, auth.Role) error {
	if !f.allow {
		return store.ErrForbidden
	}
	return nil
}

func (f *projectAuthorizationStore) ListProjectsForPrincipal(ctx context.Context, _, _ string) ([]api.Project, error) {
	return f.ListProjects(ctx)
}

func (f *projectMemberStore) ListProjectMemberships(context.Context, string, string) ([]api.ProjectMembership, error) {
	return f.members, nil
}

func (f *projectMemberStore) UpsertProjectMembership(_ context.Context, _, _, _, userID string, role auth.Role) error {
	f.members = append(f.members, api.ProjectMembership{UserID: userID, Role: string(role)})
	return nil
}

func (f *projectMemberStore) RevokeProjectMembership(context.Context, string, string, string, string) error {
	return nil
}

func (f *projectMemberStore) GrantProjectMembership(context.Context, string, string, string, auth.Role) error {
	return nil
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

func TestProjectMemberManagement(t *testing.T) {
	store := &projectMemberStore{projectAuthorizationStore: projectAuthorizationStore{fakeAuthStore: fakeAuthStore{
		principal: auth.Principal{UserID: "owner-1", Organization: "org-1", Role: auth.RoleOwner},
		token:     "token", session: auth.Session{ExpiresAt: time.Now().Add(time.Hour)},
	}, allow: true}}
	request := httptest.NewRequest(http.MethodPut, "/v1/projects/demo/members", strings.NewReader(`{"userId":"user-2","role":"developer"}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	New(store).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || len(store.members) != 1 || store.members[0].Role != "developer" {
		t.Fatalf("assign status = %d, members = %#v, body = %s", response.Code, store.members, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/projects/demo/members", nil)
	request.Header.Set("Authorization", "Bearer token")
	response = httptest.NewRecorder()
	New(store).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "user-2") {
		t.Fatalf("list status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestProjectMemberRoleValidation(t *testing.T) {
	store := &projectMemberStore{projectAuthorizationStore: projectAuthorizationStore{fakeAuthStore: fakeAuthStore{
		principal: auth.Principal{UserID: "owner-1", Organization: "org-1", Role: auth.RoleOwner},
		token:     "token", session: auth.Session{ExpiresAt: time.Now().Add(time.Hour)},
	}, allow: true}}
	request := httptest.NewRequest(http.MethodPut, "/v1/projects/demo/members", strings.NewReader(`{"userId":"user-2","role":"invalid"}`))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	New(store).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
