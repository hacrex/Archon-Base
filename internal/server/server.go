// Package server holds the control plane HTTP surface.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hacrex/Archon-Base/internal/api"
	"github.com/hacrex/Archon-Base/internal/store"
)

// Version is overridden at build time with -ldflags.
var Version = "0.0.1-dev"

const maxJSONBody = 1 << 20

type ResourceStore interface {
	CreateProject(context.Context, *api.Project) error
	GetProject(context.Context, string) (*api.Project, error)
	UpdateProject(context.Context, string, *api.Project) error
	DeleteProject(context.Context, string) error
	CreateDatabaseInstance(context.Context, *api.DatabaseInstance) error
	GetDatabaseInstance(context.Context, string, string) (*api.DatabaseInstance, error)
	UpdateDatabaseInstance(context.Context, string, string, *api.DatabaseInstance) error
	ListDatabaseInstances(context.Context, string) ([]api.DatabaseInstance, error)
	DeleteDatabaseInstance(context.Context, string, string) error
}

type Server struct {
	mux   *http.ServeMux
	store ResourceStore
}

// New creates an HTTP server. A nil store keeps health and version endpoints
// available while resource routes return 503 until persistence is configured.
func New(repositories ...ResourceStore) *Server {
	var repository ResourceStore
	if len(repositories) > 0 {
		repository = repositories[0]
	}
	s := &Server{mux: http.NewServeMux(), store: repository}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.health)
	s.mux.HandleFunc("/v1/version", s.version)
	s.mux.HandleFunc("/v1/projects", s.projects)
	s.mux.HandleFunc("/v1/projects/", s.projectRoutes)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": Version})
}

func (s *Server) projects(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/projects" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	if s.store == nil {
		serviceUnavailable(w)
		return
	}
	var project api.Project
	if err := decodeJSON(r, &project); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	project.Normalize()
	if err := project.Validate(); err != nil {
		writeValidationError(w, err)
		return
	}
	if err := s.store.CreateProject(r.Context(), &project); err != nil {
		s.writeStoreError(w, err)
		return
	}
	w.Header().Set("Location", "/v1/projects/"+project.Metadata.Name)
	writeJSON(w, http.StatusCreated, project)
}

func (s *Server) projectRoutes(w http.ResponseWriter, r *http.Request) {
	parts := routeParts(r.URL.Path)
	if len(parts) < 3 || parts[0] != "v1" || parts[1] != "projects" || parts[2] == "" {
		http.NotFound(w, r)
		return
	}
	if s.store == nil {
		serviceUnavailable(w)
		return
	}
	projectName := parts[2]
	if len(parts) == 3 {
		s.projectResource(w, r, projectName)
		return
	}
	if len(parts) == 4 && parts[3] == "databases" {
		s.databaseCollection(w, r, projectName)
		return
	}
	if len(parts) == 5 && parts[3] == "databases" {
		if parts[4] == "" {
			http.NotFound(w, r)
			return
		}
		s.databaseResource(w, r, projectName, parts[4])
		return
	}
	http.NotFound(w, r)
}

func (s *Server) projectResource(w http.ResponseWriter, r *http.Request, name string) {
	switch r.Method {
	case http.MethodGet:
		project, err := s.store.GetProject(r.Context(), name)
		if err != nil {
			s.writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, project)
	case http.MethodPatch:
		var project api.Project
		if err := decodeJSON(r, &project); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		project.Metadata.Name = name
		project.Normalize()
		if err := project.Validate(); err != nil {
			writeValidationError(w, err)
			return
		}
		if err := s.store.UpdateProject(r.Context(), name, &project); err != nil {
			s.writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, project)
	case http.MethodDelete:
		if err := s.store.DeleteProject(r.Context(), name); err != nil {
			s.writeStoreError(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPatch, http.MethodDelete)
	}
}

func (s *Server) databaseCollection(w http.ResponseWriter, r *http.Request, projectName string) {
	switch r.Method {
	case http.MethodGet:
		instances, err := s.store.ListDatabaseInstances(r.Context(), projectName)
		if err != nil {
			s.writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": instances})
	case http.MethodPost:
		var instance api.DatabaseInstance
		if err := decodeJSON(r, &instance); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		if instance.Metadata.Project != "" && instance.Metadata.Project != projectName {
			writeError(w, http.StatusBadRequest, "project_mismatch", "metadata.project must match the URL project")
			return
		}
		instance.Metadata.Project = projectName
		instance.Normalize()
		if err := instance.Validate(); err != nil {
			writeValidationError(w, err)
			return
		}
		if err := s.store.CreateDatabaseInstance(r.Context(), &instance); err != nil {
			s.writeStoreError(w, err)
			return
		}
		w.Header().Set("Location", fmt.Sprintf("/v1/projects/%s/databases/%s", projectName, instance.Metadata.Name))
		writeJSON(w, http.StatusAccepted, instance)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) databaseResource(w http.ResponseWriter, r *http.Request, projectName, name string) {
	switch r.Method {
	case http.MethodGet:
		instance, err := s.store.GetDatabaseInstance(r.Context(), projectName, name)
		if err != nil {
			s.writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, instance)
	case http.MethodPatch:
		var instance api.DatabaseInstance
		if err := decodeJSON(r, &instance); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		instance.Metadata.Project = projectName
		instance.Metadata.Name = name
		instance.Normalize()
		if err := instance.Validate(); err != nil {
			writeValidationError(w, err)
			return
		}
		if err := s.store.UpdateDatabaseInstance(r.Context(), projectName, name, &instance); err != nil {
			s.writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, instance)
	case http.MethodDelete:
		if err := s.store.DeleteDatabaseInstance(r.Context(), projectName, name); err != nil {
			s.writeStoreError(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPatch, http.MethodDelete)
	}
}

func (s *Server) writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, store.ErrProjectNotEmpty):
		writeError(w, http.StatusConflict, "project_not_empty", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "resource operation failed")
	}
}

func routeParts(path string) []string { return strings.Split(strings.Trim(path, "/"), "/") }

func decodeJSON(r *http.Request, destination any) error {
	body := io.LimitReader(r.Body, maxJSONBody)
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return errors.New("request body is required")
	}
	return api.DecodeStrict(data, destination)
}

func writeValidationError(w http.ResponseWriter, err error) {
	writeError(w, http.StatusUnprocessableEntity, "validation_failed", err.Error())
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"code": code, "error": message})
}

func serviceUnavailable(w http.ResponseWriter) {
	writeError(w, http.StatusServiceUnavailable, "store_unavailable", "resource store is not configured")
}

func methodNotAllowed(w http.ResponseWriter, methods ...string) {
	w.Header().Set("Allow", strings.Join(methods, ", "))
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not supported")
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(value)
}
