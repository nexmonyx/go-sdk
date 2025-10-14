package nexmonyx

import (
	"context"
	"fmt"
	"net/http"
)

// DatabaseService handles database schema management operations
type DatabaseService struct {
	client *Client
}

// SchemaResponse represents the response from schema operations
type SchemaResponse struct {
	SchemaName string `json:"schema_name"`
	Exists     bool   `json:"exists"`
	Message    string `json:"message"`
}

// CreateOrganizationSchema creates a database schema for an organization
//
// This creates a PostgreSQL schema (org_{id}) for single-tenant services like
// alert-controller, probe-controller, tag-controller, and job-controller.
//
// Parameters:
//   - ctx: Context for request cancellation and deadlines
//   - orgID: Organization ID for which to create the schema
//
// Returns:
//   - *SchemaResponse: Information about the created schema
//   - error: Any error that occurred during the operation
//
// Example:
//
//	response, err := client.Database.CreateOrganizationSchema(ctx, 123)
//	if err != nil {
//	    log.Fatalf("Failed to create schema: %v", err)
//	}
//	fmt.Printf("Schema created: %s\n", response.SchemaName)
func (s *DatabaseService) CreateOrganizationSchema(ctx context.Context, orgID uint) (*SchemaResponse, error) {
	schemaName := fmt.Sprintf("org_%d", orgID)

	requestBody := map[string]interface{}{
		"organization_id": orgID,
		"schema_name":     schemaName,
	}

	var response struct {
		Status string          `json:"status"`
		Data   *SchemaResponse `json:"data"`
	}

	resp, err := s.client.client.R().
		SetContext(ctx).
		SetBody(requestBody).
		SetResult(&response).
		Post("/v1/admin/database/schemas")

	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create schema: HTTP %d", resp.StatusCode())
	}

	return response.Data, nil
}

// DeleteOrganizationSchema drops a database schema for an organization
//
// This removes the PostgreSQL schema (org_{id}) and all its contents (CASCADE).
// Used during organization deletion or cleanup operations.
//
// Parameters:
//   - ctx: Context for request cancellation and deadlines
//   - orgID: Organization ID for which to delete the schema
//
// Returns:
//   - *SchemaResponse: Information about the deleted schema
//   - error: Any error that occurred during the operation
//
// Example:
//
//	response, err := client.Database.DeleteOrganizationSchema(ctx, 123)
//	if err != nil {
//	    log.Fatalf("Failed to delete schema: %v", err)
//	}
//	fmt.Printf("Schema deleted: %s\n", response.SchemaName)
func (s *DatabaseService) DeleteOrganizationSchema(ctx context.Context, orgID uint) (*SchemaResponse, error) {
	schemaName := fmt.Sprintf("org_%d", orgID)

	var response struct {
		Status string          `json:"status"`
		Data   *SchemaResponse `json:"data"`
	}

	resp, err := s.client.client.R().
		SetContext(ctx).
		SetResult(&response).
		Delete(fmt.Sprintf("/v1/admin/database/schemas/%s", schemaName))

	if err != nil {
		return nil, fmt.Errorf("failed to delete schema: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to delete schema: HTTP %d", resp.StatusCode())
	}

	return response.Data, nil
}

// CheckSchemaExists checks if an organization's database schema exists
//
// This queries whether the PostgreSQL schema (org_{id}) exists in the database.
//
// Parameters:
//   - ctx: Context for request cancellation and deadlines
//   - orgID: Organization ID for which to check the schema
//
// Returns:
//   - bool: True if the schema exists, false otherwise
//   - error: Any error that occurred during the operation
//
// Example:
//
//	exists, err := client.Database.CheckSchemaExists(ctx, 123)
//	if err != nil {
//	    log.Fatalf("Failed to check schema: %v", err)
//	}
//	if exists {
//	    fmt.Println("Schema exists")
//	}
func (s *DatabaseService) CheckSchemaExists(ctx context.Context, orgID uint) (bool, error) {
	schemaName := fmt.Sprintf("org_%d", orgID)

	var response struct {
		Status string          `json:"status"`
		Data   *SchemaResponse `json:"data"`
	}

	resp, err := s.client.client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(fmt.Sprintf("/v1/admin/database/schemas/%s", schemaName))

	if err != nil {
		return false, fmt.Errorf("failed to check schema: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return false, fmt.Errorf("failed to check schema: HTTP %d", resp.StatusCode())
	}

	return response.Data.Exists, nil
}
