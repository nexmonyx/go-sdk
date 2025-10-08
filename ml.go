package nexmonyx

import (
	"context"
	"fmt"
)

// MLService handles machine learning operations
// Provides tag suggestions, group suggestions, model management, and training jobs
type MLService struct {
	client *Client
}

// Tag Suggestion Methods

// GetTagSuggestions retrieves ML-generated tag suggestions for a server
// Authentication: JWT Token required
// Endpoint: GET /v1/servers/{serverID}/tag-suggestions
// Parameters:
//   - serverID: Server ID or UUID
// Returns: Array of tag predictions with confidence scores
func (s *MLService) GetTagSuggestions(ctx context.Context, serverID string) ([]TagSuggestion, error) {
	var resp struct {
		Data    []TagSuggestion `json:"data"`
		Status  string          `json:"status"`
		Message string          `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/servers/%s/tag-suggestions", serverID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ApplyTagSuggestion applies ML-suggested tags to a server
// Authentication: JWT Token required
// Endpoint: POST /v1/servers/{serverID}/tag-suggestions/apply
// Parameters:
//   - serverID: Server ID or UUID
//   - predictionID: Prediction ID to apply (optional, applies all if empty)
// Returns: Number of tags applied
func (s *MLService) ApplyTagSuggestion(ctx context.Context, serverID string, predictionID string) (int, error) {
	var resp struct {
		Data struct {
			TagsApplied int `json:"tags_applied"`
		} `json:"data"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	body := map[string]interface{}{}
	if predictionID != "" {
		body["prediction_id"] = predictionID
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/servers/%s/tag-suggestions/apply", serverID),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return 0, err
	}

	return resp.Data.TagsApplied, nil
}

// RejectTagSuggestion rejects an ML tag suggestion with optional feedback
// Authentication: JWT Token required
// Endpoint: POST /v1/servers/{serverID}/tag-suggestions/{predictionID}/reject
// Parameters:
//   - serverID: Server ID or UUID
//   - predictionID: Prediction ID to reject
//   - feedback: Optional feedback for model improvement
// Returns: Success confirmation
func (s *MLService) RejectTagSuggestion(ctx context.Context, serverID string, predictionID string, feedback string) error {
	body := map[string]interface{}{}
	if feedback != "" {
		body["feedback"] = feedback
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/servers/%s/tag-suggestions/%s/reject", serverID, predictionID),
		Body:   body,
	})
	return err
}

// Group Suggestion Methods

// GetGroupSuggestions retrieves ML-generated server grouping suggestions
// Authentication: JWT Token required
// Endpoint: GET /v1/groups/suggestions
// Parameters:
//   - opts: Optional pagination and filtering options
// Returns: Array of group suggestions with confidence scores
func (s *MLService) GetGroupSuggestions(ctx context.Context, opts *PaginationOptions) ([]GroupSuggestion, *ResponseMeta, error) {
	var resp struct {
		Data    []GroupSuggestion `json:"data"`
		Status  string            `json:"status"`
		Message string            `json:"message"`
		Meta    *ResponseMeta     `json:"meta,omitempty"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/groups/suggestions",
		Result: &resp,
	}
	if len(query) > 0 {
		req.Query = query
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// AcceptGroupSuggestion creates a server group from an ML suggestion
// Authentication: JWT Token required
// Endpoint: POST /v1/groups/suggestions/{id}/accept
// Parameters:
//   - suggestionID: Suggestion ID to accept
// Returns: Created group information
func (s *MLService) AcceptGroupSuggestion(ctx context.Context, suggestionID uint) (*GroupSuggestion, error) {
	var resp struct {
		Data    *GroupSuggestion `json:"data"`
		Status  string           `json:"status"`
		Message string           `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/groups/suggestions/%d/accept", suggestionID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// RejectGroupSuggestion rejects an ML grouping suggestion
// Authentication: JWT Token required
// Endpoint: POST /v1/groups/suggestions/{id}/reject
// Parameters:
//   - suggestionID: Suggestion ID to reject
// Returns: Success confirmation
func (s *MLService) RejectGroupSuggestion(ctx context.Context, suggestionID uint) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/groups/suggestions/%d/reject", suggestionID),
	})
	return err
}

// Model Management Methods

// ListModels retrieves available ML models with pagination
// Authentication: JWT Token required
// Endpoint: GET /v1/ml/models
// Parameters:
//   - opts: Optional pagination options
// Returns: Array of ML models with status and performance metrics
func (s *MLService) ListModels(ctx context.Context, opts *PaginationOptions) ([]MLModel, *ResponseMeta, error) {
	var resp struct {
		Data    []MLModel     `json:"data"`
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Meta    *ResponseMeta `json:"meta,omitempty"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/ml/models",
		Result: &resp,
	}
	if len(query) > 0 {
		req.Query = query
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// TrainModel triggers training for a specific ML model
// Authentication: JWT Token required
// Endpoint: POST /v1/ml/models/train
// Parameters:
//   - modelType: Type of model to train (e.g., "tag_prediction", "group_suggestion")
//   - parameters: Optional training parameters
// Returns: Training job information
func (s *MLService) TrainModel(ctx context.Context, modelType string, parameters map[string]interface{}) (*TrainingJob, error) {
	var resp struct {
		Data    *TrainingJob `json:"data"`
		Status  string       `json:"status"`
		Message string       `json:"message"`
	}

	body := map[string]interface{}{
		"model_type": modelType,
	}
	if parameters != nil {
		body["parameters"] = parameters
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/ml/models/train",
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ToggleModel enables or disables a specific ML model
// Authentication: JWT Token required
// Endpoint: PUT /v1/ml/models/{model_id}/toggle
// Parameters:
//   - modelID: Model ID to toggle
// Returns: Updated model information
func (s *MLService) ToggleModel(ctx context.Context, modelID uint) (*MLModel, error) {
	var resp struct {
		Data    *MLModel `json:"data"`
		Status  string   `json:"status"`
		Message string   `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/ml/models/%d/toggle", modelID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetModelPerformance retrieves performance metrics for a specific model
// Authentication: JWT Token required
// Endpoint: GET /v1/ml/models/{model_id}/performance
// Parameters:
//   - modelID: Model ID
// Returns: Model performance metrics including accuracy, precision, recall
func (s *MLService) GetModelPerformance(ctx context.Context, modelID uint) (*ModelPerformance, error) {
	var resp struct {
		Data    *ModelPerformance `json:"data"`
		Status  string            `json:"status"`
		Message string            `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/ml/models/%d/performance", modelID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// Training Job Methods

// TriggerModelTraining triggers batch model training for all models
// Authentication: JWT Token required
// Endpoint: POST /v1/ml/train-models
// Returns: Array of initiated training jobs
func (s *MLService) TriggerModelTraining(ctx context.Context) ([]TrainingJob, error) {
	var resp struct {
		Data    []TrainingJob `json:"data"`
		Status  string        `json:"status"`
		Message string        `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/ml/train-models",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetTrainingJobs retrieves training job history with filtering
// Authentication: JWT Token required
// Endpoint: GET /v1/ml/training-jobs
// Parameters:
//   - opts: Optional pagination and filtering options
//   - status: Optional status filter ("pending", "running", "completed", "failed")
// Returns: Array of training jobs with pagination metadata
func (s *MLService) GetTrainingJobs(ctx context.Context, opts *PaginationOptions, status string) ([]TrainingJob, *ResponseMeta, error) {
	var resp struct {
		Data    []TrainingJob `json:"data"`
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Meta    *ResponseMeta `json:"meta,omitempty"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}
	if status != "" {
		query["status"] = status
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/ml/training-jobs",
		Result: &resp,
	}
	if len(query) > 0 {
		req.Query = query
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// GetAggregatedModelPerformance retrieves aggregated performance metrics across all models
// Authentication: JWT Token required
// Endpoint: GET /v1/ml/model-performance
// Returns: Aggregated performance metrics
func (s *MLService) GetAggregatedModelPerformance(ctx context.Context) (*ModelPerformance, error) {
	var resp struct {
		Data    *ModelPerformance `json:"data"`
		Status  string            `json:"status"`
		Message string            `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/ml/model-performance",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
