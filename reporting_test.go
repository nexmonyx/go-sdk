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

func TestReportingService_GenerateReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/reports/generate", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody ReportConfiguration
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "usage", reqBody.ReportType)
		assert.Equal(t, "pdf", reqBody.Format)

		response := StandardResponse{
			Status:  "success",
			Message: "Report generation started",
			Data: &Report{
				ID:             1,
				OrganizationID: 10,
				Name:           "Monthly Usage Report",
				ReportType:     reqBody.ReportType,
				Format:         reqBody.Format,
				Status:         "generating",
				CreatedAt:      CustomTime{Time: time.Now()},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	config := &ReportConfiguration{
		ReportType: "usage",
		Format:     "pdf",
		Name:       "Monthly Usage Report",
	}

	report, err := client.Reporting.GenerateReport(context.Background(), config)
	require.NoError(t, err)
	assert.Equal(t, "Monthly Usage Report", report.Name)
	assert.Equal(t, "generating", report.Status)
	assert.Equal(t, "usage", report.ReportType)
}

func TestReportingService_ListReports(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/reports", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "20", r.URL.Query().Get("limit"))
		assert.Equal(t, "completed", r.URL.Query().Get("status"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data []Report        `json:"data"`
			Meta *PaginationMeta `json:"meta"`
		}{
			Data: []Report{
				{
					ID:             1,
					OrganizationID: 10,
					Name:           "Usage Report",
					ReportType:     "usage",
					Status:         "completed",
					Format:         "pdf",
					FileURL:        "/reports/1/download",
					CreatedAt:      CustomTime{Time: time.Now()},
					CompletedAt:    &CustomTime{Time: time.Now()},
				},
				{
					ID:             2,
					OrganizationID: 10,
					Name:           "Performance Report",
					ReportType:     "performance",
					Status:         "completed",
					Format:         "csv",
					FileURL:        "/reports/2/download",
					CreatedAt:      CustomTime{Time: time.Now()},
					CompletedAt:    &CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    20,
				TotalItems: 2,
				TotalPages: 1,
			},
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

	reports, meta, err := client.Reporting.ListReports(context.Background(), &PaginationOptions{Page: 1, Limit: 20}, "completed")
	require.NoError(t, err)
	assert.Len(t, reports, 2)
	assert.Equal(t, "Usage Report", reports[0].Name)
	assert.Equal(t, "completed", reports[0].Status)
	assert.NotNil(t, meta)
	assert.Equal(t, 2, meta.TotalItems)
}

func TestReportingService_GetReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/reports/1", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := StandardResponse{
			Status:  "success",
			Message: "Report retrieved successfully",
			Data: &Report{
				ID:             1,
				OrganizationID: 10,
				Name:           "Q4 Compliance Report",
				Description:    "Quarterly compliance summary",
				ReportType:     "compliance",
				Status:         "completed",
				Format:         "pdf",
				FileURL:        "/reports/1/download",
				FileSize:       1024000,
				CreatedAt:      CustomTime{Time: time.Now()},
				CompletedAt:    &CustomTime{Time: time.Now()},
			},
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

	report, err := client.Reporting.GetReport(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Q4 Compliance Report", report.Name)
	assert.Equal(t, "completed", report.Status)
	assert.Equal(t, int64(1024000), report.FileSize)
}

func TestReportingService_DownloadReport(t *testing.T) {
	expectedContent := []byte("PDF file content here...")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/reports/1/download", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
		w.WriteHeader(http.StatusOK)
		w.Write(expectedContent)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	content, err := client.Reporting.DownloadReport(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}

func TestReportingService_ScheduleReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/reports/schedule", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody ReportSchedule
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "Weekly Usage Report", reqBody.Name)
		assert.Equal(t, "0 0 * * 1", reqBody.Schedule) // Every Monday at midnight

		nextRun := time.Now().Add(7 * 24 * time.Hour)
		response := StandardResponse{
			Status:  "success",
			Message: "Report scheduled successfully",
			Data: &ReportSchedule{
				ID:             1,
				OrganizationID: 10,
				Name:           reqBody.Name,
				Description:    reqBody.Description,
				Configuration:  reqBody.Configuration,
				Schedule:       reqBody.Schedule,
				Enabled:        true,
				NextRunAt:      &CustomTime{Time: nextRun},
				CreatedAt:      CustomTime{Time: time.Now()},
				UpdatedAt:      CustomTime{Time: time.Now()},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	schedule := &ReportSchedule{
		Name:        "Weekly Usage Report",
		Description: "Automated weekly usage summary",
		Configuration: &ReportConfiguration{
			ReportType: "usage",
			Format:     "pdf",
		},
		Schedule: "0 0 * * 1", // Every Monday at midnight
		Enabled:  true,
	}

	result, err := client.Reporting.ScheduleReport(context.Background(), schedule)
	require.NoError(t, err)
	assert.Equal(t, "Weekly Usage Report", result.Name)
	assert.True(t, result.Enabled)
	assert.NotNil(t, result.NextRunAt)
}

func TestReportingService_ListSchedules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/reports/schedules", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "20", r.URL.Query().Get("limit"))
		assert.Equal(t, "true", r.URL.Query().Get("enabled"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data []ReportSchedule `json:"data"`
			Meta *PaginationMeta  `json:"meta"`
		}{
			Data: []ReportSchedule{
				{
					ID:             1,
					OrganizationID: 10,
					Name:           "Daily Performance Report",
					Schedule:       "0 6 * * *", // Every day at 6 AM
					Enabled:        true,
					Configuration: &ReportConfiguration{
						ReportType: "performance",
						Format:     "csv",
					},
					NextRunAt: &CustomTime{Time: time.Now().Add(24 * time.Hour)},
					CreatedAt: CustomTime{Time: time.Now()},
					UpdatedAt: CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    20,
				TotalItems: 1,
				TotalPages: 1,
			},
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

	enabled := true
	schedules, meta, err := client.Reporting.ListSchedules(context.Background(), &PaginationOptions{Page: 1, Limit: 20}, &enabled)
	require.NoError(t, err)
	assert.Len(t, schedules, 1)
	assert.Equal(t, "Daily Performance Report", schedules[0].Name)
	assert.True(t, schedules[0].Enabled)
	assert.NotNil(t, meta)
	assert.Equal(t, 1, meta.TotalItems)
}

func TestReportingService_DeleteSchedule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/reports/schedules/1", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	err = client.Reporting.DeleteSchedule(context.Background(), 1)
	require.NoError(t, err)
}

func TestReportingService_ErrorHandling(t *testing.T) {
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

			_, _, err = client.Reporting.ListReports(context.Background(), nil, "")
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
