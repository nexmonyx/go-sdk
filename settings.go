package nexmonyx

import (
	"context"
	"fmt"
)

// GetSettings retrieves settings for an organization
func (s *SettingsService) Get(ctx context.Context, organizationID string) (*Settings, error) {
	var resp StandardResponse
	resp.Data = &Settings{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/organizations/%s/settings", organizationID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if settings, ok := resp.Data.(*Settings); ok {
		return settings, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateSettings updates settings for an organization
func (s *SettingsService) Update(ctx context.Context, organizationID string, settings *Settings) (*Settings, error) {
	var resp StandardResponse
	resp.Data = &Settings{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/api/v1/organizations/%s/settings", organizationID),
		Body:   settings,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*Settings); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetNotificationSettings retrieves notification settings
func (s *SettingsService) GetNotificationSettings(ctx context.Context, organizationID string) (*NotificationSettings, error) {
	var resp StandardResponse
	resp.Data = &NotificationSettings{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/organizations/%s/settings/notifications", organizationID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if settings, ok := resp.Data.(*NotificationSettings); ok {
		return settings, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateNotificationSettings updates notification settings
func (s *SettingsService) UpdateNotificationSettings(ctx context.Context, organizationID string, settings *NotificationSettings) (*NotificationSettings, error) {
	var resp StandardResponse
	resp.Data = &NotificationSettings{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/api/v1/organizations/%s/settings/notifications", organizationID),
		Body:   settings,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*NotificationSettings); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Settings represents organization settings
type Settings struct {
	OrganizationID          uint                   `json:"organization_id"`
	GeneralSettings         *GeneralSettings       `json:"general,omitempty"`
	SecuritySettings        *SecuritySettings      `json:"security,omitempty"`
	NotificationSettings    *NotificationSettings  `json:"notifications,omitempty"`
	MonitoringSettings      *MonitoringSettings    `json:"monitoring,omitempty"`
	IntegrationSettings     *IntegrationSettings   `json:"integrations,omitempty"`
	CustomSettings          map[string]interface{} `json:"custom,omitempty"`
}

// GeneralSettings represents general organization settings
type GeneralSettings struct {
	TimeZone              string   `json:"timezone"`
	DateFormat            string   `json:"date_format"`
	TimeFormat            string   `json:"time_format"`
	Language              string   `json:"language"`
	DefaultDashboard      string   `json:"default_dashboard"`
	AllowedIPAddresses    []string `json:"allowed_ip_addresses,omitempty"`
	SessionTimeout        int      `json:"session_timeout"` // minutes
	PasswordPolicy        *PasswordPolicy `json:"password_policy,omitempty"`
}

// SecuritySettings represents security-related settings
type SecuritySettings struct {
	Require2FA            bool     `json:"require_2fa"`
	AllowAPIKeys          bool     `json:"allow_api_keys"`
	APIKeyExpiration      int      `json:"api_key_expiration"` // days
	AllowedDomains        []string `json:"allowed_domains,omitempty"`
	SSOEnabled            bool     `json:"sso_enabled"`
	SSOProvider           string   `json:"sso_provider,omitempty"`
	SSOConfig             map[string]interface{} `json:"sso_config,omitempty"`
	AuditLogRetention     int      `json:"audit_log_retention"` // days
}

// NotificationSettings represents notification settings
type NotificationSettings struct {
	EmailEnabled          bool                   `json:"email_enabled"`
	EmailRecipients       []string               `json:"email_recipients"`
	SlackEnabled          bool                   `json:"slack_enabled"`
	SlackWebhook          string                 `json:"slack_webhook,omitempty"`
	SlackChannels         map[string]string      `json:"slack_channels,omitempty"`
	PagerDutyEnabled      bool                   `json:"pagerduty_enabled"`
	PagerDutyIntegration  *PagerDutyIntegration  `json:"pagerduty_integration,omitempty"`
	WebhooksEnabled       bool                   `json:"webhooks_enabled"`
	Webhooks              []WebhookConfig        `json:"webhooks,omitempty"`
	NotificationRules     []NotificationRule     `json:"notification_rules,omitempty"`
}

// MonitoringSettings represents monitoring-related settings
type MonitoringSettings struct {
	MetricsRetention      int      `json:"metrics_retention"` // days
	DefaultInterval       int      `json:"default_interval"` // seconds
	EnableAutoDiscovery   bool     `json:"enable_auto_discovery"`
	AutoDiscoveryFilters  []string `json:"auto_discovery_filters,omitempty"`
	DefaultAlertThresholds map[string]float64 `json:"default_alert_thresholds,omitempty"`
	MaintenanceWindows    []MaintenanceWindow `json:"maintenance_windows,omitempty"`
}

// IntegrationSettings represents third-party integration settings
type IntegrationSettings struct {
	AWSEnabled            bool                   `json:"aws_enabled"`
	AWSConfig             map[string]interface{} `json:"aws_config,omitempty"`
	AzureEnabled          bool                   `json:"azure_enabled"`
	AzureConfig           map[string]interface{} `json:"azure_config,omitempty"`
	GCPEnabled            bool                   `json:"gcp_enabled"`
	GCPConfig             map[string]interface{} `json:"gcp_config,omitempty"`
	DatadogEnabled        bool                   `json:"datadog_enabled"`
	DatadogConfig         map[string]interface{} `json:"datadog_config,omitempty"`
	PrometheusEnabled     bool                   `json:"prometheus_enabled"`
	PrometheusConfig      map[string]interface{} `json:"prometheus_config,omitempty"`
}

// PasswordPolicy represents password policy settings
type PasswordPolicy struct {
	MinLength             int  `json:"min_length"`
	RequireUppercase      bool `json:"require_uppercase"`
	RequireLowercase      bool `json:"require_lowercase"`
	RequireNumbers        bool `json:"require_numbers"`
	RequireSpecialChars   bool `json:"require_special_chars"`
	ExpirationDays        int  `json:"expiration_days"`
	HistoryCount          int  `json:"history_count"` // Prevent reuse of last N passwords
}

// PagerDutyIntegration represents PagerDuty integration settings
type PagerDutyIntegration struct {
	APIKey          string `json:"api_key"`
	ServiceID       string `json:"service_id"`
	EscalationPolicy string `json:"escalation_policy"`
}

// WebhookConfig represents a webhook configuration
type WebhookConfig struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	URL              string   `json:"url"`
	Secret           string   `json:"secret,omitempty"`
	Events           []string `json:"events"`
	Enabled          bool     `json:"enabled"`
	Headers          map[string]string `json:"headers,omitempty"`
}

// NotificationRule represents a notification rule
type NotificationRule struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	EventType        string                 `json:"event_type"`
	Conditions       map[string]interface{} `json:"conditions"`
	Channels         []string               `json:"channels"`
	Enabled          bool                   `json:"enabled"`
	SeverityFilter   []string               `json:"severity_filter,omitempty"`
}

// MaintenanceWindow represents a maintenance window
type MaintenanceWindow struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	StartTime        *CustomTime `json:"start_time"`
	EndTime          *CustomTime `json:"end_time"`
	Recurrence       string      `json:"recurrence,omitempty"` // daily, weekly, monthly
	AffectedServices []string    `json:"affected_services,omitempty"`
	Enabled          bool        `json:"enabled"`
}