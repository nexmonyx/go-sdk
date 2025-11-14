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
func (s *NotificationsService) TestChannel(ctx context.Context, orgID uint, channelID uint, testReq *ChannelTestRequest) (*ChannelTestResult, error) {
	var resp StandardResponse
	resp.Data = &ChannelTestResult{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/organizations/%d/channels/%d/test", orgID, channelID),
		Body:   testReq,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if result, ok := resp.Data.(*ChannelTestResult); ok {
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
