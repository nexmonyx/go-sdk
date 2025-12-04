package nexmonyx

import (
	"context"
	"fmt"
)

// NotificationsService handles operations related to notifications
type NotificationsService struct {
	client *Client
}

// SendNotification sends a notification through configured channels
func (s *NotificationsService) SendNotification(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error) {
	var resp StandardResponse
	resp.Data = &NotificationResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/notifications/send",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if notification, ok := resp.Data.(*NotificationResponse); ok {
		return notification, nil
	}
	return nil, ErrUnexpectedResponse
}

// SendBatchNotifications sends multiple notifications in a single request
func (s *NotificationsService) SendBatchNotifications(ctx context.Context, req *BatchNotificationRequest) (*BatchNotificationResponse, error) {
	var resp StandardResponse
	resp.Data = &BatchNotificationResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/notifications/send/batch",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if batchResp, ok := resp.Data.(*BatchNotificationResponse); ok {
		return batchResp, nil
	}
	return nil, ErrUnexpectedResponse
}

// GetNotificationStatus retrieves status information for notifications
func (s *NotificationsService) GetNotificationStatus(ctx context.Context, req *NotificationStatusRequest) (*NotificationStatusResponse, error) {
	var resp StandardResponse
	resp.Data = &NotificationStatusResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/notifications/status",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if statusResp, ok := resp.Data.(*NotificationStatusResponse); ok {
		return statusResp, nil
	}
	return nil, ErrUnexpectedResponse
}

// SendQuotaAlert is a convenience method for sending quota-related notifications
func (s *NotificationsService) SendQuotaAlert(ctx context.Context, orgID uint, subject, content string, priority NotificationPriority, metadata map[string]interface{}) (*NotificationResponse, error) {
	req := &NotificationRequest{
		OrganizationID: orgID,
		Subject:        subject,
		Content:        content,
		ContentType:    "html",
		Priority:       priority,
		Metadata:       metadata,
		// Use all available channels - notification-service will filter appropriately
	}

	return s.SendNotification(ctx, req)
}

// CreateChannel creates a new notification channel for an organization
func (s *NotificationsService) CreateChannel(ctx context.Context, orgID uint, channel *NotificationChannel) (*NotificationChannel, error) {
	var resp StandardResponse
	resp.Data = &NotificationChannel{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/organizations/%d/channels", orgID),
		Body:   channel,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if ch, ok := resp.Data.(*NotificationChannel); ok {
		return ch, nil
	}
	return nil, ErrUnexpectedResponse
}

// TestChannel tests a notification channel's connectivity and configuration
func (s *NotificationsService) TestChannel(ctx context.Context, orgID uint, channelID uint, testReq *ChannelTestRequest) (*NotificationChannelTestResult, error) {
	var resp StandardResponse
	resp.Data = &NotificationChannelTestResult{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/organizations/%d/channels/%d/test", orgID, channelID),
		Body:   testReq,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if result, ok := resp.Data.(*NotificationChannelTestResult); ok {
		return result, nil
	}
	return nil, ErrUnexpectedResponse
}

// ListChannels retrieves all notification channels for an organization
func (s *NotificationsService) ListChannels(ctx context.Context, orgID uint, opts *ListOptions) ([]*NotificationChannel, *PaginationMeta, error) {
	var resp PaginatedResponse
	var channels []*NotificationChannel
	resp.Data = &channels

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/organizations/%d/channels", orgID),
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

// ============================================================================
// Notification Preferences Methods
// ============================================================================

// GetPreferences retrieves notification preferences for the organization
func (s *NotificationsService) GetPreferences(ctx context.Context) (*NotificationPreferences, error) {
	var resp StandardResponse
	resp.Data = &NotificationPreferences{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/notifications/preferences",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if prefs, ok := resp.Data.(*NotificationPreferences); ok {
		return prefs, nil
	}
	return nil, ErrUnexpectedResponse
}

// UpdatePreferences updates notification preferences for the organization
func (s *NotificationsService) UpdatePreferences(ctx context.Context, req *UpdatePreferencesRequest) (*NotificationPreferences, error) {
	var resp StandardResponse
	resp.Data = &NotificationPreferences{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   "/v1/notifications/preferences",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if prefs, ok := resp.Data.(*NotificationPreferences); ok {
		return prefs, nil
	}
	return nil, ErrUnexpectedResponse
}

// GetUserPreferences retrieves notification preferences for a specific user
func (s *NotificationsService) GetUserPreferences(ctx context.Context, userID uint) (*NotificationPreferences, error) {
	var resp StandardResponse
	resp.Data = &NotificationPreferences{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/notifications/preferences/user/%d", userID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if prefs, ok := resp.Data.(*NotificationPreferences); ok {
		return prefs, nil
	}
	return nil, ErrUnexpectedResponse
}

// UpdateUserPreferences updates notification preferences for a specific user
func (s *NotificationsService) UpdateUserPreferences(ctx context.Context, userID uint, req *UpdatePreferencesRequest) (*NotificationPreferences, error) {
	var resp StandardResponse
	resp.Data = &NotificationPreferences{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/notifications/preferences/user/%d", userID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if prefs, ok := resp.Data.(*NotificationPreferences); ok {
		return prefs, nil
	}
	return nil, ErrUnexpectedResponse
}

// DeleteUserPreferences removes user-specific preferences (falls back to org defaults)
func (s *NotificationsService) DeleteUserPreferences(ctx context.Context, userID uint) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/notifications/preferences/user/%d", userID),
	})
	return err
}

// UpdateQuietHours updates quiet hours configuration
func (s *NotificationsService) UpdateQuietHours(ctx context.Context, config *QuietHoursConfig) (*NotificationPreferences, error) {
	var resp StandardResponse
	resp.Data = &NotificationPreferences{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   "/v1/notifications/preferences/quiet-hours",
		Body:   config,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if prefs, ok := resp.Data.(*NotificationPreferences); ok {
		return prefs, nil
	}
	return nil, ErrUnexpectedResponse
}

// UpdateDigestSettings updates digest notification settings
func (s *NotificationsService) UpdateDigestSettings(ctx context.Context, config *DigestConfig) (*NotificationPreferences, error) {
	var resp StandardResponse
	resp.Data = &NotificationPreferences{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   "/v1/notifications/preferences/digest",
		Body:   config,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if prefs, ok := resp.Data.(*NotificationPreferences); ok {
		return prefs, nil
	}
	return nil, ErrUnexpectedResponse
}

// UpdateAlertFilters updates alert filter configuration
func (s *NotificationsService) UpdateAlertFilters(ctx context.Context, filters []AlertFilterConfig) (*NotificationPreferences, error) {
	var resp StandardResponse
	resp.Data = &NotificationPreferences{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   "/v1/notifications/preferences/filters",
		Body:   map[string]interface{}{"filters": filters},
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if prefs, ok := resp.Data.(*NotificationPreferences); ok {
		return prefs, nil
	}
	return nil, ErrUnexpectedResponse
}

// ============================================================================
// Notification Rate Limit Methods
// ============================================================================

// GetRateLimits retrieves rate limit configuration
func (s *NotificationsService) GetRateLimits(ctx context.Context) (*NotificationRateLimit, error) {
	var resp StandardResponse
	resp.Data = &NotificationRateLimit{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/notifications/rate-limits",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if limits, ok := resp.Data.(*NotificationRateLimit); ok {
		return limits, nil
	}
	return nil, ErrUnexpectedResponse
}

// UpdateRateLimits updates rate limit configuration
func (s *NotificationsService) UpdateRateLimits(ctx context.Context, req *UpdateRateLimitsRequest) (*NotificationRateLimit, error) {
	var resp StandardResponse
	resp.Data = &NotificationRateLimit{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   "/v1/notifications/rate-limits",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if limits, ok := resp.Data.(*NotificationRateLimit); ok {
		return limits, nil
	}
	return nil, ErrUnexpectedResponse
}

// GetRateLimitStatus retrieves current rate limit usage status
func (s *NotificationsService) GetRateLimitStatus(ctx context.Context) (*RateLimitStatus, error) {
	var resp StandardResponse
	resp.Data = &RateLimitStatus{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/notifications/rate-limits/status",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp.Data.(*RateLimitStatus); ok {
		return status, nil
	}
	return nil, ErrUnexpectedResponse
}

// ============================================================================
// Notification History Methods
// ============================================================================

// ListHistory retrieves notification history with optional filtering
func (s *NotificationsService) ListHistory(ctx context.Context, opts *ListHistoryOptions) ([]*NotificationHistory, *PaginationMeta, error) {
	var resp PaginatedResponse
	var history []*NotificationHistory
	resp.Data = &history

	req := &Request{
		Method: "GET",
		Path:   "/v1/notifications/history",
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

// GetHistory retrieves a specific notification history entry
func (s *NotificationsService) GetHistory(ctx context.Context, historyID uint) (*NotificationHistory, error) {
	var resp StandardResponse
	resp.Data = &NotificationHistory{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/notifications/history/%d", historyID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if history, ok := resp.Data.(*NotificationHistory); ok {
		return history, nil
	}
	return nil, ErrUnexpectedResponse
}

// GetHistoryStats retrieves notification statistics
func (s *NotificationsService) GetHistoryStats(ctx context.Context, startDate, endDate string) (*NotificationHistoryStats, error) {
	var resp StandardResponse
	resp.Data = &NotificationHistoryStats{}

	query := make(map[string]string)
	if startDate != "" {
		query["start_date"] = startDate
	}
	if endDate != "" {
		query["end_date"] = endDate
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/notifications/history/stats",
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if stats, ok := resp.Data.(*NotificationHistoryStats); ok {
		return stats, nil
	}
	return nil, ErrUnexpectedResponse
}

// ExportHistory exports notification history in specified format
func (s *NotificationsService) ExportHistory(ctx context.Context, opts *ExportHistoryOptions) ([]byte, error) {
	query := make(map[string]string)
	if opts != nil {
		if opts.Format != "" {
			query["format"] = opts.Format
		}
		if opts.StartDate != "" {
			query["start_date"] = opts.StartDate
		}
		if opts.EndDate != "" {
			query["end_date"] = opts.EndDate
		}
		if opts.Channel != "" {
			query["channel"] = opts.Channel
		}
		if opts.Status != "" {
			query["status"] = opts.Status
		}
	}

	resp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/notifications/history/export",
		Query:  query,
	})
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// ============================================================================
// Notification Digest Methods
// ============================================================================

// GetDigestStatus retrieves current digest queue status
func (s *NotificationsService) GetDigestStatus(ctx context.Context) (*DigestStatus, error) {
	var resp StandardResponse
	resp.Data = &DigestStatus{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/notifications/digest/status",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp.Data.(*DigestStatus); ok {
		return status, nil
	}
	return nil, ErrUnexpectedResponse
}

// AddToDigest adds a notification to the digest queue
func (s *NotificationsService) AddToDigest(ctx context.Context, req *AddToDigestRequest) (*DigestEntry, error) {
	var resp StandardResponse
	resp.Data = &AddToDigestResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/notifications/digest",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if addResp, ok := resp.Data.(*AddToDigestResponse); ok && addResp.Entry != nil {
		return addResp.Entry, nil
	}
	return nil, ErrUnexpectedResponse
}

// ProcessDigest triggers immediate digest processing
func (s *NotificationsService) ProcessDigest(ctx context.Context) (int, error) {
	var resp StandardResponse
	resp.Data = &ProcessDigestResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/notifications/digest/process",
		Result: &resp,
	})
	if err != nil {
		return 0, err
	}

	if processResp, ok := resp.Data.(*ProcessDigestResponse); ok {
		return processResp.BatchesProcessed, nil
	}
	return 0, ErrUnexpectedResponse
}

// CancelDigestEntries cancels pending digest entries
func (s *NotificationsService) CancelDigestEntries(ctx context.Context, alertID *uint) (int, error) {
	var resp StandardResponse
	resp.Data = &CancelDigestResponse{}

	body := &CancelDigestRequest{AlertID: alertID}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/notifications/digest/cancel",
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return 0, err
	}

	if cancelResp, ok := resp.Data.(*CancelDigestResponse); ok {
		return cancelResp.CancelledCount, nil
	}
	return 0, ErrUnexpectedResponse
}

// ============================================================================
// Notification Template Methods
// ============================================================================

// ListTemplates retrieves all notification templates
func (s *NotificationsService) ListTemplates(ctx context.Context, opts *ListOptions) ([]*NotificationTemplate, *PaginationMeta, error) {
	var resp PaginatedResponse
	var templates []*NotificationTemplate
	resp.Data = &templates

	req := &Request{
		Method: "GET",
		Path:   "/v1/notifications/templates",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return templates, resp.Meta, nil
}

// GetTemplate retrieves a specific notification template
func (s *NotificationsService) GetTemplate(ctx context.Context, templateID uint) (*NotificationTemplate, error) {
	var resp StandardResponse
	resp.Data = &NotificationTemplate{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/notifications/templates/%d", templateID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if template, ok := resp.Data.(*NotificationTemplate); ok {
		return template, nil
	}
	return nil, ErrUnexpectedResponse
}

// CreateTemplate creates a new notification template
func (s *NotificationsService) CreateTemplate(ctx context.Context, req *CreateNotificationTemplateRequest) (*NotificationTemplate, error) {
	var resp StandardResponse
	resp.Data = &NotificationTemplate{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/notifications/templates",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if template, ok := resp.Data.(*NotificationTemplate); ok {
		return template, nil
	}
	return nil, ErrUnexpectedResponse
}

// UpdateTemplate updates an existing notification template
func (s *NotificationsService) UpdateTemplate(ctx context.Context, templateID uint, req *UpdateNotificationTemplateRequest) (*NotificationTemplate, error) {
	var resp StandardResponse
	resp.Data = &NotificationTemplate{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/notifications/templates/%d", templateID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if template, ok := resp.Data.(*NotificationTemplate); ok {
		return template, nil
	}
	return nil, ErrUnexpectedResponse
}

// DeleteTemplate deletes a notification template
func (s *NotificationsService) DeleteTemplate(ctx context.Context, templateID uint) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/notifications/templates/%d", templateID),
	})
	return err
}

// PreviewTemplate previews a template with provided variables
func (s *NotificationsService) PreviewTemplate(ctx context.Context, templateID uint, variables map[string]interface{}) (*PreviewNotificationTemplateResponse, error) {
	var resp StandardResponse
	resp.Data = &PreviewNotificationTemplateResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/notifications/templates/%d/preview", templateID),
		Body:   &PreviewNotificationTemplateRequest{Variables: variables},
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if preview, ok := resp.Data.(*PreviewNotificationTemplateResponse); ok {
		return preview, nil
	}
	return nil, ErrUnexpectedResponse
}

// GetAvailableVariables retrieves available template variables
func (s *NotificationsService) GetAvailableVariables(ctx context.Context) (*AvailableTemplateVariables, error) {
	var resp StandardResponse
	resp.Data = &AvailableTemplateVariables{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/notifications/templates/variables",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if vars, ok := resp.Data.(*AvailableTemplateVariables); ok {
		return vars, nil
	}
	return nil, ErrUnexpectedResponse
}
