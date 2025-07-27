package nexmonyx

import (
	"context"
	"fmt"
)

// IncidentsService handles communication with the incidents endpoints
type IncidentsService struct {
	client *Client
}

// CreateIncident creates a new incident
func (s *IncidentsService) CreateIncident(ctx context.Context, req CreateIncidentRequest) (*Incident, error) {
	var result struct {
		Status  string     `json:"status"`
		Message string     `json:"message"`
		Data    *Incident  `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/incidents",
		Body:   req,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetIncident retrieves a specific incident by ID
func (s *IncidentsService) GetIncident(ctx context.Context, id uint) (*Incident, error) {
	var result struct {
		Status  string     `json:"status"`
		Message string     `json:"message"`
		Data    *Incident  `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/incidents/%d", id),
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// UpdateIncident updates an existing incident
func (s *IncidentsService) UpdateIncident(ctx context.Context, id uint, req UpdateIncidentRequest) (*Incident, error) {
	var result struct {
		Status  string     `json:"status"`
		Message string     `json:"message"`
		Data    *Incident  `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/incidents/%d", id),
		Body:   req,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// ListIncidents retrieves a paginated list of incidents
func (s *IncidentsService) ListIncidents(ctx context.Context, opts *IncidentListOptions) (*IncidentListResponse, error) {
	var result struct {
		Status  string                `json:"status"`
		Message string                `json:"message"`
		Data    *IncidentListResponse `json:"data"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Status != "" {
			query["status"] = opts.Status
		}
		if opts.Severity != "" {
			query["severity"] = opts.Severity
		}
		if opts.ServerID > 0 {
			query["server_id"] = fmt.Sprintf("%d", opts.ServerID)
		}
		if opts.ProbeID > 0 {
			query["probe_id"] = fmt.Sprintf("%d", opts.ProbeID)
		}
		if opts.Sort != "" {
			query["sort"] = opts.Sort
		}
		
		// Add pagination parameters from ListOptions
		if opts.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/incidents",
		Query:  query,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetRecentIncidents retrieves recent incidents
func (s *IncidentsService) GetRecentIncidents(ctx context.Context, limit int, severity string) ([]Incident, error) {
	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Incidents []Incident `json:"incidents"`
			Total     int64      `json:"total"`
			Limit     int        `json:"limit"`
		} `json:"data"`
	}

	query := make(map[string]string)
	if limit > 0 {
		query["limit"] = fmt.Sprintf("%d", limit)
	}
	if severity != "" {
		query["severity"] = severity
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/incidents/recent",
		Query:  query,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	return result.Data.Incidents, nil
}

// GetIncidentStats retrieves incident statistics
func (s *IncidentsService) GetIncidentStats(ctx context.Context) (*IncidentStats, error) {
	var result struct {
		Status  string         `json:"status"`
		Message string         `json:"message"`
		Data    *IncidentStats `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/incidents/stats",
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// ResolveIncident marks an incident as resolved
func (s *IncidentsService) ResolveIncident(ctx context.Context, id uint) (*Incident, error) {
	status := IncidentStatusResolved
	req := UpdateIncidentRequest{
		Status: status,
	}
	return s.UpdateIncident(ctx, id, req)
}

// AcknowledgeIncident marks an incident as acknowledged
func (s *IncidentsService) AcknowledgeIncident(ctx context.Context, id uint) (*Incident, error) {
	status := IncidentStatusAcknowledged
	req := UpdateIncidentRequest{
		Status: status,
	}
	return s.UpdateIncident(ctx, id, req)
}

// IncidentListResponse represents the response from listing incidents
type IncidentListResponse struct {
	Incidents []Incident `json:"incidents"`
	Total     int64      `json:"total"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
	Pages     int        `json:"pages"`
}

// CreateIncidentFromAlert creates an incident from an alert (for internal use by alerts-controller)
func (s *IncidentsService) CreateIncidentFromAlert(ctx context.Context, organizationID uint, alertID uint, alertName string, severity IncidentSeverity, serverID *uint, description string) (*Incident, error) {
	req := CreateIncidentRequest{
		Title:       fmt.Sprintf("Alert: %s", alertName),
		Description: description,
		Severity:    severity,
		ServerID:    serverID,
		Metadata: map[string]interface{}{
			"source":    "alert",
			"alert_id":  alertID,
		},
	}
	
	return s.CreateIncident(ctx, req)
}

// CreateIncidentFromProbe creates an incident from a probe failure (for internal use by monitoring-controller)
func (s *IncidentsService) CreateIncidentFromProbe(ctx context.Context, organizationID uint, probeID uint, probeName string, description string) (*Incident, error) {
	req := CreateIncidentRequest{
		Title:       fmt.Sprintf("Probe Failure: %s", probeName),
		Description: description,
		Severity:    IncidentSeverityCritical, // Probe failures are typically critical
		ProbeID:     &probeID,
		Metadata: map[string]interface{}{
			"source":   "probe",
			"probe_id": probeID,
		},
	}
	
	return s.CreateIncident(ctx, req)
}

// ResolveIncidentFromAlert resolves an incident that was created from an alert
func (s *IncidentsService) ResolveIncidentFromAlert(ctx context.Context, alertID uint) error {
	// List incidents related to this alert
	opts := &IncidentListOptions{
		Status: string(IncidentStatusActive),
	}
	
	incidents, err := s.ListIncidents(ctx, opts)
	if err != nil {
		return err
	}
	
	// Find and resolve incidents created by this alert
	for _, incident := range incidents.Incidents {
		if incident.Source == IncidentSourceAlert && incident.SourceID != nil && *incident.SourceID == alertID {
			_, err := s.ResolveIncident(ctx, incident.ID)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

// ResolveIncidentFromProbe resolves an incident that was created from a probe failure
func (s *IncidentsService) ResolveIncidentFromProbe(ctx context.Context, probeID uint) error {
	// List incidents related to this probe
	opts := &IncidentListOptions{
		Status:  string(IncidentStatusActive),
		ProbeID: probeID,
	}
	
	incidents, err := s.ListIncidents(ctx, opts)
	if err != nil {
		return err
	}
	
	// Resolve all active incidents for this probe
	for _, incident := range incidents.Incidents {
		if incident.Source == IncidentSourceProbe && incident.SourceID != nil && *incident.SourceID == probeID {
			_, err := s.ResolveIncident(ctx, incident.ID)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}