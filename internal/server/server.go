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
	"github.com/hacrex/Archon-Base/internal/auth"
	"github.com/hacrex/Archon-Base/internal/provision"
	"github.com/hacrex/Archon-Base/internal/store"
)

// Version is overridden at build time with -ldflags.
var Version = "0.0.1-dev"

const maxJSONBody = 1 << 20
const requestIDHeader = "X-Request-ID"

type ResourceStore interface {
	CreateProject(context.Context, *api.Project) error
	ListProjects(context.Context) ([]api.Project, error)
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

type statusStore interface {
	UpdateDatabaseInstanceStatus(context.Context, string, string, string, int64) error
}

type authenticationStore interface {
	store.AuthRepository
}

type Server struct {
	mux         *http.ServeMux
	store       ResourceStore
	provisioner *provision.Qdrant
	auth        authenticationStore
}

// New creates an HTTP server. A nil store keeps health and version endpoints
// available while resource routes return 503 until persistence is configured.
func New(repositories ...ResourceStore) *Server {
	var repository ResourceStore
	if len(repositories) > 0 {
		repository = repositories[0]
	}
	s := &Server{mux: http.NewServeMux(), store: repository}
	if authStore, ok := repository.(authenticationStore); ok {
		s.auth = authStore
	}
	s.routes()
	return s
}

// NewWithProvisioner creates an API server with the optional Qdrant adapter.
func NewWithProvisioner(repository ResourceStore, qdrant *provision.Qdrant) *Server {
	s := &Server{mux: http.NewServeMux(), store: repository, provisioner: qdrant}
	if authStore, ok := repository.(authenticationStore); ok {
		s.auth = authStore
	}
	s.routes()
	return s
}

// Handler returns the server with request ID and structured access-log middleware.
func (s *Server) Handler() http.Handler {
	return requestLogging(withRequestID(cors(s.authenticate(s.mux))))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://127.0.0.1:4174" || origin == "http://localhost:4174" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.livez)
	s.mux.HandleFunc("/livez", s.livez)
	s.mux.HandleFunc("/readyz", s.readyz)
	s.mux.HandleFunc("/v1/version", s.version)
	s.mux.HandleFunc("/v1/auth/login", s.login)
	s.mux.HandleFunc("/v1/auth/logout", s.logout)
	s.mux.HandleFunc("/v1/auth/me", s.me)
	s.mux.HandleFunc("/v1/projects", s.projects)
	s.mux.HandleFunc("/v1/projects/", s.projectRoutes)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token        string    `json:"token"`
	ExpiresAt    time.Time `json:"expiresAt"`
	UserID       string    `json:"userId"`
	Email        string    `json:"email"`
	Organization string    `json:"organizationId"`
	Role         auth.Role `json:"role"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	if s.auth == nil {
		serviceUnavailable(w)
		return
	}
	var request loginRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	principal, session, token, err := s.auth.Login(r.Context(), request.Email, request.Password, 24*time.Hour)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "email or password is invalid")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "login failed")
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{Token: token, ExpiresAt: session.ExpiresAt, UserID: principal.UserID, Email: principal.Email, Organization: principal.Organization, Role: principal.Role})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	token, ok := bearerToken(r)
	if !ok || s.auth == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "a bearer token is required")
		return
	}
	if err := s.auth.RevokeSession(r.Context(), token); err != nil && !errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, http.StatusInternalServerError, "internal_error", "logout failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "signed_out"})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "a valid bearer token is required")
		return
	}
	writeJSON(w, http.StatusOK, principal)
}

type principalContextKey struct{}

func principalFromContext(ctx context.Context) (auth.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(auth.Principal)
	return principal, ok
}

func bearerToken(r *http.Request) (string, bool) {
	parts := strings.Fields(strings.TrimSpace(r.Header.Get("Authorization")))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.auth == nil || r.URL.Path == "/healthz" || r.URL.Path == "/livez" || r.URL.Path == "/readyz" || r.URL.Path == "/v1/version" || r.URL.Path == "/v1/auth/login" {
			next.ServeHTTP(w, r)
			return
		}
		token, ok := bearerToken(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "a bearer token is required")
			return
		}
		principal, _, err := s.auth.LookupSession(r.Context(), token)
		if err != nil {
			if errors.Is(err, auth.ErrSessionExpired) || errors.Is(err, auth.ErrSessionRevoked) || errors.Is(err, auth.ErrInvalidCredentials) {
				writeError(w, http.StatusUnauthorized, "unauthorized", "the bearer token is invalid or expired")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", "authentication lookup failed")
			return
		}
		ctx := context.WithValue(r.Context(), principalContextKey{}, principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
	if s.store == nil {
		serviceUnavailable(w)
		return
	}
	if r.Method == http.MethodGet {
		projects, err := s.store.ListProjects(r.Context())
		if err != nil {
			s.writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": projects})
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodGet, http.MethodPost)
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
		if s.provisioner != nil {
			phase, message := "Ready", "Qdrant collection is ready"
			if err := s.provisioner.Ensure(r.Context(), &instance); err != nil {
				phase, message = "Degraded", err.Error()
			}
			if statuses, ok := s.store.(statusStore); ok {
				if err := statuses.UpdateDatabaseInstanceStatus(r.Context(), instance.Metadata.UID, phase, message, instance.Metadata.Generation); err != nil {
					writeError(w, http.StatusInternalServerError, "status_update_failed", "database instance was created but its status could not be updated")
					return
				}
			}
			instance.Status.Phase, instance.Status.Message = phase, message
			instance.Status.ObservedGeneration = instance.Metadata.Generation
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
