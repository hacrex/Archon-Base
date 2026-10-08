// Package server holds the control plane HTTP surface.
package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hacrex/Archon-Base/internal/api"
	"github.com/hacrex/Archon-Base/internal/store"
)

// Version is overridden at build time with -ldflags.
var Version = "0.0.1-dev"

const maxJSONBody = 1 << 20
const requestIDHeader = "X-Request-ID"

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

type readinessChecker interface {
	Ready(context.Context) error
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

// Handler returns the server with request ID and structured access-log middleware.
func (s *Server) Handler() http.Handler {
	return requestLogging(withRequestID(s.mux))
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.livez)
	s.mux.HandleFunc("/livez", s.livez)
	s.mux.HandleFunc("/readyz", s.readyz)
	s.mux.HandleFunc("/v1/version", s.version)
	s.mux.HandleFunc("/v1/projects", s.projects)
	s.mux.HandleFunc("/v1/projects/", s.projectRoutes)
}

func (s *Server) livez(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store_unavailable", "resource store is not configured")
		return
	}
	if checker, ok := s.store.(readinessChecker); ok {
		if err := checker.Ready(r.Context()); err != nil {
			writeError(w, http.StatusServiceUnavailable, "dependency_unavailable", "resource store is not ready")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
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
	body := io.LimitReader(r.Body, maxJSONBody+1)
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	if len(data) > maxJSONBody {
		return errors.New("request body exceeds 1 MiB limit")
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return errors.New("request body is required")
	}
	return api.DecodeStrict(data, destination)
}

func writeValidationError(w http.ResponseWriter, err error) {
	var fields api.ValidationErrors
	if errors.As(err, &fields) {
		writeErrorFields(w, http.StatusUnprocessableEntity, "validation_failed", "request validation failed", fields)
		return
	}
	writeError(w, http.StatusUnprocessableEntity, "validation_failed", err.Error())
}

type errorResponse struct {
	Code      string           `json:"code"`
	Error     string           `json:"error"`
	RequestID string           `json:"requestId,omitempty"`
	Fields    []api.FieldError `json:"fields,omitempty"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeErrorFields(w, status, code, message, nil)
}

func writeErrorFields(w http.ResponseWriter, status int, code, message string, fields []api.FieldError) {
	writeJSON(w, status, errorResponse{Code: code, Error: message, RequestID: requestIDFrom(w), Fields: fields})
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

type requestIDKey struct{}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get(requestIDHeader))
		if id == "" || len(id) > 128 {
			id = newRequestID()
		}
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

func requestIDFrom(w http.ResponseWriter) string { return w.Header().Get(requestIDHeader) }

func newRequestID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(raw[:])
}

type loggingWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *loggingWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += n
	return n, err
}

func requestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		lw := &loggingWriter{ResponseWriter: w}
		next.ServeHTTP(lw, r)
		status := lw.status
		if status == 0 {
			status = http.StatusOK
		}
		log.Printf(`{"request_id":%q,"method":%q,"path":%q,"status":%d,"bytes":%d,"duration_ms":%d}`,
			r.Header.Get(requestIDHeader), r.Method, r.URL.Path, status, lw.bytes, time.Since(started).Milliseconds())
	})
}
