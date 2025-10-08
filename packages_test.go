package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackagesService_GetAvailablePackageTiers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/package/tiers", r.URL.Path)

		response := struct {
			Data    map[string]interface{} `json:"data"`
			Status  string                 `json:"status"`
			Message string                 `json:"message"`
		}{
			Data: map[string]interface{}{
				"standard": map[string]interface{}{
					"name":              "Standard",
					"max_probes":        5,
					"max_regions":       1,
					"min_frequency":     300,
					"max_alert_channels": 3,
					"max_status_pages":  1,
					"monthly_price":     29.99,
					"features": []string{
						"Basic monitoring",
						"Email alerts",
						"1 status page",
					},
				},
				"silver": map[string]interface{}{
					"name":              "Silver",
					"max_probes":        25,
					"max_regions":       3,
					"min_frequency":     60,
					"max_alert_channels": 10,
					"max_status_pages":  3,
					"monthly_price":     99.99,
					"features": []string{
						"Advanced monitoring",
						"Multi-region support",
						"Slack/PagerDuty integration",
						"3 status pages",
					},
				},
				"gold": map[string]interface{}{
					"name":              "Gold",
					"max_probes":        100,
					"max_regions":       10,
					"min_frequency":     30,
					"max_alert_channels": 50,
					"max_status_pages":  10,
					"monthly_price":     299.99,
					"features": []string{
						"Enterprise monitoring",
						"Global coverage",
						"Priority support",
						"10 status pages",
						"Custom integrations",
					},
				},
			},
			Status:  "success",
			Message: "Package tiers retrieved successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	tiers, err := client.Packages.GetAvailablePackageTiers(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, tiers)
	assert.Contains(t, tiers, "standard")
	assert.Contains(t, tiers, "silver")
	assert.Contains(t, tiers, "gold")
}

func TestPackagesService_GetOrganizationPackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organization/package", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		trialEndsAt := CustomTime{Time: time.Now().Add(14 * 24 * time.Hour)}

		response := struct {
			Data    *OrganizationPackage `json:"data"`
			Status  string               `json:"status"`
			Message string               `json:"message"`
		}{
			Data: &OrganizationPackage{
				ID:                    1,
				OrganizationID:        100,
				OrganizationUUID:      "org-uuid-123",
				PackageTier:           "standard",
				MaxProbes:             5,
				MaxRegions:            1,
				MinFrequency:          300,
				ProbeFrequencySeconds: 300,
				MaxAlertChannels:      3,
				MaxStatusPages:        1,
				AllowedProbeTypes:     []string{"HTTP", "ICMP"},
				Features:              []string{"basic_monitoring", "email_alerts"},
				SelectedRegions:       []string{"us-east-1"},
				Active:                true,
				SubscriptionStatus:    "trial",
				CurrentPeriodStart:    CustomTime{Time: time.Now()},
				CurrentPeriodEnd:      CustomTime{Time: time.Now().Add(30 * 24 * time.Hour)},
				CancelAtPeriodEnd:     false,
				TrialEndsAt:           &trialEndsAt,
				CreatedAt:             CustomTime{Time: time.Now()},
				UpdatedAt:             CustomTime{Time: time.Now()},
			},
			Status:  "success",
			Message: "Package information retrieved successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	pkg, err := client.Packages.GetOrganizationPackage(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint(100), pkg.OrganizationID)
	assert.Equal(t, "standard", pkg.PackageTier)
	assert.Equal(t, 5, pkg.MaxProbes)
	assert.Equal(t, 1, pkg.MaxRegions)
	assert.Equal(t, 300, pkg.MinFrequency)
	assert.True(t, pkg.Active)
	assert.Equal(t, "trial", pkg.SubscriptionStatus)
	assert.NotNil(t, pkg.TrialEndsAt)
	assert.Contains(t, pkg.AllowedProbeTypes, "HTTP")
	assert.Contains(t, pkg.AllowedProbeTypes, "ICMP")
}

func TestPackagesService_UpgradeOrganizationPackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/organization/package/upgrade", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody PackageUpgradeRequest
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "silver", reqBody.NewTier)
		assert.NotNil(t, reqBody.PaymentMethodID)

		response := struct {
			Data    *OrganizationPackage `json:"data"`
			Status  string               `json:"status"`
			Message string               `json:"message"`
		}{
			Data: &OrganizationPackage{
				ID:                    1,
				OrganizationID:        100,
				OrganizationUUID:      "org-uuid-123",
				PackageTier:           "silver",
				MaxProbes:             25,
				MaxRegions:            3,
				MinFrequency:          60,
				ProbeFrequencySeconds: 60,
				MaxAlertChannels:      10,
				MaxStatusPages:        3,
				AllowedProbeTypes:     []string{"HTTP", "ICMP", "TCP", "DNS"},
				Features:              []string{"advanced_monitoring", "multi_region", "slack_integration"},
				Active:                true,
				SubscriptionStatus:    "active",
				CurrentPeriodStart:    CustomTime{Time: time.Now()},
				CurrentPeriodEnd:      CustomTime{Time: time.Now().Add(30 * 24 * time.Hour)},
				CancelAtPeriodEnd:     false,
				CreatedAt:             CustomTime{Time: time.Now()},
				UpdatedAt:             CustomTime{Time: time.Now()},
			},
			Status:  "success",
			Message: "Package upgraded successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	paymentMethodID := "pm_123abc"
	pkg, err := client.Packages.UpgradeOrganizationPackage(context.Background(), &PackageUpgradeRequest{
		NewTier:         "silver",
		PaymentMethodID: &paymentMethodID,
	})
	require.NoError(t, err)
	assert.Equal(t, "silver", pkg.PackageTier)
	assert.Equal(t, 25, pkg.MaxProbes)
	assert.Equal(t, 3, pkg.MaxRegions)
	assert.Equal(t, "active", pkg.SubscriptionStatus)
	assert.Contains(t, pkg.AllowedProbeTypes, "TCP")
	assert.Contains(t, pkg.AllowedProbeTypes, "DNS")
}

func TestPackagesService_ValidateProbeConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/organization/package/validate-probe-config", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody ProbeConfigValidationRequest
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "HTTP", reqBody.ProbeType)
		assert.Equal(t, 60, reqBody.Frequency)
		assert.Len(t, reqBody.Regions, 2)

		response := struct {
			Data    *ProbeConfigValidationResult `json:"data"`
			Status  string                       `json:"status"`
			Message string                       `json:"message"`
		}{
			Data: &ProbeConfigValidationResult{
				Valid:             false,
				ProbeTypeAllowed:  true,
				FrequencyAllowed:  false,
				RegionsAllowed:    false,
				ProbeCountAllowed: true,
				Violations: []string{
					"Frequency 60s is below minimum allowed 300s",
					"Number of regions 2 exceeds maximum allowed 1",
				},
				CurrentProbeCount: 3,
				MaxProbes:         5,
				MinFrequency:      300,
				MaxRegions:        1,
				AllowedProbeTypes: []string{"HTTP", "ICMP"},
				UpgradeSuggestion: "Upgrade to Silver tier for 60s frequency and multi-region support",
			},
			Status:  "success",
			Message: "Validation completed",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Packages.ValidateProbeConfig(context.Background(), &ProbeConfigValidationRequest{
		ProbeType: "HTTP",
		Frequency: 60,
		Regions:   []string{"us-east-1", "eu-west-1"},
	})
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.ProbeTypeAllowed)
	assert.False(t, result.FrequencyAllowed)
	assert.False(t, result.RegionsAllowed)
	assert.Len(t, result.Violations, 2)
	assert.Equal(t, 300, result.MinFrequency)
	assert.Equal(t, 1, result.MaxRegions)
	assert.Contains(t, result.UpgradeSuggestion, "Silver")
}

func TestPackagesService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError bool
	}{
		{
			name:          "Unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: true,
		},
		{
			name:          "Forbidden",
			statusCode:    http.StatusForbidden,
			expectedError: true,
		},
		{
			name:          "Not Found",
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(StandardResponse{
					Status:  "error",
					Message: "Error occurred",
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			_, err = client.Packages.GetOrganizationPackage(context.Background())
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
