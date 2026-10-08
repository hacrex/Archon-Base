package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hacrex/Archon-Base/internal/api"
)

type fakeStore struct {
	project   *api.Project
	instance  *api.DatabaseInstance
	deleted   string
	projectDB string
}

func (f *fakeStore) CreateProject(_ context.Context, p *api.Project) error {
	f.project = p
	p.Metadata.UID = "project-1"
	return nil
}
func (f *fakeStore) GetProject(_ context.Context, name string) (*api.Project, error) {
	if f.project == nil || f.project.Metadata.Name != name {
		return nil, errors.New("not found")
	}
	return f.project, nil
}
func (f *fakeStore) UpdateProject(_ context.Context, name string, p *api.Project) error {
	p.Metadata.Name = name
	f.project = p
	return nil
}
func (f *fakeStore) DeleteProject(_ context.Context, name string) error { f.deleted = name; return nil }
func (f *fakeStore) CreateDatabaseInstance(_ context.Context, d *api.DatabaseInstance) error {
	f.instance = d
	d.Metadata.UID = "database-1"
	return nil
}
func (f *fakeStore) GetDatabaseInstance(_ context.Context, project, name string) (*api.DatabaseInstance, error) {
	if f.instance == nil || f.instance.Metadata.Project != project || f.instance.Metadata.Name != name {
		return nil, errors.New("not found")
	}
	return f.instance, nil
}
func (f *fakeStore) UpdateDatabaseInstance(_ context.Context, project, name string, d *api.DatabaseInstance) error {
	d.Metadata.Project, d.Metadata.Name = project, name
	f.instance = d
	return nil
}
func (f *fakeStore) ListDatabaseInstances(_ context.Context, project string) ([]api.DatabaseInstance, error) {
	if f.instance == nil || f.instance.Metadata.Project != project {
		return []api.DatabaseInstance{}, nil
	}
	return []api.DatabaseInstance{*f.instance}, nil
}
func (f *fakeStore) DeleteDatabaseInstance(_ context.Context, project, name string) error {
	f.projectDB = project + "/" + name
	return nil
}

func TestProjectCRUDRoutes(t *testing.T) {
	store := &fakeStore{}
	srv := New(store)
	body := `{"metadata":{"name":"support-bot"},"spec":{"displayName":"Support Bot"}}`
	request := httptest.NewRequest(http.MethodPost, "/v1/projects", strings.NewReader(body))
	response := httptest.NewRecorder()
	srv.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", response.Code, response.Body.String())
	}
	if store.project.Metadata.UID != "project-1" {
		t.Fatalf("project was not persisted: %#v", store.project)
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/projects/support-bot", nil)
	response = httptest.NewRecorder()
	srv.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodDelete, "/v1/projects/support-bot", nil)
	response = httptest.NewRecorder()
	srv.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusAccepted || store.deleted != "support-bot" {
		t.Fatalf("delete status = %d, deleted = %q", response.Code, store.deleted)
	}
}

func TestDatabaseRoutes(t *testing.T) {
	store := &fakeStore{}
	srv := New(store)
	body := `{"metadata":{"name":"docs-index"},"spec":{"engine":"qdrant","plan":"standard-2","storage":{"size":"20Gi"}}}`
	request := httptest.NewRequest(http.MethodPost, "/v1/projects/support-bot/databases", strings.NewReader(body))
	response := httptest.NewRecorder()
	srv.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("create status = %d, body = %s", response.Code, response.Body.String())
	}
	if store.instance.Metadata.Project != "support-bot" {
		t.Fatalf("project path was not applied: %#v", store.instance)
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/projects/support-bot/databases", nil)
	response = httptest.NewRecorder()
	srv.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "docs-index") {
		t.Fatalf("list status = %d, body = %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPatch, "/v1/projects/support-bot/databases/docs-index", strings.NewReader(body))
	response = httptest.NewRecorder()
	srv.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("patch status = %d, body = %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodDelete, "/v1/projects/support-bot/databases/docs-index", nil)
	response = httptest.NewRecorder()
	srv.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusAccepted || store.projectDB != "support-bot/docs-index" {
		t.Fatalf("delete status = %d, deleted = %q", response.Code, store.projectDB)
	}
}

func TestHandlersRejectUnknownFields(t *testing.T) {
	srv := New(&fakeStore{})
	request := httptest.NewRequest(http.MethodPost, "/v1/projects", strings.NewReader(`{"metadata":{"name":"demo"},"spec":{"displayName":"Demo","unexpected":true}}`))
	response := httptest.NewRecorder()
	srv.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestRoutesWithoutStoreReturnUnavailable(t *testing.T) {
	response := httptest.NewRecorder()
	srv := New()
	srv.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/projects/demo", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestHealthReadinessAndRequestID(t *testing.T) {
	srv := New(&fakeStore{})
	for _, path := range []string{"/healthz", "/livez", "/readyz"} {
		response := httptest.NewRecorder()
		srv.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d, body = %s", path, response.Code, response.Body.String())
		}
		if response.Header().Get("X-Request-ID") == "" {
			t.Fatalf("%s did not return a request ID", path)
		}
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/livez", nil)
	request.Header.Set("X-Request-ID", "client-request-123")
	srv.Handler().ServeHTTP(response, request)
	if response.Header().Get("X-Request-ID") != "client-request-123" {
		t.Fatalf("request ID was not preserved: %q", response.Header().Get("X-Request-ID"))
	}
}

func TestValidationErrorHasStableFields(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/projects", strings.NewReader(`{"metadata":{"name":"INVALID"},"spec":{}}`))
	New(&fakeStore{}).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, field := range []string{`"code":"validation_failed"`, `"requestId"`, `"fields"`, `"field":"metadata.name"`} {
		if !strings.Contains(body, field) {
			t.Fatalf("validation response missing %s: %s", field, body)
		}
	}
}
