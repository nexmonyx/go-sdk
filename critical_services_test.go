package nexmonyx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCriticalServicesAccessible verifies that the critical services needed by monitoring-controller are accessible
func TestCriticalServicesAccessible(t *testing.T) {
	config := &Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			Token: "test-token",
		},
	}

	client, err := NewClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify Incidents service is accessible
	assert.NotNil(t, client.Incidents, "IncidentsService should be accessible")
	
	// Verify Probes service is accessible
	assert.NotNil(t, client.Probes, "ProbesService should be accessible")
	
	// Verify both services have the client reference
	assert.Equal(t, client, client.Incidents.client, "IncidentsService should have client reference")
	assert.Equal(t, client, client.Probes.client, "ProbesService should have client reference")
}