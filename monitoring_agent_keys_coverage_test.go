package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonitoringAgentKeysService_CreateAdmin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/admin/monitoring-agent-keys", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"key_id":     "key-123",
				"secret_key": "secret-456",
				"full_token": "token-789",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &CreateMonitoringAgentKeyRequest{
		OrganizationID:     1,
		Description:        "Test Agent Key",
		NamespaceName:      "default",
		AgentType:          "public",
		RegionCode:         "us-east-1",
		AllowedProbeScopes: []string{"public"},
	}
	response, err := client.MonitoringAgentKeys.CreateAdmin(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "key-123", response.KeyID)
}

func TestMonitoringAgentKeysService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/organizations/")
		assert.Contains(t, r.URL.Path, "/monitoring-agent-keys")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"key_id":     "key-456",
				"secret_key": "secret-789",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &CreateMonitoringAgentKeyRequest{
		Description:        "Org Agent Key",
		NamespaceName:      "production",
		AgentType:          "private",
		AllowedProbeScopes: []string{"public", "private"},
	}
	response, err := client.MonitoringAgentKeys.Create(context.Background(), "org-123", req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestMonitoringAgentKeysService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/organizations/")
		assert.Contains(t, r.URL.Path, "/monitoring-agent-keys")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"keys": []map[string]interface{}{
				{"key_id": "key-1", "status": "active"},
				{"key_id": "key-2", "status": "active"},
			},
			"pagination": map[string]interface{}{"page": 1, "limit": 25, "total": 2},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	opts := &ListMonitoringAgentKeysOptions{Page: 1, Limit: 25, Namespace: "production"}
	keys, meta, err := client.MonitoringAgentKeys.List(context.Background(), "org-123", opts)
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, meta)
}

func TestMonitoringAgentKeysService_List_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "success",
			"keys":       []map[string]interface{}{},
			"pagination": map[string]interface{}{"page": 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.MonitoringAgentKeys.List(context.Background(), "org-123", nil)
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, meta)
}

func TestMonitoringAgentKeysService_Revoke(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/revoke")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Key revoked",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.MonitoringAgentKeys.Revoke(context.Background(), "org-123", "key-456")
	assert.NoError(t, err)
}

func TestMonitoringAgentKeysService_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	_, err := client.MonitoringAgentKeys.CreateAdmin(context.Background(), &CreateMonitoringAgentKeyRequest{})
	assert.Error(t, err)

	_, err = client.MonitoringAgentKeys.Create(context.Background(), "org-123", &CreateMonitoringAgentKeyRequest{})
	assert.Error(t, err)

	_, _, err = client.MonitoringAgentKeys.List(context.Background(), "org-123", nil)
	assert.Error(t, err)

	err = client.MonitoringAgentKeys.Revoke(context.Background(), "org-123", "key-456")
	assert.Error(t, err)
}

func TestMonitoringAgentKey_HelperMethods(t *testing.T) {
	activeKey := &MonitoringAgentKey{Status: "active", AgentType: "public"}
	assert.True(t, activeKey.IsActive())
	assert.False(t, activeKey.IsRevoked())
	assert.True(t, activeKey.IsPublic())
	assert.False(t, activeKey.IsPrivate())

	revokedKey := &MonitoringAgentKey{Status: "revoked", AgentType: "private"}
	assert.False(t, revokedKey.IsActive())
	assert.True(t, revokedKey.IsRevoked())
	assert.False(t, revokedKey.IsPublic())
	assert.True(t, revokedKey.IsPrivate())
}

func TestMonitoringAgentKey_Constructors(t *testing.T) {
	publicReq := NewPublicAgentKeyRequest("Public Agent", "default", "us-west-2")
	assert.Equal(t, "public", publicReq.AgentType)
	assert.Equal(t, "us-west-2", publicReq.RegionCode)
	assert.Contains(t, publicReq.AllowedProbeScopes, "public")

	privateReq := NewPrivateAgentKeyRequest("Private Agent", "production", "eu-west-1")
	assert.Equal(t, "private", privateReq.AgentType)
	assert.Equal(t, "eu-west-1", privateReq.RegionCode)
	assert.Contains(t, privateReq.AllowedProbeScopes, "public")
	assert.Contains(t, privateReq.AllowedProbeScopes, "private")
}

func TestListMonitoringAgentKeysOptions_ToQuery(t *testing.T) {
	enabled := true
	clusterID := uint(123)
	opts := &ListMonitoringAgentKeysOptions{
		Page:      1,
		Limit:     50,
		Namespace: "production",
		Enabled:   &enabled,
		ClusterID: &clusterID,
	}

	query := opts.ToQuery()
	assert.Equal(t, "1", query["page"])
	assert.Equal(t, "50", query["limit"])
	assert.Equal(t, "production", query["namespace"])
	assert.Equal(t, "true", query["enabled"])
	assert.Equal(t, "123", query["cluster_id"])
}

func TestMonitoringAgentKeysService_NetworkError(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://invalid-server:9999"})

	_, err := client.MonitoringAgentKeys.CreateAdmin(context.Background(), &CreateMonitoringAgentKeyRequest{})
	assert.Error(t, err)

	_, err = client.MonitoringAgentKeys.Create(context.Background(), "org-123", &CreateMonitoringAgentKeyRequest{})
	assert.Error(t, err)

	_, _, err = client.MonitoringAgentKeys.List(context.Background(), "org-123", nil)
	assert.Error(t, err)

	err = client.MonitoringAgentKeys.Revoke(context.Background(), "org-123", "key-456")
	assert.Error(t, err)
}
