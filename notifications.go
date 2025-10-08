package nexmonyx

import (
	"context"
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
