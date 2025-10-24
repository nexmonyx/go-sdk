package nexmonyx

import (
	"context"
	"fmt"
)

// CreateAlert creates a new alert
func (s *AlertsService) Create(ctx context.Context, alert *Alert) (*Alert, error) {
	var resp StandardResponse
	resp.Data = &Alert{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/alerts/rules",
		Body:   alert,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if created, ok := resp.Data.(*Alert); ok {
		return created, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetAlert retrieves an alert by ID
func (s *AlertsService) Get(ctx context.Context, id string) (*Alert, error) {
	var resp StandardResponse
	resp.Data = &Alert{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/alerts/rules/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if alert, ok := resp.Data.(*Alert); ok {
		return alert, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListAlerts retrieves a list of alerts
func (s *AlertsService) List(ctx context.Context, opts *ListOptions) ([]*Alert, *PaginationMeta, error) {
	var resp PaginatedResponse
	var alerts []*Alert
	resp.Data = &alerts

	req := &Request{
		Method: "GET",
		Path:   "/v1/alerts/rules",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return alerts, resp.Meta, nil
}

// UpdateAlert updates an existing alert
func (s *AlertsService) Update(ctx context.Context, id string, alert *Alert) (*Alert, error) {
	var resp StandardResponse
	resp.Data = &Alert{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/alerts/rules/%s", id),
		Body:   alert,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*Alert); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteAlert deletes an alert
func (s *AlertsService) Delete(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/alerts/rules/%s", id),
		Result: &resp,
	})
	return err
}

// EnableAlert enables an alert
func (s *AlertsService) Enable(ctx context.Context, id string) (*Alert, error) {
	var resp StandardResponse
	resp.Data = &Alert{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/alerts/%s/enable", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if alert, ok := resp.Data.(*Alert); ok {
		return alert, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DisableAlert disables an alert
func (s *AlertsService) Disable(ctx context.Context, id string) (*Alert, error) {
	var resp StandardResponse
	resp.Data = &Alert{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/alerts/%s/disable", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if alert, ok := resp.Data.(*Alert); ok {
		return alert, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetAlertHistory retrieves alert history
func (s *AlertsService) GetHistory(ctx context.Context, id string, opts *ListOptions) ([]*AlertHistoryEntry, *PaginationMeta, error) {
	var resp PaginatedResponse
	var history []*AlertHistoryEntry
	resp.Data = &history

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/alerts/%s/history", id),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return history, resp.Meta, nil
}

// TestAlert tests an alert configuration
func (s *AlertsService) Test(ctx context.Context, id string) (*AlertTestResult, error) {
	var resp StandardResponse
	resp.Data = &AlertTestResult{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/alerts/%s/test", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if result, ok := resp.Data.(*AlertTestResult); ok {
		return result, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// AcknowledgeAlert acknowledges an alert
func (s *AlertsService) Acknowledge(ctx context.Context, id string, message string) error {
	var resp StandardResponse

	body := map[string]interface{}{
		"message": message,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/alerts/%s/acknowledge", id),
		Body:   body,
		Result: &resp,
	})
	return err
}

// AlertHistoryEntry represents an alert history entry
type AlertHistoryEntry struct {
	ID          uint                   `json:"id"`
	AlertID     uint                   `json:"alert_id"`
	TriggeredAt *CustomTime            `json:"triggered_at"`
	ResolvedAt  *CustomTime            `json:"resolved_at,omitempty"`
	Status      string                 `json:"status"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// AlertTestResult represents the result of an alert test
type AlertTestResult struct {
	Success   bool                   `json:"success"`
	Triggered bool                   `json:"triggered"`
	Message   string                 `json:"message"`
	Value     float64                `json:"value,omitempty"`
	Threshold float64                `json:"threshold,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Errors    []string               `json:"errors,omitempty"`
}

// ListChannels retrieves all notification channels for an organization
func (s *AlertsService) ListChannels(ctx context.Context, opts *ListOptions) ([]*AlertChannel, *PaginationMeta, error) {
	var resp PaginatedResponse
	var channels []*AlertChannel
	resp.Data = &channels

	req := &Request{
		Method: "GET",
		Path:   "/v1/alerts/channels",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return channels, resp.Meta, nil
}

// CreateChannel creates a new alert notification channel
func (s *AlertsService) CreateChannel(ctx context.Context, channel *AlertChannel) (*AlertChannel, error) {
	var resp StandardResponse
	resp.Data = &AlertChannel{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/alerts/channels",
		Body:   channel,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if created, ok := resp.Data.(*AlertChannel); ok {
		return created, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetChannel retrieves an alert notification channel by ID
func (s *AlertsService) GetChannel(ctx context.Context, id string) (*AlertChannel, error) {
	var resp StandardResponse
	resp.Data = &AlertChannel{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/alerts/channels/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if channel, ok := resp.Data.(*AlertChannel); ok {
		return channel, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateChannel updates an existing alert notification channel
func (s *AlertsService) UpdateChannel(ctx context.Context, id string, channel *AlertChannel) (*AlertChannel, error) {
	var resp StandardResponse
	resp.Data = &AlertChannel{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/alerts/channels/%s", id),
		Body:   channel,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*AlertChannel); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteChannel deletes an alert notification channel
func (s *AlertsService) DeleteChannel(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/alerts/channels/%s", id),
		Result: &resp,
	})
	return err
}

// TestChannel tests an alert notification channel configuration
func (s *AlertsService) TestChannel(ctx context.Context, id string) (*ChannelTestResult, error) {
	var resp StandardResponse
	resp.Data = &ChannelTestResult{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/alerts/channels/%s/test", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if result, ok := resp.Data.(*ChannelTestResult); ok {
		return result, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ChannelTestResult represents the result of testing an alert notification channel
type ChannelTestResult struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Errors  []string               `json:"errors,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Alert Instances Management

// ListInstances retrieves active alert instances
func (s *AlertsService) ListInstances(ctx context.Context, opts *ListOptions) ([]*AlertInstance, *PaginationMeta, error) {
	var resp PaginatedResponse
	var instances []*AlertInstance
	resp.Data = &instances

	req := &Request{
		Method: "GET",
		Path:   "/v1/alerts/active",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return instances, resp.Meta, nil
}

// GetInstance retrieves an alert instance by ID
func (s *AlertsService) GetInstance(ctx context.Context, id string) (*AlertInstance, error) {
	var resp StandardResponse
	resp.Data = &AlertInstance{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/alerts/instances/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if instance, ok := resp.Data.(*AlertInstance); ok {
		return instance, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ResolveInstance resolves an alert instance
func (s *AlertsService) ResolveInstance(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/alerts/instances/%s/resolve", id),
		Result: &resp,
	})
	return err
}

// CreateInstance creates a new alert instance
func (s *AlertsService) CreateInstance(ctx context.Context, req *CreateAlertInstanceRequest) (*AlertInstance, error) {
	var resp StandardResponse
	resp.Data = &AlertInstance{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/alerts/instances",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if instance, ok := resp.Data.(*AlertInstance); ok {
		return instance, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateInstance updates an existing alert instance
func (s *AlertsService) UpdateInstance(ctx context.Context, id string, req *UpdateAlertInstanceRequest) (*AlertInstance, error) {
	var resp StandardResponse
	resp.Data = &AlertInstance{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/alerts/instances/%s", id),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if instance, ok := resp.Data.(*AlertInstance); ok {
		return instance, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// SilenceInstance silences an alert instance for a specified duration
func (s *AlertsService) SilenceInstance(ctx context.Context, id string, duration int) error {
	var resp StandardResponse

	body := map[string]interface{}{
		"duration": duration,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/alerts/instances/%s/silence", id),
		Body:   body,
		Result: &resp,
	})
	return err
}

// GetInstanceHistory retrieves history for alert instances
func (s *AlertsService) GetInstanceHistory(ctx context.Context, opts *ListOptions) ([]*AlertInstance, *PaginationMeta, error) {
	var resp PaginatedResponse
	var instances []*AlertInstance
	resp.Data = &instances

	req := &Request{
		Method: "GET",
		Path:   "/v1/alerts/history",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return instances, resp.Meta, nil
}

// GetInstanceMetrics retrieves alert metrics
func (s *AlertsService) GetInstanceMetrics(ctx context.Context) (map[string]interface{}, error) {
	var resp StandardResponse
	var metrics map[string]interface{}
	resp.Data = &metrics

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/alerts/metrics",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// Alert Channels V2 Management

// ListChannelsV2 retrieves all notification channels (V2 API)
func (s *AlertsService) ListChannelsV2(ctx context.Context, opts *ListOptions) ([]*AlertChannel, *PaginationMeta, error) {
	var resp PaginatedResponse
	var channels []*AlertChannel
	resp.Data = &channels

	req := &Request{
		Method: "GET",
		Path:   "/v2/alerts/channels",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return channels, resp.Meta, nil
}

// CreateChannelV2 creates a new alert notification channel (V2 API)
func (s *AlertsService) CreateChannelV2(ctx context.Context, channel *AlertChannel) (*AlertChannel, error) {
	var resp StandardResponse
	resp.Data = &AlertChannel{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v2/alerts/channels",
		Body:   channel,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if created, ok := resp.Data.(*AlertChannel); ok {
		return created, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetChannelV2 retrieves an alert notification channel by ID (V2 API)
func (s *AlertsService) GetChannelV2(ctx context.Context, id string) (*AlertChannel, error) {
	var resp StandardResponse
	resp.Data = &AlertChannel{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/alerts/channels/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if channel, ok := resp.Data.(*AlertChannel); ok {
		return channel, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateChannelV2 updates an existing alert notification channel (V2 API)
func (s *AlertsService) UpdateChannelV2(ctx context.Context, id string, channel *AlertChannel) (*AlertChannel, error) {
	var resp StandardResponse
	resp.Data = &AlertChannel{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v2/alerts/channels/%s", id),
		Body:   channel,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*AlertChannel); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteChannelV2 deletes an alert notification channel (V2 API)
func (s *AlertsService) DeleteChannelV2(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v2/alerts/channels/%s", id),
		Result: &resp,
	})
	return err
}

// Alert Contacts Management

// ListContacts retrieves all alert contacts
func (s *AlertsService) ListContacts(ctx context.Context, opts *ListOptions) ([]*Contact, *PaginationMeta, error) {
	var resp PaginatedResponse
	var contacts []*Contact
	resp.Data = &contacts

	req := &Request{
		Method: "GET",
		Path:   "/v1/alerts/contacts",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return contacts, resp.Meta, nil
}

// CreateContact creates a new alert contact
func (s *AlertsService) CreateContact(ctx context.Context, contact *Contact) (*Contact, error) {
	var resp StandardResponse
	resp.Data = &Contact{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/alerts/contacts",
		Body:   contact,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if created, ok := resp.Data.(*Contact); ok {
		return created, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetContact retrieves an alert contact by ID
func (s *AlertsService) GetContact(ctx context.Context, id string) (*Contact, error) {
	var resp StandardResponse
	resp.Data = &Contact{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/alerts/contacts/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if contact, ok := resp.Data.(*Contact); ok {
		return contact, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateContact updates an existing alert contact
func (s *AlertsService) UpdateContact(ctx context.Context, id string, contact *Contact) (*Contact, error) {
	var resp StandardResponse
	resp.Data = &Contact{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/alerts/contacts/%s", id),
		Body:   contact,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*Contact); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteContact deletes an alert contact
func (s *AlertsService) DeleteContact(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/alerts/contacts/%s", id),
		Result: &resp,
	})
	return err
}

// Alert Silences Management

// ListSilences retrieves all alert silences
func (s *AlertsService) ListSilences(ctx context.Context, opts *ListOptions) ([]*AlertSilence, *PaginationMeta, error) {
	var resp PaginatedResponse
	var silences []*AlertSilence
	resp.Data = &silences

	req := &Request{
		Method: "GET",
		Path:   "/v1/alerts/silences",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return silences, resp.Meta, nil
}

// CreateSilence creates a new alert silence
func (s *AlertsService) CreateSilence(ctx context.Context, silence *AlertSilence) (*AlertSilence, error) {
	var resp StandardResponse
	resp.Data = &AlertSilence{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/alerts/silences",
		Body:   silence,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if created, ok := resp.Data.(*AlertSilence); ok {
		return created, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetSilence retrieves an alert silence by ID
func (s *AlertsService) GetSilence(ctx context.Context, id string) (*AlertSilence, error) {
	var resp StandardResponse
	resp.Data = &AlertSilence{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/alerts/silences/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if silence, ok := resp.Data.(*AlertSilence); ok {
		return silence, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateSilence updates an existing alert silence
func (s *AlertsService) UpdateSilence(ctx context.Context, id string, silence *AlertSilence) (*AlertSilence, error) {
	var resp StandardResponse
	resp.Data = &AlertSilence{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/alerts/silences/%s", id),
		Body:   silence,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*AlertSilence); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteSilence deletes an alert silence
func (s *AlertsService) DeleteSilence(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/alerts/silences/%s", id),
		Result: &resp,
	})
	return err
}

// Controller Alert Endpoints (for alert-controller service)

// ListRulesForController retrieves alert rules for controller processing
func (s *AlertsService) ListRulesForController(ctx context.Context) ([]*Alert, error) {
	var resp StandardResponse
	var rules []*Alert
	resp.Data = &rules

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/controllers/alerts/rules",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return rules, nil
}

// ListActiveForController retrieves active alerts for controller processing
func (s *AlertsService) ListActiveForController(ctx context.Context) ([]*AlertInstance, error) {
	var resp StandardResponse
	var instances []*AlertInstance
	resp.Data = &instances

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/controllers/alerts/active",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return instances, nil
}

// GetControllerStatus retrieves alert controller status
func (s *AlertsService) GetControllerStatus(ctx context.Context) (map[string]interface{}, error) {
	var resp StandardResponse
	var status map[string]interface{}
	resp.Data = &status

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/controllers/alerts/status",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return status, nil
}
