package nexmonyx

import (
	"context"
	"fmt"
)

// Submit submits hardware inventory for a server
func (s *HardwareInventoryService) Submit(ctx context.Context, inventory *HardwareInventoryRequest) (*HardwareInventorySubmitResponse, error) {
	var resp map[string]HardwareInventorySubmitResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v2/hardware/inventory",
		Body:   inventory,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if data, ok := resp["data"]; ok {
		return &data, nil
	}
	return nil, fmt.Errorf("unexpected response format")
}

// GetInventory retrieves hardware inventory for a server
func (s *HardwareInventoryService) Get(ctx context.Context, serverUUID string) (*HardwareInventoryInfo, error) {
	var resp StandardResponse
	resp.Data = &HardwareInventoryInfo{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/hardware-inventory/%s", serverUUID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if inventory, ok := resp.Data.(*HardwareInventoryInfo); ok {
		return inventory, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListInventory retrieves hardware inventory for multiple servers
func (s *HardwareInventoryService) List(ctx context.Context, opts *ListOptions) ([]*HardwareInventoryInfo, *PaginationMeta, error) {
	var resp PaginatedResponse
	var inventories []*HardwareInventoryInfo
	resp.Data = &inventories

	req := &Request{
		Method: "GET",
		Path:   "/v1/hardware-inventory",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return inventories, resp.Meta, nil
}

// GetInventoryHistory retrieves historical hardware inventory data
func (s *HardwareInventoryService) GetHistory(ctx context.Context, serverUUID string, opts *ListOptions) ([]*HardwareInventoryInfo, *PaginationMeta, error) {
	var resp PaginatedResponse
	var inventories []*HardwareInventoryInfo
	resp.Data = &inventories

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/hardware-inventory/%s/history", serverUUID),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return inventories, resp.Meta, nil
}

// GetInventoryChanges retrieves hardware changes for a server
func (s *HardwareInventoryService) GetChanges(ctx context.Context, serverUUID string, timeRange *QueryTimeRange) ([]HardwareChange, error) {
	var resp StandardResponse
	var changes []HardwareChange
	resp.Data = &changes

	query := make(map[string]string)
	if timeRange != nil {
		start, end := timeRange.ToStrings()
		query["start"] = start
		query["end"] = end
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/hardware-inventory/%s/changes", serverUUID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return changes, nil
}

// SearchInventory searches hardware inventory
func (s *HardwareInventoryService) Search(ctx context.Context, search *HardwareSearch) ([]*HardwareInventoryInfo, *PaginationMeta, error) {
	var resp PaginatedResponse
	var inventories []*HardwareInventoryInfo
	resp.Data = &inventories

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/hardware-inventory/search",
		Body:   search,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return inventories, resp.Meta, nil
}

// ExportInventory exports hardware inventory data
func (s *HardwareInventoryService) Export(ctx context.Context, format string, serverUUIDs []string) ([]byte, error) {
	body := map[string]interface{}{
		"format":       format,
		"server_uuids": serverUUIDs,
	}

	resp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/hardware-inventory/export",
		Body:   body,
	})
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// GetHardwareInventory retrieves hardware inventory for a server within a time range
func (s *HardwareInventoryService) GetHardwareInventory(ctx context.Context, serverUUID string, timeRange *TimeRange) ([]*HardwareInventoryRecord, error) {
	var resp map[string][]*HardwareInventoryRecord

	query := make(map[string]string)
	if timeRange != nil {
		query["start"] = timeRange.Start
		query["end"] = timeRange.End
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/hardware/inventory/%s", serverUUID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if data, ok := resp["data"]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("unexpected response format")
}

// GetLatestHardwareInventory retrieves the latest hardware inventory for a server
func (s *HardwareInventoryService) GetLatestHardwareInventory(ctx context.Context, serverUUID string) (*HardwareInventoryRecord, error) {
	var resp map[string]*HardwareInventoryRecord

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/hardware/inventory/%s/latest", serverUUID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if data, ok := resp["data"]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("unexpected response format")
}

// HardwareChange represents a hardware change event
type HardwareChange struct {
	ID            uint        `json:"id"`
	ServerUUID    string      `json:"server_uuid"`
	ComponentType string      `json:"component_type"`
	ChangeType    string      `json:"change_type"` // added, removed, modified
	OldValue      interface{} `json:"old_value,omitempty"`
	NewValue      interface{} `json:"new_value,omitempty"`
	ChangedAt     *CustomTime `json:"changed_at"`
	Details       string      `json:"details,omitempty"`
}

// HardwareSearch represents hardware search parameters
type HardwareSearch struct {
	Manufacturer  string   `json:"manufacturer,omitempty"`
	Model         string   `json:"model,omitempty"`
	SerialNumber  string   `json:"serial_number,omitempty"`
	ComponentType string   `json:"component_type,omitempty"`
	ServerUUIDs   []string `json:"server_uuids,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// ListHardwareHistory retrieves hardware inventory history for a server
func (s *HardwareInventoryService) ListHardwareHistory(ctx context.Context, serverUUID string, opts *HardwareInventoryListOptions) ([]*HardwareInventoryRecord, *PaginationMeta, error) {
	var resp PaginatedResponse
	var records []*HardwareInventoryRecord
	resp.Data = &records

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/hardware/inventory/%s/history", serverUUID),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return records, resp.Meta, nil
}
