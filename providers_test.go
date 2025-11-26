package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProvidersService_List(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/providers" {
			t.Errorf("Expected path /v1/providers, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("page") != "1" {
			t.Errorf("Expected page=1, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "20" {
			t.Errorf("Expected page_size=20, got %s", query.Get("page_size"))
		}
		if query.Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", query.Get("status"))
		}

		// Send response
		providerResp := ProviderListResponse{
			Providers: []Provider{
				{
					ID:           "provider-1",
					Name:         "AWS Production",
					ProviderType: "aws",
					Status:       "active",
					VMCount:      5,
				},
				{
					ID:           "provider-2",
					Name:         "DigitalOcean Dev",
					ProviderType: "digitalocean",
					Status:       "active",
					VMCount:      2,
				},
			},
			Total:      2,
			Page:       1,
			PageSize:   20,
			TotalPages: 1,
		}

		response := PaginatedResponse{
			Status: "success",
			Data:   providerResp,
			Meta: &PaginationMeta{
				Page:       1,
				Limit:      20,
				TotalItems: 2,
				TotalPages: 1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test list providers
	opts := &ProviderListOptions{
		Page:     1,
		PageSize: 20,
		Status:   "active",
	}

	result, _, err := client.Providers.List(context.Background(), opts)
	if err != nil {
		t.Fatalf("Failed to list providers: %v", err)
	}

	// Verify result
	if len(result.Providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(result.Providers))
	}
	if result.Providers[0].Name != "AWS Production" {
		t.Errorf("Expected provider name 'AWS Production', got %s", result.Providers[0].Name)
	}
}

func TestProvidersService_Create(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/providers" {
			t.Errorf("Expected path /v1/providers, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Parse request body
		var req ProviderCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to parse request body: %v", err)
		}

		// Verify request
		if req.Name != "AWS Production" {
			t.Errorf("Expected name 'AWS Production', got %s", req.Name)
		}
		if req.ProviderType != "aws" {
			t.Errorf("Expected provider type 'aws', got %s", req.ProviderType)
		}

		// Send response
		provider := Provider{
			ID:           "provider-123",
			Name:         req.Name,
			ProviderType: req.ProviderType,
			Status:       "active",
			VMCount:      0,
			CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		response := StandardResponse{
			Status: "success",
			Data:   provider,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test create provider
	createReq := &ProviderCreateRequest{
		Name:         "AWS Production",
		ProviderType: "aws",
		Credentials: map[string]interface{}{
			"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			"region":            "us-east-1",
		},
	}

	result, _, err := client.Providers.Create(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Verify result
	if result.ID != "provider-123" {
		t.Errorf("Expected provider ID 'provider-123', got %s", result.ID)
	}
	if result.Name != "AWS Production" {
		t.Errorf("Expected provider name 'AWS Production', got %s", result.Name)
	}
}

func TestProvidersService_Get(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/providers/provider-123" {
			t.Errorf("Expected path /v1/providers/provider-123, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Send response
		provider := Provider{
			ID:           "provider-123",
			Name:         "AWS Production",
			ProviderType: "aws",
			Status:       "active",
			VMCount:      5,
			CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		response := StandardResponse{
			Status: "success",
			Data:   provider,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test get provider
	result, err := client.Providers.Get(context.Background(), "provider-123")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}

	// Verify result
	if result.ID != "provider-123" {
		t.Errorf("Expected provider ID 'provider-123', got %s", result.ID)
	}
	if result.Name != "AWS Production" {
		t.Errorf("Expected provider name 'AWS Production', got %s", result.Name)
	}
}

func TestProvidersService_Update(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/providers/provider-123" {
			t.Errorf("Expected path /v1/providers/provider-123, got %s", r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}

		// Parse request body
		var req ProviderUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to parse request body: %v", err)
		}

		// Verify request
		if req.Name != "AWS Production Updated" {
			t.Errorf("Expected name 'AWS Production Updated', got %s", req.Name)
		}

		// Send response
		provider := Provider{
			ID:           "provider-123",
			Name:         req.Name,
			ProviderType: "aws",
			Status:       "active",
			VMCount:      5,
			CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC),
		}

		response := StandardResponse{
			Status: "success",
			Data:   provider,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test update provider
	updateReq := &ProviderUpdateRequest{
		Name: "AWS Production Updated",
	}

	result, err := client.Providers.Update(context.Background(), "provider-123", updateReq)
	if err != nil {
		t.Fatalf("Failed to update provider: %v", err)
	}

	// Verify result
	if result.Name != "AWS Production Updated" {
		t.Errorf("Expected provider name 'AWS Production Updated', got %s", result.Name)
	}
}

func TestProvidersService_Delete(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/providers/provider-123" {
			t.Errorf("Expected path /v1/providers/provider-123, got %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		// Send response
		response := StandardResponse{
			Status:  "success",
			Message: "Provider deleted successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test delete provider
	err = client.Providers.Delete(context.Background(), "provider-123")
	if err != nil {
		t.Fatalf("Failed to delete provider: %v", err)
	}
}

func TestProvidersService_Sync(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/providers/provider-123/sync" {
			t.Errorf("Expected path /v1/providers/provider-123/sync, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Send response
		syncResp := SyncResponse{
			ID:         "sync-456",
			ProviderID: "provider-123",
			Status:     "completed",
			Message:    "Sync completed successfully",
			StartedAt:  "2025-01-01T00:00:00Z",
			VMsFound:   10,
			VMsAdded:   2,
			VMsUpdated: 5,
			VMsRemoved: 1,
		}

		response := StandardResponse{
			Status: "success",
			Data:   syncResp,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test sync provider
	result, err := client.Providers.Sync(context.Background(), "provider-123")
	if err != nil {
		t.Fatalf("Failed to sync provider: %v", err)
	}

	// Verify result
	if result.Status != "completed" {
		t.Errorf("Expected sync status 'completed', got %s", result.Status)
	}
	if result.VMsFound != 10 {
		t.Errorf("Expected 10 VMs found, got %d", result.VMsFound)
	}
}
