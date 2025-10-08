package nexmonyx

import (
	"context"
	"fmt"
)

// ClustersService handles Kubernetes cluster management and monitoring
type ClustersService struct {
	client *Client
}

// CreateCluster creates a new Kubernetes cluster for monitoring
// Authentication: JWT Token required (admin)
// Endpoint: POST /v1/admin/clusters
// Parameters:
//   - req: Cluster configuration including API server URL and credentials
// Returns: Created Cluster object
func (s *ClustersService) CreateCluster(ctx context.Context, req *ClusterCreateRequest) (*Cluster, error) {
	var resp struct {
		Data    *Cluster `json:"data"`
		Status  string   `json:"status"`
		Message string   `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/admin/clusters",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ListClusters retrieves a list of clusters with pagination
// Authentication: JWT Token required (admin)
// Endpoint: GET /v1/admin/clusters
// Parameters:
//   - opts: Optional pagination options
// Returns: Array of Cluster objects with pagination metadata
func (s *ClustersService) ListClusters(ctx context.Context, opts *PaginationOptions) ([]Cluster, *PaginationMeta, error) {
	var resp struct {
		Data []Cluster        `json:"data"`
		Meta *PaginationMeta `json:"meta"`
	}

	queryParams := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			queryParams["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/admin/clusters",
		Result: &resp,
	}
	if len(queryParams) > 0 {
		req.Query = queryParams
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// GetCluster retrieves a specific cluster by ID
// Authentication: JWT Token required (admin)
// Endpoint: GET /v1/admin/clusters/{id}
// Parameters:
//   - clusterID: Cluster ID
// Returns: Cluster object with full details including connection status
func (s *ClustersService) GetCluster(ctx context.Context, clusterID uint) (*Cluster, error) {
	var resp struct {
		Data    *Cluster `json:"data"`
		Status  string   `json:"status"`
		Message string   `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/admin/clusters/%d", clusterID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// UpdateCluster updates an existing cluster's configuration
// Authentication: JWT Token required (admin)
// Endpoint: PUT /v1/admin/clusters/{id}
// Parameters:
//   - clusterID: Cluster ID
//   - req: Updated cluster configuration
// Returns: Updated Cluster object
func (s *ClustersService) UpdateCluster(ctx context.Context, clusterID uint, req *ClusterUpdateRequest) (*Cluster, error) {
	var resp struct {
		Data    *Cluster `json:"data"`
		Status  string   `json:"status"`
		Message string   `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/admin/clusters/%d", clusterID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// DeleteCluster removes a cluster from monitoring
// Authentication: JWT Token required (admin)
// Endpoint: DELETE /v1/admin/clusters/{id}
// Parameters:
//   - clusterID: Cluster ID
// Returns: Error if deletion fails
func (s *ClustersService) DeleteCluster(ctx context.Context, clusterID uint) error {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/admin/clusters/%d", clusterID),
		Result: &resp,
	})
	if err != nil {
		return err
	}

	return nil
}
