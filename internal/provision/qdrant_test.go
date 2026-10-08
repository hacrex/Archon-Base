package provision

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hacrex/Archon-Base/internal/api"
)

func TestQdrantEnsureCreatesDeterministicCollection(t *testing.T) {
	var gotPath string
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/readyz":
			w.WriteHeader(http.StatusOK)
		case "/collections/archon_support-bot_docs-index":
			gotPath = r.URL.Path
			if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	qdrant, err := NewQdrant(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	instance := &api.DatabaseInstance{Metadata: api.Metadata{Name: "docs-index", Project: "support-bot"}, Spec: api.DatabaseSpec{Engine: "qdrant"}}
	if err := qdrant.Ensure(context.Background(), instance); err != nil {
		t.Fatal(err)
	}
	if gotPath == "" || !strings.HasSuffix(gotPath, "archon_support-bot_docs-index") {
		t.Fatalf("unexpected collection path: %s", gotPath)
	}
	if gotPayload["vectors"].(map[string]any)["size"] != float64(384) {
		t.Fatalf("unexpected vector configuration: %#v", gotPayload)
	}
}

func TestQdrantDeleteTreatsMissingCollectionAsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/archon_support-bot_docs-index" && r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	qdrant, err := NewQdrant(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	instance := &api.DatabaseInstance{Metadata: api.Metadata{Name: "docs-index", Project: "support-bot"}}
	if err := qdrant.Delete(context.Background(), instance); err != nil {
		t.Fatal(err)
	}
}
