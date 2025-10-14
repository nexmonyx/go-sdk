package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// Version is the current version of the SDK
	Version = "2.3.1"

	defaultTimeout = 30 * time.Second
	defaultBaseURL = "https://api.nexmonyx.com"
	userAgent      = "nexmonyx-go-sdk/" + Version
)

// Client is the main entry point for the Nexmonyx SDK
type Client struct {
	// HTTP client
	client *resty.Client

	// Configuration
	config *Config

	// Service clients
	Organizations         *OrganizationsService
	Servers               *ServersService
	Users                 *UsersService
	Metrics               *MetricsService
	Monitoring            *MonitoringService
	Billing               *BillingService
	BillingUsage          *BillingUsageService
	Settings              *SettingsService
	Alerts                *AlertsService
	ProbeAlerts           *ProbeAlertsService
	Admin                 *AdminService
	StatusPages           *StatusPagesService
	Providers             *ProvidersService
	Jobs                  *JobsService
	BackgroundJobs        *BackgroundJobsService
	APIKeys               *APIKeysService
	System                *SystemService
	Terms                 *TermsService
	EmailQueue            *EmailQueueService
	Public                *PublicService
	Distros               *DistrosService
	AgentDownload         *AgentDownloadService
	Controllers           *ControllersService
	HardwareInventory     *HardwareInventoryService
	IPMI                  *IPMIService
	Systemd               *SystemdService
	NetworkHardware       *NetworkHardwareService
	MonitoringDeployments *MonitoringDeploymentsService
	NamespaceDeployments  *NamespaceDeploymentsService
	MonitoringAgentKeys   *MonitoringAgentKeysService
	RemoteClusters        *RemoteClustersService
	Health                *HealthService
	ServiceMonitoring     *ServiceMonitoringService
	Probes                *ProbesService
	Incidents             *IncidentsService
	AgentVersions         *AgentVersionsService
	DiskIO                *DiskIOService
	SmartHealth           *SmartHealthService
	Filesystem            *FilesystemService
	Tags                  *TagsService
	Analytics             *AnalyticsService
	ML                    *MLService
	VMs                   *VMsService
	Reporting             *ReportingService
	ServerGroups          *ServerGroupsService
	Search                *SearchService
	Audit                 *AuditService
	Tasks                 *TasksService
	Clusters              *ClustersService
	Packages              *PackagesService
	Notifications         *NotificationsService
	WebSocket             *WebSocketServiceImpl
	ProbeController       *ProbeControllerService
	Database              *DatabaseService
}

// Config holds the configuration for the client
type Config struct {
	// Base URL of the Nexmonyx API
	BaseURL string

	// Authentication configuration
	Auth AuthConfig

	// HTTP client configuration
	HTTPClient *http.Client

	// Request timeout
	Timeout time.Duration

	// Custom headers to add to all requests
	Headers map[string]string

	// Debug mode enables request/response logging
	Debug bool

	// Retry configuration
	RetryCount    int
	RetryWaitTime time.Duration
	RetryMaxWait  time.Duration
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// JWT Token (for user authentication via Auth0)
	Token string

	// Unified API Key authentication (preferred method)
	UnifiedAPIKey string // Complete token for bearer auth or key for key/secret auth
	APIKeySecret  string // Secret part when using key/secret authentication

	// Legacy API Key authentication (deprecated, use UnifiedAPIKey instead)
	APIKey    string
	APISecret string

	// Server authentication (for agents) - will be migrated to unified keys
	ServerUUID   string
	ServerSecret string

	// Monitoring key authentication (deprecated, use UnifiedAPIKey instead)
	MonitoringKey string

	// Registration key authentication (for server registration)
	RegistrationKey string
}

// NewClient creates a new Nexmonyx API client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = &Config{}
	}

	// Set defaults
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}
	if config.RetryWaitTime == 0 {
		config.RetryWaitTime = 1 * time.Second
	}
	if config.RetryMaxWait == 0 {
		config.RetryMaxWait = 30 * time.Second
	}

	// Create HTTP client if not provided
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	// Create resty client
	restyClient := resty.NewWithClient(httpClient)
	restyClient.SetBaseURL(config.BaseURL)
	restyClient.SetTimeout(config.Timeout)
	restyClient.SetHeader("User-Agent", userAgent)
	restyClient.SetHeader("Content-Type", "application/json")
	restyClient.SetHeader("Accept", "application/json")

	// Set authentication headers (priority order: JWT Token, Unified API Key, Legacy methods)
	if config.Auth.Token != "" {
		// JWT Token authentication (highest priority)
		restyClient.SetAuthToken(config.Auth.Token)
	} else if config.Auth.UnifiedAPIKey != "" {
		// Unified API Key authentication (preferred method)
		if config.Auth.APIKeySecret != "" {
			// Key/Secret authentication
			restyClient.SetHeader("Access-Key", config.Auth.UnifiedAPIKey)
			restyClient.SetHeader("Access-Secret", config.Auth.APIKeySecret)
		} else {
			// Bearer token authentication (for monitoring agents, etc.)
			restyClient.SetAuthToken(config.Auth.UnifiedAPIKey)
		}
	} else if config.Auth.RegistrationKey != "" {
		// Registration key authentication (for server registration)
		restyClient.SetHeader("X-Registration-Key", config.Auth.RegistrationKey)
	} else if config.Auth.APIKey != "" && config.Auth.APISecret != "" {
		// Legacy API Key authentication (deprecated)
		restyClient.SetHeader("Access-Key", config.Auth.APIKey)
		restyClient.SetHeader("Access-Secret", config.Auth.APISecret)
	} else if config.Auth.ServerUUID != "" && config.Auth.ServerSecret != "" {
		// Server authentication (for agents) - will be migrated to unified keys
		// Note: Server authentication uses X- prefix headers while API Key/Secret uses Access- prefix
		// This inconsistency should be addressed in future API standardization
		restyClient.SetHeader("X-Server-UUID", config.Auth.ServerUUID)
		restyClient.SetHeader("X-Server-Secret", config.Auth.ServerSecret)
	} else if config.Auth.MonitoringKey != "" {
		// Legacy monitoring key authentication (deprecated)
		restyClient.SetAuthToken(config.Auth.MonitoringKey)
	}

	// Set custom headers
	for k, v := range config.Headers {
		restyClient.SetHeader(k, v)
	}

	// Configure retry
	restyClient.SetRetryCount(config.RetryCount)
	restyClient.SetRetryWaitTime(config.RetryWaitTime)
	restyClient.SetRetryMaxWaitTime(config.RetryMaxWait)
	restyClient.AddRetryCondition(func(r *resty.Response, err error) bool {
		return err != nil || r.StatusCode() >= 500 || r.StatusCode() == 429
	})

	// Set debug mode
	restyClient.SetDebug(config.Debug)

	// Create client
	client := &Client{
		client: restyClient,
		config: config,
	}

	// Initialize service clients
	client.Organizations = &OrganizationsService{client: client}
	client.Servers = &ServersService{client: client}
	client.Users = &UsersService{client: client}
	client.Metrics = &MetricsService{client: client}
	client.Monitoring = &MonitoringService{client: client}
	client.Billing = &BillingService{client: client}
	client.BillingUsage = &BillingUsageService{client: client}
	client.Settings = &SettingsService{client: client}
	client.Alerts = &AlertsService{client: client}
	client.ProbeAlerts = &ProbeAlertsService{client: client}
	client.Admin = &AdminService{client: client}
	client.StatusPages = &StatusPagesService{client: client}
	client.Providers = &ProvidersService{client: client}
	client.Jobs = &JobsService{client: client}
	client.BackgroundJobs = &BackgroundJobsService{client: client}
	client.APIKeys = &APIKeysService{client: client}
	client.System = &SystemService{client: client}
	client.Terms = &TermsService{client: client}
	client.EmailQueue = &EmailQueueService{client: client}
	client.Public = &PublicService{client: client}
	client.Distros = &DistrosService{client: client}
	client.AgentDownload = &AgentDownloadService{client: client}
	client.Controllers = &ControllersService{client: client}
	client.HardwareInventory = &HardwareInventoryService{client: client}
	client.IPMI = &IPMIService{client: client}
	client.Systemd = &SystemdService{client: client}
	client.NetworkHardware = &NetworkHardwareService{client: client}
	client.MonitoringDeployments = &MonitoringDeploymentsService{client: client}
	client.NamespaceDeployments = &NamespaceDeploymentsService{client: client}
	client.MonitoringAgentKeys = &MonitoringAgentKeysService{client: client}
	client.RemoteClusters = &RemoteClustersService{client: client}
	client.Health = &HealthService{client: client}
	client.ServiceMonitoring = &ServiceMonitoringService{client: client}
	client.Probes = &ProbesService{client: client}
	client.Incidents = &IncidentsService{client: client}
	client.AgentVersions = &AgentVersionsService{client: client}
	client.DiskIO = &DiskIOService{client: client}
	client.SmartHealth = &SmartHealthService{client: client}
	client.Filesystem = &FilesystemService{client: client}
	client.Tags = &TagsService{client: client}
	client.Analytics = &AnalyticsService{client: client}
	client.ML = &MLService{client: client}
	client.VMs = &VMsService{client: client}
	client.Reporting = &ReportingService{client: client}
	client.ServerGroups = &ServerGroupsService{client: client}
	client.Search = &SearchService{client: client}
	client.Audit = &AuditService{client: client}
	client.Tasks = &TasksService{client: client}
	client.Clusters = &ClustersService{client: client}
	client.Packages = &PackagesService{client: client}
	client.Notifications = &NotificationsService{client: client}
	client.ProbeController = &ProbeControllerService{client: client}
	client.Database = &DatabaseService{client: client}

	// Note: WebSocket service requires separate initialization via NewWebSocketService()
	// to ensure proper server credentials validation and connection management
	client.WebSocket = nil

	return client, nil
}

// WithToken creates a new client with the specified authentication token
func (c *Client) WithToken(token string) *Client {
	newConfig := *c.config
	newConfig.Auth.Token = token
	newConfig.Auth.UnifiedAPIKey = ""
	newConfig.Auth.APIKeySecret = ""
	newConfig.Auth.APIKey = ""
	newConfig.Auth.APISecret = ""
	newConfig.Auth.ServerUUID = ""
	newConfig.Auth.ServerSecret = ""
	newConfig.Auth.MonitoringKey = ""
	newConfig.Auth.RegistrationKey = ""

	newClient, _ := NewClient(&newConfig)
	return newClient
}

// WithUnifiedAPIKey creates a new client with unified API key authentication (bearer token)
func (c *Client) WithUnifiedAPIKey(key string) *Client {
	newConfig := *c.config
	newConfig.Auth.Token = ""
	newConfig.Auth.UnifiedAPIKey = key
	newConfig.Auth.APIKeySecret = ""
	newConfig.Auth.APIKey = ""
	newConfig.Auth.APISecret = ""
	newConfig.Auth.ServerUUID = ""
	newConfig.Auth.ServerSecret = ""
	newConfig.Auth.MonitoringKey = ""
	newConfig.Auth.RegistrationKey = ""

	newClient, _ := NewClient(&newConfig)
	return newClient
}

// WithUnifiedAPIKeyAndSecret creates a new client with unified API key authentication (key/secret)
func (c *Client) WithUnifiedAPIKeyAndSecret(key, secret string) *Client {
	newConfig := *c.config
	newConfig.Auth.Token = ""
	newConfig.Auth.UnifiedAPIKey = key
	newConfig.Auth.APIKeySecret = secret
	newConfig.Auth.APIKey = ""
	newConfig.Auth.APISecret = ""
	newConfig.Auth.ServerUUID = ""
	newConfig.Auth.ServerSecret = ""
	newConfig.Auth.MonitoringKey = ""
	newConfig.Auth.RegistrationKey = ""

	newClient, _ := NewClient(&newConfig)
	return newClient
}

// WithRegistrationKey creates a new client with registration key authentication
func (c *Client) WithRegistrationKey(key string) *Client {
	newConfig := *c.config
	newConfig.Auth.Token = ""
	newConfig.Auth.UnifiedAPIKey = ""
	newConfig.Auth.APIKeySecret = ""
	newConfig.Auth.APIKey = ""
	newConfig.Auth.APISecret = ""
	newConfig.Auth.ServerUUID = ""
	newConfig.Auth.ServerSecret = ""
	newConfig.Auth.MonitoringKey = ""
	newConfig.Auth.RegistrationKey = key

	newClient, _ := NewClient(&newConfig)
	return newClient
}

// WithAPIKey creates a new client with API key authentication (legacy method)
func (c *Client) WithAPIKey(key, secret string) *Client {
	newConfig := *c.config
	newConfig.Auth.Token = ""
	newConfig.Auth.UnifiedAPIKey = ""
	newConfig.Auth.APIKeySecret = ""
	newConfig.Auth.APIKey = key
	newConfig.Auth.APISecret = secret
	newConfig.Auth.ServerUUID = ""
	newConfig.Auth.ServerSecret = ""
	newConfig.Auth.MonitoringKey = ""
	newConfig.Auth.RegistrationKey = ""

	newClient, _ := NewClient(&newConfig)
	return newClient
}

// WithServerCredentials creates a new client with server authentication
func (c *Client) WithServerCredentials(uuid, secret string) *Client {
	newConfig := *c.config
	newConfig.Auth.Token = ""
	newConfig.Auth.UnifiedAPIKey = ""
	newConfig.Auth.APIKeySecret = ""
	newConfig.Auth.APIKey = ""
	newConfig.Auth.APISecret = ""
	newConfig.Auth.ServerUUID = uuid
	newConfig.Auth.ServerSecret = secret
	newConfig.Auth.MonitoringKey = ""
	newConfig.Auth.RegistrationKey = ""

	newClient, _ := NewClient(&newConfig)
	return newClient
}

// WithMonitoringKey creates a new client with monitoring key authentication (legacy method)
func (c *Client) WithMonitoringKey(key string) *Client {
	newConfig := *c.config
	newConfig.Auth.Token = ""
	newConfig.Auth.UnifiedAPIKey = ""
	newConfig.Auth.APIKeySecret = ""
	newConfig.Auth.APIKey = ""
	newConfig.Auth.APISecret = ""
	newConfig.Auth.ServerUUID = ""
	newConfig.Auth.ServerSecret = ""
	newConfig.Auth.MonitoringKey = key
	newConfig.Auth.RegistrationKey = ""

	newClient, _ := NewClient(&newConfig)
	return newClient
}

// NewMonitoringAgentClient creates a new client specifically for monitoring agents
func NewMonitoringAgentClient(config *Config) (*Client, error) {
	if config == nil {
		config = &Config{}
	}

	// Validate that monitoring key is provided
	if config.Auth.MonitoringKey == "" {
		return nil, fmt.Errorf("monitoring key is required for monitoring agent client")
	}

	// Ensure no other auth methods are set
	config.Auth.Token = ""
	config.Auth.APIKey = ""
	config.Auth.APISecret = ""
	config.Auth.ServerUUID = ""
	config.Auth.ServerSecret = ""

	return NewClient(config)
}

// Do performs a raw HTTP request
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	// Build resty request
	r := c.client.R().SetContext(ctx)

	// Set body if provided
	if req.Body != nil {
		r.SetBody(req.Body)
	}

	// Set query parameters
	if req.Query != nil {
		r.SetQueryParams(req.Query)
	}

	// Set additional headers
	for k, v := range req.Headers {
		r.SetHeader(k, v)
	}

	// Set result and error objects
	if req.Result != nil {
		r.SetResult(req.Result)
	}
	if req.Error != nil {
		r.SetError(req.Error)
	}

	// Debug logging for authentication headers
	if c.config.Debug {
		fmt.Printf("[DEBUG] Request: %s %s\n", req.Method, req.Path)
		fmt.Printf("[DEBUG] Headers being sent:\n")
		for k, v := range r.Header {
			// Mask sensitive headers for security
			if k == "Server-Secret" || k == "X-Server-Secret" || k == "X-Api-Secret" || k == "Authorization" {
				fmt.Printf("[DEBUG]   %s: [REDACTED]\n", k)
			} else {
				fmt.Printf("[DEBUG]   %s: %v\n", k, v)
			}
		}
		if req.Body != nil {
			fmt.Printf("[DEBUG] Request has body (type: %T)\n", req.Body)
		}
	}

	// Execute request
	resp, err := r.Execute(req.Method, req.Path)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Handle errors
	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &Response{
		StatusCode: resp.StatusCode(),
		Headers:    resp.Header(),
		Body:       resp.Body(),
	}, nil
}

// handleError converts HTTP errors to SDK error types
func (c *Client) handleError(resp *resty.Response) error {
	// Debug logging for error responses
	if c.config.Debug {
		fmt.Printf("[DEBUG] Error Response: Status=%d\n", resp.StatusCode())
		fmt.Printf("[DEBUG] Error Body: %s\n", string(resp.Body()))
		fmt.Printf("[DEBUG] Response Headers:\n")
		for k, v := range resp.Header() {
			fmt.Printf("[DEBUG]   %s: %v\n", k, v)
		}
	}

	var apiErr APIError
	if err := json.Unmarshal(resp.Body(), &apiErr); err == nil && apiErr.ErrorType != "" {
		return &apiErr
	}

	// Try to parse error message from response body
	errorMessage := string(resp.Body())

	switch resp.StatusCode() {
	case 400:
		return &ValidationError{
			StatusCode: resp.StatusCode(),
			Message:    errorMessage,
		}
	case 401:
		// Use actual error message from API if available
		if errorMessage != "" && errorMessage != "{}" {
			return &UnauthorizedError{
				Message: errorMessage,
			}
		}
		return &UnauthorizedError{
			Message: "authentication required",
		}
	case 403:
		if errorMessage != "" && errorMessage != "{}" {
			return &ForbiddenError{
				Message: errorMessage,
			}
		}
		return &ForbiddenError{
			Message: "insufficient permissions",
		}
	case 404:
		return &NotFoundError{
			Message: "resource not found",
		}
	case 429:
		return &RateLimitError{
			RetryAfter: resp.Header().Get("Retry-After"),
			Message:    "rate limit exceeded",
		}
	case 500, 502, 503, 504:
		return &InternalServerError{
			StatusCode: resp.StatusCode(),
			Message:    "internal server error",
			RequestID:  resp.Header().Get("X-Request-ID"),
		}
	default:
		return &APIError{
			Status:    "error",
			ErrorCode: fmt.Sprintf("HTTP_%d", resp.StatusCode()),
			Message:   errorMessage,
		}
	}
}

// HealthCheck performs a lightweight health check on the API
// This is a convenience method that calls Health.GetHealth() and returns only the error.
// It's designed for use in readiness probes and health checks where you only need to know
// if the API is reachable and healthy.
func (c *Client) HealthCheck(ctx context.Context) error {
	health, err := c.Health.GetHealth(ctx)
	if err != nil {
		return err
	}

	// If the healthy boolean is explicitly true, consider it healthy
	if health.Healthy {
		return nil
	}

	// If the healthy field is false/missing, check if status indicates health
	// Some APIs may return status="healthy" but omit the healthy boolean field
	if health.Status == "healthy" || health.Status == "operational" || health.Status == "ok" {
		return nil
	}

	// API is definitively unhealthy
	if health.Status != "" {
		return fmt.Errorf("API is unhealthy: %s", health.Status)
	}
	return fmt.Errorf("API is unhealthy")
}

// Request represents an API request
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Query   map[string]string
	Body    interface{}
	Result  interface{}
	Error   interface{}
}

// Response represents an API response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Service structs for each API domain
type OrganizationsService struct{ client *Client }
type ServersService struct{ client *Client }
type UsersService struct{ client *Client }
type MetricsService struct{ client *Client }
type MonitoringService struct{ client *Client }
type BillingService struct{ client *Client }
type BillingUsageService struct{ client *Client }
type SettingsService struct{ client *Client }
type AlertsService struct{ client *Client }
type AdminService struct{ client *Client }
type StatusPagesService struct{ client *Client }
type ProvidersService struct{ client *Client }
type JobsService struct{ client *Client }
type BackgroundJobsService struct{ client *Client }
type APIKeysService struct{ client *Client }
type SystemService struct{ client *Client }
type TermsService struct{ client *Client }
type EmailQueueService struct{ client *Client }
type PublicService struct{ client *Client }
type DistrosService struct{ client *Client }
type AgentDownloadService struct{ client *Client }
type ControllersService struct{ client *Client }
type HardwareInventoryService struct{ client *Client }
type IPMIService struct{ client *Client }
type SystemdService struct{ client *Client }
type MonitoringDeploymentsService struct{ client *Client }
type NamespaceDeploymentsService struct{ client *Client }
type MonitoringAgentKeysService struct{ client *Client }
type RemoteClustersService struct{ client *Client }
type HealthService struct{ client *Client }
type ProbesService struct{ client *Client }
type IncidentsService struct{ client *Client }
type AgentVersionsService struct{ client *Client }

// getAuthMethod returns a string describing the authentication method being used
func (c *Client) getAuthMethod() string {
	if c.config.Auth.Token != "" {
		return "JWT Token"
	}
	if c.config.Auth.UnifiedAPIKey != "" {
		if c.config.Auth.APIKeySecret != "" {
			return "Unified API Key (Key/Secret)"
		}
		return "Unified API Key (Bearer)"
	}
	if c.config.Auth.RegistrationKey != "" {
		return "Registration Key"
	}
	if c.config.Auth.APIKey != "" && c.config.Auth.APISecret != "" {
		return "API Key/Secret (Legacy)"
	}
	if c.config.Auth.ServerUUID != "" && c.config.Auth.ServerSecret != "" {
		return "Server Credentials"
	}
	if c.config.Auth.MonitoringKey != "" {
		return "Monitoring Key (Legacy)"
	}
	return "None"
}
