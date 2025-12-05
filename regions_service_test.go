package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegionsService_List(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/regions" {
			t.Errorf("expected path /v1/regions, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		response := map[string]interface{}{
			"status": "success",
			"data": []map[string]string{
				{
					"code":      "us-east-1",
					"name":      "US East (Virginia)",
					"country":   "United States",
					"continent": "North America",
				},
				{
					"code":      "eu-west-1",
					"name":      "Europe (Ireland)",
					"country":   "Ireland",
					"continent": "Europe",
				},
				{
					"code":      "ap-southeast-1",
					"name":      "Asia Pacific (Singapore)",
					"country":   "Singapore",
					"continent": "Asia",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Call List
	regions, err := client.Regions.List(context.Background())
	if err != nil {
		t.Fatalf("failed to list regions: %v", err)
	}

	// Verify results
	if len(regions) != 3 {
		t.Errorf("expected 3 regions, got %d", len(regions))
	}

	// Verify first region
	if regions[0].Code != "us-east-1" {
		t.Errorf("expected code 'us-east-1', got %s", regions[0].Code)
	}
	if regions[0].Name != "US East (Virginia)" {
		t.Errorf("expected name 'US East (Virginia)', got %s", regions[0].Name)
	}
	if regions[0].Country != "United States" {
		t.Errorf("expected country 'United States', got %s", regions[0].Country)
	}
	if regions[0].Continent != "North America" {
		t.Errorf("expected continent 'North America', got %s", regions[0].Continent)
	}

	// Verify second region
	if regions[1].Code != "eu-west-1" {
		t.Errorf("expected code 'eu-west-1', got %s", regions[1].Code)
	}

	// Verify third region
	if regions[2].Code != "ap-southeast-1" {
		t.Errorf("expected code 'ap-southeast-1', got %s", regions[2].Code)
	}
}

func TestRegionsService_List_Error(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "internal server error",
		})
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Call List - should return error
	_, err = client.Regions.List(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestRegionsService_List_EmptyResponse(t *testing.T) {
	// Create a mock server that returns empty data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status": "success",
			"data":   []map[string]string{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Call List
	regions, err := client.Regions.List(context.Background())
	if err != nil {
		t.Fatalf("failed to list regions: %v", err)
	}

	// Verify empty result
	if len(regions) != 0 {
		t.Errorf("expected 0 regions, got %d", len(regions))
	}
}
