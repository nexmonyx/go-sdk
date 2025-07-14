package nexmonyx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricsServiceAutoSetServerUUID tests that the SDK automatically sets ServerUUID in metrics payload
func TestMetricsServiceAutoSetServerUUID(t *testing.T) {
	tests := []struct {
		name               string
		authConfig         AuthConfig
		payloadServerUUID  string
		expectedServerUUID string
		description        string
	}{
		{
			name: "auto_set_server_uuid_when_empty",
			authConfig: AuthConfig{
				ServerUUID:   "test-server-uuid-123",
				ServerSecret: "test-server-secret",
			},
			payloadServerUUID:  "",                     // Empty in payload
			expectedServerUUID: "test-server-uuid-123", // Should be auto-set from config
			description:        "Should automatically set ServerUUID from config when payload has empty ServerUUID",
		},
		{
			name: "preserve_existing_server_uuid",
			authConfig: AuthConfig{
				ServerUUID:   "config-server-uuid",
				ServerSecret: "test-server-secret",
			},
			payloadServerUUID:  "payload-server-uuid", // Already set in payload
			expectedServerUUID: "payload-server-uuid", // Should preserve the existing value
			description:        "Should preserve existing ServerUUID in payload even when config has different ServerUUID",
		},
		{
			name: "no_server_auth_no_auto_set",
			authConfig: AuthConfig{
				Token: "jwt-token", // Using JWT auth, not server auth
			},
			payloadServerUUID:  "", // Empty in payload
			expectedServerUUID: "", // Should remain empty
			description:        "Should not auto-set ServerUUID when not using server authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// For this test we don't need to capture the payload
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success"}`))
			}))
			defer server.Close()

			// Create client with test server URL
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    tt.authConfig,
			})
			require.NoError(t, err)

			// Create metrics submission
			metrics := &ComprehensiveMetricsSubmission{
				Timestamp: time.Now().Unix(),
				Hostname:  "test-host",
				Metrics: &ComprehensiveMetricsPayload{
					ServerUUID:  tt.payloadServerUUID,
					CollectedAt: time.Now().Format(time.RFC3339),
				},
			}

			// Call the method
			err = client.Metrics.SubmitComprehensiveToTimescale(context.Background(), metrics)
			assert.NoError(t, err)

			// Verify the ServerUUID was set correctly
			assert.Equal(t, tt.expectedServerUUID, metrics.Metrics.ServerUUID, tt.description)
		})
	}
}

// TestLegacyMetricsAutoSetServerUUID tests the legacy SubmitComprehensive method
func TestLegacyMetricsAutoSetServerUUID(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	// Setup client with server auth
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			ServerUUID:   "legacy-server-uuid",
			ServerSecret: "legacy-server-secret",
		},
	})
	require.NoError(t, err)

	// Create legacy metrics request
	metrics := &ComprehensiveMetricsRequest{
		ServerUUID:  "", // Empty initially
		CollectedAt: time.Now().Format(time.RFC3339),
	}

	// Call the method
	err = client.Metrics.SubmitComprehensive(context.Background(), metrics)
	assert.NoError(t, err)

	// Verify ServerUUID was auto-set
	assert.Equal(t, "legacy-server-uuid", metrics.ServerUUID)
}

// TestBackwardsCompatibility ensures the changes don't break existing code
func TestBackwardsCompatibility(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	// This test ensures that code that already sets ServerUUID continues to work
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			ServerUUID:   "config-uuid",
			ServerSecret: "config-secret",
		},
	})
	require.NoError(t, err)

	// Existing code that already sets ServerUUID
	metrics := &ComprehensiveMetricsSubmission{
		Timestamp: time.Now().Unix(),
		Hostname:  "existing-host",
		Metrics: &ComprehensiveMetricsPayload{
			ServerUUID:  "explicitly-set-uuid", // Existing code sets this
			CollectedAt: time.Now().Format(time.RFC3339),
		},
	}

	// Call the method
	err = client.Metrics.SubmitComprehensiveToTimescale(context.Background(), metrics)
	assert.NoError(t, err)

	// Verify the explicitly set UUID is preserved
	assert.Equal(t, "explicitly-set-uuid", metrics.Metrics.ServerUUID)
}
