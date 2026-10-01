// Package server holds the control plane HTTP surface.
// This is a skeleton: handlers return static data until the
// provisioner, function controller and storage service are wired in.
package server

import (
	"encoding/json"
	"net/http"
)

// Version is overridden at build time with -ldflags.
var Version = "0.0.1-dev"

type Server struct {
	mux *http.ServeMux
}

func New() *Server {
	s := &Server{mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.health)
	s.mux.HandleFunc("/v1/version", s.version)

	// TODO(phase-1): databases, functions, buckets, secrets, iam
	s.mux.HandleFunc("/v1/projects/", s.notImplemented)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": Version})
}

func (s *Server) notImplemented(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented yet"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
