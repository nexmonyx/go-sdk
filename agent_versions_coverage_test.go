package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgentVersionsService_RegisterVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/agent/versions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Version registered",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &AgentVersionRequest{
		Version:  "v1.0.0",
		Platform: "linux",
	}
	err := client.AgentVersions.RegisterVersion(context.Background(), req)
	assert.NoError(t, err)
}

func TestAgentVersionsService_CreateVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/agent/versions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":      1,
				"version": "v1.0.0",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &AgentVersionRequest{
		Version:  "v1.0.0",
		Platform: "linux",
	}
	version, err := client.AgentVersions.CreateVersion(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, version)
}

func TestAgentVersionsService_GetVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/agent/versions/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":      1,
				"version": "v1.0.0",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	version, err := client.AgentVersions.GetVersion(context.Background(), "v1.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, version)
}

func TestAgentVersionsService_ListVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/agent/versions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": []map[string]interface{}{
				{"id": 1, "version": "v1.0.0"},
				{"id": 2, "version": "v1.0.1"},
			},
			"meta": map[string]interface{}{
				"page":  1,
				"limit": 25,
				"total": 2,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	versions, meta, err := client.AgentVersions.ListVersions(context.Background(), &ListOptions{Page: 1, Limit: 25})
	assert.NoError(t, err)
	assert.NotNil(t, versions)
	assert.NotNil(t, meta)
}

func TestAgentVersionsService_ListVersions_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{},
			"meta": map[string]interface{}{
				"page":  1,
				"limit": 25,
				"total": 0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	versions, meta, err := client.AgentVersions.ListVersions(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, versions)
	assert.NotNil(t, meta)
}

func TestAgentVersionsService_AddBinary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/binaries")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Binary added",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &AgentBinaryRequest{
		Architecture: "amd64",
		DownloadURL:  "https://example.com/agent-amd64",
	}
	err := client.AgentVersions.AddBinary(context.Background(), 1, req)
	assert.NoError(t, err)
}

func TestAgentVersionsService_AdminCreateVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/admin/agent-versions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":      1,
				"version": "v2.0.0",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	version, err := client.AgentVersions.AdminCreateVersion(context.Background(), "v2.0.0", "Admin release")
	assert.NoError(t, err)
	assert.NotNil(t, version)
}

func TestAgentVersionsService_AdminAddBinary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/admin/agent-versions/")
		assert.Contains(t, r.URL.Path, "/binaries")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Binary added",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &AgentBinaryRequest{
		Architecture: "arm64",
		DownloadURL:  "https://example.com/agent-arm64",
	}
	err := client.AgentVersions.AdminAddBinary(context.Background(), 1, req)
	assert.NoError(t, err)
}

func TestAgentVersionsService_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal server error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	// Use short timeout context to prevent retry hangs
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test error handling for all methods
	err := client.AgentVersions.RegisterVersion(ctx, &AgentVersionRequest{})
	assert.Error(t, err)

	_, err = client.AgentVersions.CreateVersion(ctx, &AgentVersionRequest{})
	assert.Error(t, err)

	_, err = client.AgentVersions.GetVersion(ctx, "v1.0.0")
	assert.Error(t, err)

	_, _, err = client.AgentVersions.ListVersions(ctx, nil)
	assert.Error(t, err)

	err = client.AgentVersions.AddBinary(ctx, 1, &AgentBinaryRequest{})
	assert.Error(t, err)

	_, err = client.AgentVersions.AdminCreateVersion(ctx, "v2.0.0", "notes")
	assert.Error(t, err)

	err = client.AgentVersions.AdminAddBinary(ctx, 1, &AgentBinaryRequest{})
	assert.Error(t, err)
}
