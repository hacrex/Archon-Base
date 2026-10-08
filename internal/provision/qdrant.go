// Package provision contains data-plane provisioning adapters.
package provision

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hacrex/Archon-Base/internal/api"
)

type Qdrant struct {
	BaseURL       string
	Client        *http.Client
	VectorSize    int
	Distance      string
	CollectionTTL time.Duration
}

func NewQdrant(baseURL string) (*Qdrant, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid qdrant URL: %q", baseURL)
	}
	return &Qdrant{
		BaseURL:    baseURL,
		Client:     &http.Client{Timeout: 10 * time.Second},
		VectorSize: 384,
		Distance:   "Cosine",
	}, nil
}

func (q *Qdrant) Ensure(ctx context.Context, instance *api.DatabaseInstance) error {
	if instance == nil {
		return errors.New("database instance is required")
	}
	if instance.Spec.Engine != "qdrant" {
		return fmt.Errorf("unsupported provisioning engine: %s", instance.Spec.Engine)
	}
	if err := q.Health(ctx); err != nil {
		return err
	}
	payload := map[string]any{
		"vectors": map[string]any{"size": q.VectorSize, "distance": q.Distance},
	}
	status, body, err := q.request(ctx, http.MethodPut, "/collections/"+url.PathEscape(collectionName(instance)), payload)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return fmt.Errorf("qdrant collection create returned %d: %s", status, trimBody(body))
	}
	return nil
}

func (q *Qdrant) Delete(ctx context.Context, instance *api.DatabaseInstance) error {
	if instance == nil {
		return errors.New("database instance is required")
	}
	status, body, err := q.request(ctx, http.MethodDelete, "/collections/"+url.PathEscape(collectionName(instance)), nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNotFound {
		return fmt.Errorf("qdrant collection delete returned %d: %s", status, trimBody(body))
	}
	return nil
}

func (q *Qdrant) Health(ctx context.Context) error {
	status, body, err := q.request(ctx, http.MethodGet, "/readyz", nil)
	if err != nil {
		return fmt.Errorf("qdrant health check: %w", err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("qdrant is not ready: %s", trimBody(body))
	}
	return nil
}

func collectionName(instance *api.DatabaseInstance) string {
	return "archon_" + instance.Metadata.Project + "_" + instance.Metadata.Name
}

func (q *Qdrant) request(ctx context.Context, method, path string, payload any) (int, []byte, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, err
		}
		body = strings.NewReader(string(encoded))
	}
	req, err := http.NewRequestWithContext(ctx, method, q.BaseURL+path, body)
	if err != nil {
		return 0, nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := q.Client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, data, err
}

func trimBody(body []byte) string {
	value := strings.TrimSpace(string(body))
	if len(value) > 512 {
		return value[:512]
	}
	return value
}
