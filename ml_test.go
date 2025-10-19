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

func TestMLService_GetTagSuggestions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/servers/server-123/tag-suggestions", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := StandardResponse{
			Status:  "success",
			Message: "Tag suggestions retrieved successfully",
			Data: []TagSuggestion{
				{
					ID:           1,
					ServerID:     123,
					ServerUUID:   "server-123",
					PredictionID: "pred-001",
					TagKey:       "environment",
					TagValue:     "production",
					Confidence:   0.95,
					Reason:       "High traffic patterns consistent with production",
					Applied:      false,
					Rejected:     false,
					CreatedAt:    CustomTime{Time: time.Now()},
					UpdatedAt:    CustomTime{Time: time.Now()},
				},
				{
					ID:           2,
					ServerID:     123,
					ServerUUID:   "server-123",
					PredictionID: "pred-002",
					TagKey:       "role",
					TagValue:     "web-server",
					Confidence:   0.88,
					Reason:       "Port 80/443 activity detected",
					Applied:      false,
					Rejected:     false,
					CreatedAt:    CustomTime{Time: time.Now()},
					UpdatedAt:    CustomTime{Time: time.Now()},
				},
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

	suggestions, err := client.ML.GetTagSuggestions(context.Background(), "server-123")
	require.NoError(t, err)
	assert.Len(t, suggestions, 2)
	assert.Equal(t, "environment", suggestions[0].TagKey)
	assert.Equal(t, "production", suggestions[0].TagValue)
	assert.Equal(t, 0.95, suggestions[0].Confidence)
}

func TestMLService_ApplyTagSuggestion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/servers/server-123/tag-suggestions/apply", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "pred-001", reqBody["prediction_id"])

		response := StandardResponse{
			Status:  "success",
			Message: "Tag suggestion applied successfully",
			Data: map[string]interface{}{
				"tags_applied": 1,
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

	tagsApplied, err := client.ML.ApplyTagSuggestion(context.Background(), "server-123", "pred-001")
	require.NoError(t, err)
	assert.Equal(t, 1, tagsApplied)
}

func TestMLService_RejectTagSuggestion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/servers/server-123/tag-suggestions/pred-001/reject", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "Incorrect prediction", reqBody["feedback"])

		response := StandardResponse{
			Status:  "success",
			Message: "Tag suggestion rejected",
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

	err = client.ML.RejectTagSuggestion(context.Background(), "server-123", "pred-001", "Incorrect prediction")
	require.NoError(t, err)
}

func TestMLService_GetGroupSuggestions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/groups/suggestions", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "20", r.URL.Query().Get("limit"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Status  string            `json:"status"`
			Message string            `json:"message"`
			Data    []GroupSuggestion `json:"data"`
			Meta    *PaginationMeta   `json:"meta"`
		}{
			Status:  "success",
			Message: "Group suggestions retrieved successfully",
			Data: []GroupSuggestion{
				{
					ID:              1,
					OrganizationID:  10,
					GroupName:       "Production Web Servers",
					Description:     "High-traffic production web servers",
					ServerIDs:       []uint{1, 2, 3},
					ServerUUIDs:     []string{"srv-001", "srv-002", "srv-003"},
					Confidence:      0.92,
					Reason:          "Similar traffic patterns and configurations",
					Criteria:        []string{"traffic_volume", "port_usage", "tags"},
					Accepted:        false,
					Rejected:        false,
					CreatedAt:       CustomTime{Time: time.Now()},
					UpdatedAt:       CustomTime{Time: time.Now()},
					EstimatedBenefit: "Easier monitoring and deployment management",
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    20,
				TotalItems: 5,
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

	suggestions, meta, err := client.ML.GetGroupSuggestions(context.Background(), &PaginationOptions{Page: 1, Limit: 20})
	require.NoError(t, err)
	assert.Len(t, suggestions, 1)
	assert.Equal(t, "Production Web Servers", suggestions[0].GroupName)
	assert.Equal(t, 0.92, suggestions[0].Confidence)
	assert.NotNil(t, meta)
	assert.Equal(t, 5, meta.TotalItems)
}

func TestMLService_AcceptGroupSuggestion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/groups/suggestions/1/accept", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		groupID := uint(100)
		now := CustomTime{Time: time.Now()}
		response := StandardResponse{
			Status:  "success",
			Message: "Group suggestion accepted and group created",
			Data: &GroupSuggestion{
				ID:             1,
				GroupName:      "Production Web Servers",
				Accepted:       true,
				CreatedGroupID: &groupID,
				CreatedAt:      now,
				UpdatedAt:      now,
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

	suggestion, err := client.ML.AcceptGroupSuggestion(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, suggestion.Accepted)
	assert.NotNil(t, suggestion.CreatedGroupID)
	assert.Equal(t, uint(100), *suggestion.CreatedGroupID)
}

func TestMLService_RejectGroupSuggestion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/groups/suggestions/1/reject", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := StandardResponse{
			Status:  "success",
			Message: "Group suggestion rejected",
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

	err = client.ML.RejectGroupSuggestion(context.Background(), 1)
	require.NoError(t, err)
}

func TestMLService_ListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/ml/models", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Status  string          `json:"status"`
			Message string          `json:"message"`
			Data    []MLModel       `json:"data"`
			Meta    *PaginationMeta `json:"meta"`
		}{
			Status:  "success",
			Message: "ML models retrieved successfully",
			Data: []MLModel{
				{
					ID:        1,
					Name:      "Tag Prediction Model v2",
					ModelType: "tag_prediction",
					Version:   "2.1.0",
					Status:    "active",
					Enabled:   true,
					Accuracy:  0.89,
					Precision: 0.87,
					Recall:    0.91,
					F1Score:   0.89,
					CreatedAt: CustomTime{Time: time.Now()},
					UpdatedAt: CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    10,
				TotalItems: 3,
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

	models, meta, err := client.ML.ListModels(context.Background(), &PaginationOptions{Page: 1, Limit: 10})
	require.NoError(t, err)
	assert.Len(t, models, 1)
	assert.Equal(t, "Tag Prediction Model v2", models[0].Name)
	assert.Equal(t, 0.89, models[0].Accuracy)
	assert.NotNil(t, meta)
	assert.Equal(t, 3, meta.TotalItems)
}

func TestMLService_TrainModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/ml/models/train", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "tag_prediction", reqBody["model_type"])

		response := StandardResponse{
			Status:  "success",
			Message: "Model training initiated",
			Data: &TrainingJob{
				ID:        1,
				ModelID:   1,
				ModelType: "tag_prediction",
				Status:    "pending",
				Progress:  0,
				CreatedAt: CustomTime{Time: time.Now()},
				UpdatedAt: CustomTime{Time: time.Now()},
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

	job, err := client.ML.TrainModel(context.Background(), "tag_prediction", nil)
	require.NoError(t, err)
	assert.Equal(t, "tag_prediction", job.ModelType)
	assert.Equal(t, "pending", job.Status)
}

func TestMLService_ToggleModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/ml/models/1/toggle", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		now := CustomTime{Time: time.Now()}
		response := StandardResponse{
			Status:  "success",
			Message: "Model toggled successfully",
			Data: &MLModel{
				ID:        1,
				Name:      "Tag Prediction Model",
				Enabled:   false,
				CreatedAt: now,
				UpdatedAt: now,
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

	model, err := client.ML.ToggleModel(context.Background(), 1)
	require.NoError(t, err)
	assert.False(t, model.Enabled)
}

func TestMLService_GetModelPerformance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/ml/models/1/performance", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := StandardResponse{
			Status:  "success",
			Message: "Model performance retrieved",
			Data: &ModelPerformance{
				ModelID:           1,
				ModelType:         "tag_prediction",
				Accuracy:          0.89,
				Precision:         0.87,
				Recall:            0.91,
				F1Score:           0.89,
				PredictionsCount:  1000,
				CorrectCount:      890,
				IncorrectCount:    110,
				AverageConfidence: 0.85,
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

	perf, err := client.ML.GetModelPerformance(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 0.89, perf.Accuracy)
	assert.Equal(t, 1000, perf.PredictionsCount)
	assert.Equal(t, 890, perf.CorrectCount)
}

func TestMLService_TriggerModelTraining(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/ml/train-models", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := StandardResponse{
			Status:  "success",
			Message: "Batch training initiated",
			Data: []TrainingJob{
				{
					ID:        1,
					ModelID:   1,
					ModelType: "tag_prediction",
					Status:    "pending",
					CreatedAt: CustomTime{Time: time.Now()},
					UpdatedAt: CustomTime{Time: time.Now()},
				},
				{
					ID:        2,
					ModelID:   2,
					ModelType: "group_suggestion",
					Status:    "pending",
					CreatedAt: CustomTime{Time: time.Now()},
					UpdatedAt: CustomTime{Time: time.Now()},
				},
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

	jobs, err := client.ML.TriggerModelTraining(context.Background())
	require.NoError(t, err)
	assert.Len(t, jobs, 2)
	assert.Equal(t, "tag_prediction", jobs[0].ModelType)
	assert.Equal(t, "group_suggestion", jobs[1].ModelType)
}

func TestMLService_GetTrainingJobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/ml/training-jobs", r.URL.Path)
		assert.Equal(t, "completed", r.URL.Query().Get("status"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Status  string          `json:"status"`
			Message string          `json:"message"`
			Data    []TrainingJob   `json:"data"`
			Meta    *PaginationMeta `json:"meta"`
		}{
			Status:  "success",
			Message: "Training jobs retrieved",
			Data: []TrainingJob{
				{
					ID:        1,
					ModelID:   1,
					ModelType: "tag_prediction",
					Status:    "completed",
					Progress:  100,
					Duration:  1800,
					CreatedAt: CustomTime{Time: time.Now()},
					UpdatedAt: CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    20,
				TotalItems: 10,
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

	jobs, meta, err := client.ML.GetTrainingJobs(context.Background(), &PaginationOptions{Page: 1, Limit: 20}, "completed")
	require.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "completed", jobs[0].Status)
	assert.Equal(t, 100, jobs[0].Progress)
	assert.NotNil(t, meta)
	assert.Equal(t, 10, meta.TotalItems)
}

func TestMLService_GetAggregatedModelPerformance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/ml/model-performance", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := StandardResponse{
			Status:  "success",
			Message: "Aggregated model performance retrieved",
			Data: &ModelPerformance{
				Accuracy:          0.87,
				Precision:         0.85,
				Recall:            0.89,
				F1Score:           0.87,
				PredictionsCount:  5000,
				CorrectCount:      4350,
				IncorrectCount:    650,
				AverageConfidence: 0.83,
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

	perf, err := client.ML.GetAggregatedModelPerformance(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0.87, perf.Accuracy)
	assert.Equal(t, 5000, perf.PredictionsCount)
	assert.Equal(t, 4350, perf.CorrectCount)
}

func TestMLService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   StandardResponse
		expectedError  bool
		expectedStatus string
	}{
		{
			name:       "Unauthorized",
			statusCode: http.StatusUnauthorized,
			responseBody: StandardResponse{
				Status:  "error",
				Message: "Authentication required",
			},
			expectedError:  true,
			expectedStatus: "error",
		},
		{
			name:       "Forbidden",
			statusCode: http.StatusForbidden,
			responseBody: StandardResponse{
				Status:  "error",
				Message: "Insufficient permissions",
			},
			expectedError:  true,
			expectedStatus: "error",
		},
		{
			name:       "Not Found",
			statusCode: http.StatusNotFound,
			responseBody: StandardResponse{
				Status:  "error",
				Message: "Model not found",
			},
			expectedError:  true,
			expectedStatus: "error",
		},
		{
			name:       "Internal Server Error",
			statusCode: http.StatusInternalServerError,
			responseBody: StandardResponse{
				Status:  "error",
				Message: "Internal server error",
			},
			expectedError:  true,
			expectedStatus: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			_, err = client.ML.GetTagSuggestions(context.Background(), "server-123")
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
