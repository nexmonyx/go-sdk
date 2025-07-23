# Nexmonyx Go SDK

The official Go SDK for the Nexmonyx API - a comprehensive server monitoring and management platform.

## Features

- **Multiple Authentication Methods**: JWT tokens, API keys, server credentials, and monitoring keys
- **Complete API Coverage**: Full support for all Nexmonyx API endpoints
- **Type Safety**: Comprehensive Go types for all API models and responses
- **Error Handling**: Structured error types with detailed error information
- **Retry Logic**: Built-in retry mechanism with exponential backoff
- **Rate Limiting**: Automatic handling of rate limit responses
- **Pagination**: Easy-to-use pagination support for list operations
- **Context Support**: Full context.Context support for cancellation and timeouts
- **Debug Mode**: Optional request/response logging for debugging

## Installation

```bash
go get github.com/supporttools/nexmonyx-go-sdk
```

## Quick Start

### JWT Authentication (User API)

```go
package main

import (
    "context"
    "log"
    
    "github.com/supporttools/nexmonyx-go-sdk"
)

func main() {
    config := &nexmonyx.Config{
        BaseURL: "https://api.nexmonyx.com",
        Auth: nexmonyx.AuthConfig{
            Token: "your-jwt-token",
        },
    }

    client, err := nexmonyx.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Get current user
    user, err := client.Users.GetMe(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Current user: %s", user.Email)
}
```

### Server Credentials (Agent API)

```go
config := &nexmonyx.Config{
    BaseURL: "https://api.nexmonyx.com",
    Auth: nexmonyx.AuthConfig{
        ServerUUID:   "your-server-uuid",
        ServerSecret: "your-server-secret",
    },
}

client, err := nexmonyx.NewClient(config)
if err != nil {
    log.Fatal(err)
}

// Send heartbeat
err = client.Servers.Heartbeat(context.Background())
if err != nil {
    log.Fatal(err)
}
```

### API Key Authentication

```go
config := &nexmonyx.Config{
    BaseURL: "https://api.nexmonyx.com",
    Auth: nexmonyx.AuthConfig{
        APIKey:    "your-api-key",
        APISecret: "your-api-secret",
    },
}

client, err := nexmonyx.NewClient(config)
```

## API Services

The SDK is organized into service clients for different API domains:

| Service | Description | Authentication | Example Endpoints |
|---------|-------------|----------------|-------------------|
| **Organizations** | Organization management and membership | JWT, API Key | List, Create, Invite, Members |
| **Servers** | Server registration, monitoring, and management | JWT, Server Credentials | List, Register, Metrics, Credentials |
| **Users** | User profile and preference management | JWT | Profile, Preferences, Avatar |
| **Metrics** | Metrics submission and querying | Server Credentials, JWT | Submit, Query, History |
| **Monitoring** | Probes, regions, and monitoring infrastructure | JWT, Monitoring Key | Probes, Results, Regions |
| **Billing** | Subscription and billing management | JWT | Plans, Checkout, Usage |
| **Settings** | Platform configuration and settings | JWT, Public | Categories, Update, Cache |
| **Alerts** | Alert rules and notification channels | JWT | Rules, Contacts, Silences |
| **StatusPages** | Public status page management | JWT, Public | Create, Publish, History |
| **VMs** | Virtual machine and cloud provider management | JWT | Providers, Create, Lifecycle |
| **Jobs** | Background job and task management | JWT | Create, Monitor, Admin |
| **APIKeys** | API key creation and management | JWT | Create, Scopes, Monitor |
| **System** | Health, version, and system status | Public | Health, Readiness, Version |
| **Terms** | Terms of service management | JWT | Accept, Check, History |
| **EmailQueue** | Email delivery and queue management | JWT Admin | Stats, Retry, Monitor |
| **Public** | Public endpoints and statistics | Public | Stats, Newsletter, Testimonials |
| **Distros** | OS distribution icons and metadata | JWT, Public | List, Search, Popular |
| **AgentDownload** | Agent binary downloads | Public, Server | Download, Version, Platform |
| **Controllers** | Microservice health and status | JWT | Heartbeat, Status, Summary |
| **Admin** | Administrative operations | JWT Admin | Users, Organizations, Jobs |

## LLM Decision Tree for Choosing Services

When working with the Nexmonyx SDK, use this decision tree to select the appropriate service:

```
1. What type of operation are you performing?
   ├── User Management → Users service
   ├── Organization Management → Organizations service
   ├── Server Management → Servers service
   ├── Monitoring/Alerting → Monitoring, Alerts services
   ├── Billing/Subscriptions → Billing service
   ├── System Information → System service
   └── Administrative Tasks → Admin service

2. What authentication do you have?
   ├── JWT Token → Most services available
   ├── API Key/Secret → Limited services (Organizations, Admin)
   ├── Server Credentials → Servers, Metrics, AgentDownload
   ├── Monitoring Key → Monitoring service
   └── No Auth → Public, System, AgentDownload, StatusPages (public)

3. What is your use case?
   ├── Building an Agent → Servers, Metrics, AgentDownload
   ├── Building a Dashboard → Users, Organizations, Servers, Monitoring
   ├── Managing Infrastructure → VMs, Servers, Organizations
   ├── Handling Notifications → Alerts, EmailQueue
   ├── Public Website → Public, StatusPages, Distros
   └── Administrative Tool → Admin, Settings, Jobs
```

## Comprehensive Examples

### Organizations

```go
// List organizations with pagination
orgs, meta, err := client.Organizations.List(ctx, &nexmonyx.ListOptions{
    Page:  1,
    Limit: 10,
    Search: "production",
    Sort: "name",
    Order: "asc",
})

// Get organization by UUID (v1 API - requires specific org UUID)
org, err := client.Organizations.Get(ctx, "org-uuid")

// Create organization
orgReq := &nexmonyx.OrganizationCreateRequest{
    Name:        "My Organization",
    Description: "Production infrastructure",
    Industry:    "technology",
}
org, err := client.Organizations.Create(ctx, orgReq)

// Team management
inviteReq := &nexmonyx.OrganizationInviteRequest{
    Email: "user@example.com",
    Role:  "member",
    Message: "Welcome to our monitoring platform!",
}
err = client.Organizations.InviteUser(ctx, org.UUID, inviteReq)

// List members
members, _, err := client.Organizations.ListMembers(ctx, org.UUID, nil)

// Bulk invitations
bulkReq := &nexmonyx.BulkInviteRequest{
    Invitations: []nexmonyx.InvitationRequest{
        {Email: "dev1@example.com", Role: "developer"},
        {Email: "admin@example.com", Role: "admin"},
    },
}
err = client.Organizations.BulkInvite(ctx, org.UUID, bulkReq)
```

### Servers

```go
// List servers with filtering
servers, meta, err := client.Servers.List(ctx, &nexmonyx.ListOptions{
    Search: "web-server",
    Sort:   "hostname",
    Order:  "asc",
    Filters: map[string]string{
        "environment": "production",
        "location":    "us-east-1",
        "status":      "online",
    },
})

// Get server details
server, err := client.Servers.Get(ctx, "server-uuid")

// Register new server (agent use case)
regReq := &nexmonyx.ServerCreateRequest{
    Hostname:       "web-01.example.com",
    Environment:    "production",
    Location:       "us-east-1",
    Classification: "web_server",
    Tags: map[string]string{
        "team":     "backend",
        "project":  "ecommerce",
        "tier":     "production",
    },
}
regResp, err := client.Servers.Register(ctx, "registration-key", regReq)

// Get server credentials
creds, err := client.Servers.GetCredentials(ctx, "server-uuid")

// Regenerate server secret
newSecret, err := client.Servers.RegenerateSecret(ctx, "server-uuid")

// Get comprehensive metrics
timeRange := &nexmonyx.TimeRange{
    Start: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
    End:   time.Now().Format(time.RFC3339),
}
metrics, err := client.Servers.GetMetrics(ctx, "server-uuid", timeRange)

// Get specific metric types
cpuMetrics, err := client.Servers.GetCPUMetrics(ctx, "server-uuid", timeRange)
memoryMetrics, err := client.Servers.GetMemoryMetrics(ctx, "server-uuid", timeRange)
diskMetrics, err := client.Servers.GetDiskMetrics(ctx, "server-uuid", timeRange)

// Get ZFS metrics (if applicable)
zfsMetrics, err := client.Servers.GetZFSMetrics(ctx, "server-uuid", timeRange)
```

### Users

```go
// Get current user
user, err := client.Users.GetMe(ctx)

// Update user profile
updateReq := &nexmonyx.UserUpdateRequest{
    FirstName: &firstName,
    LastName:  &lastName,
    JobTitle:  &jobTitle,
    Company:   &company,
    Phone:     &phone,
}
user, err := client.Users.UpdateMe(ctx, updateReq)

// Avatar management
avatar, err := client.Users.GetAvatar(ctx)
err = client.Users.UpdateAvatar(ctx, avatarData)
err = client.Users.DeleteAvatar(ctx)
defaultAvatar, err := client.Users.GenerateDefaultAvatar(ctx)

// User preferences
prefs, err := client.Users.GetPreferences(ctx, userID)
prefs.Theme = "dark"
prefs.EmailNotifications = true
prefs.Timezone = "America/New_York"
prefs, err = client.Users.UpdatePreferences(ctx, userID, prefs)

// Update single preference
err = client.Users.UpdateSinglePreference(ctx, userID, "theme", "dark")

// Search users (admin)
searchResults, _, err := client.Users.Search(ctx, &nexmonyx.UserSearchRequest{
    Query: "john@example.com",
    Role:  "admin",
})
```

### Metrics

```go
// Submit comprehensive metrics (agent use case)
metrics := &nexmonyx.ComprehensiveMetricsRequest{
    ServerUUID:  "server-uuid",
    CollectedAt: time.Now().Format(time.RFC3339),
    SystemInfo: &nexmonyx.SystemInfo{
        Hostname:      "web-server-01",
        OS:            "Ubuntu",
        OSVersion:     "22.04 LTS",
        Kernel:        "5.15.0-47-generic",
        Uptime:        3600,
        BootTime:      time.Now().Add(-1*time.Hour).Unix(),
        Processes:     142,
        UsersLoggedIn: 2,
    },
    CPU: &nexmonyx.CPUMetrics{
        UsagePercent:    45.2,
        LoadAverage1:    1.2,
        LoadAverage5:    1.5,
        LoadAverage15:   1.8,
        CoreCount:       4,
        ThreadCount:     8,
        Frequency:       2400,
        CacheSize:       8192,
        UserPercent:     25.5,
        SystemPercent:   19.7,
        IdlePercent:     54.8,
        IOWaitPercent:   2.1,
        IRQPercent:      0.3,
        SoftIRQPercent:  0.6,
    },
    Memory: &nexmonyx.MemoryMetrics{
        TotalBytes:     8589934592,
        UsedBytes:      3865470976,
        FreeBytes:      1073741824,
        AvailableBytes: 4294967296,
        BuffersBytes:   268435456,
        CachedBytes:    1073741824,
        UsagePercent:   45.1,
        SwapTotal:      2147483648,
        SwapUsed:       0,
        SwapFree:       2147483648,
    },
    Disk: []nexmonyx.DiskMetrics{
        {
            Device:           "/dev/sda1",
            Mountpoint:       "/",
            Filesystem:       "ext4",
            TotalBytes:       107374182400,
            UsedBytes:        53687091200,
            FreeBytes:        53687091200,
            UsagePercent:     50.0,
            InodesTotal:      6553600,
            InodesUsed:       327680,
            InodesFree:       6225920,
            InodesUsedPercent: 5.0,
            ReadBytes:        1073741824,
            WriteBytes:       536870912,
            ReadOps:          1000,
            WriteOps:         500,
            ReadTime:         100,
            WriteTime:        50,
        },
    },
    Network: []nexmonyx.NetworkMetrics{
        {
            Interface:    "eth0",
            BytesRecv:    1073741824,
            BytesSent:    536870912,
            PacketsRecv:  1000000,
            PacketsSent:  500000,
            ErrorsRecv:   0,
            ErrorsSent:   0,
            DroppedRecv:  0,
            DroppedSent:  0,
            Speed:        1000000000, // 1 Gbps
            Duplex:       "full",
            MTU:          1500,
        },
    },
}

err = client.Metrics.SubmitComprehensive(ctx, metrics)

// Query historical metrics
query := &nexmonyx.MetricsQuery{
    ServerUUIDs: []string{"server-uuid"},
    MetricTypes: []string{"cpu", "memory", "disk", "network"},
    TimeRange: nexmonyx.TimeRange{
        Start: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
        End:   time.Now().Format(time.RFC3339),
    },
    Aggregation: "avg",
    Interval:    "1h",
    GroupBy:     []string{"hostname", "environment"},
}

data, err := client.Metrics.Query(ctx, query)

// Get latest metrics for dashboard
latest, err := client.Metrics.GetLatest(ctx, "server-uuid")

// Get metrics with time range and aggregation
rangeData, err := client.Metrics.GetRange(ctx, "server-uuid", &nexmonyx.MetricsRangeRequest{
    Start:       time.Now().Add(-6*time.Hour).Format(time.RFC3339),
    End:         time.Now().Format(time.RFC3339),
    Interval:    "5m",
    MetricTypes: []string{"cpu", "memory"},
    Aggregation: "avg",
})
```

### Monitoring (Probes)

```go
// List monitoring regions
regions, err := client.Monitoring.ListRegions(ctx)

// Create HTTP probe
probeReq := &nexmonyx.ProbeCreateRequest{
    Name:     "Website Health Check",
    Type:     "http",
    Target:   "https://example.com",
    Interval: 60,
    Timeout:  30,
    Regions:  []string{"us-east-1", "eu-west-1"},
    Configuration: map[string]interface{}{
        "method":           "GET",
        "expected_status":  200,
        "follow_redirects": true,
        "headers": map[string]string{
            "User-Agent": "Nexmonyx-Monitor/1.0",
        },
        "body": "",
        "verify_ssl": true,
    },
    AlertChannels: []string{"channel-uuid-1"},
}

probe, err := client.Monitoring.CreateProbe(ctx, probeReq)

// Create TCP probe
tcpProbeReq := &nexmonyx.ProbeCreateRequest{
    Name:     "Database Connection Check",
    Type:     "tcp",
    Target:   "db.example.com:5432",
    Interval: 300,
    Timeout:  10,
    Configuration: map[string]interface{}{
        "port": 5432,
    },
}

tcpProbe, err := client.Monitoring.CreateProbe(ctx, tcpProbeReq)

// Get probe results
timeRange := &nexmonyx.TimeRange{
    Start: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
    End:   time.Now().Format(time.RFC3339),
}

results, err := client.Monitoring.GetProbeResults(ctx, probe.UUID, timeRange)

// Get probe uptime statistics
uptime, err := client.Monitoring.GetProbeUptime(ctx, probe.UUID, timeRange)

// Get probe incidents
incidents, err := client.Monitoring.GetProbeIncidents(ctx, probe.UUID, timeRange)

// Get regional status breakdown
regionalStatus, err := client.Monitoring.GetProbeRegionalStatus(ctx, probe.UUID)

// Test probe configuration
testResult, err := client.Monitoring.TestProbe(ctx, probe.UUID)

// Manage probe alert channels
channelReq := &nexmonyx.ProbeAlertChannelRequest{
    Type: "slack",
    Configuration: map[string]interface{}{
        "webhook_url": "https://hooks.slack.com/...",
        "channel":     "#alerts",
        "username":    "Nexmonyx Monitor",
    },
}

channel, err := client.Monitoring.CreateProbeAlertChannel(ctx, probe.UUID, channelReq)

// Toggle probe status
err = client.Monitoring.ToggleProbe(ctx, probe.UUID, true) // enable
err = client.Monitoring.ToggleProbe(ctx, probe.UUID, false) // disable
```

### Alerts

```go
// Create alert rule
alertReq := &nexmonyx.AlertCreateRequest{
    Name:        "High CPU Usage",
    Description: "Alert when CPU usage exceeds 80% for 5 minutes",
    MetricType:  "cpu",
    Condition:   "greater_than",
    Threshold:   80.0,
    Duration:    300, // 5 minutes
    Severity:    "warning",
    ServerUUIDs: []string{"server-uuid-1", "server-uuid-2"},
    Tags: map[string]string{
        "environment": "production",
        "team":        "infrastructure",
    },
}

alert, err := client.Alerts.Create(ctx, alertReq)

// Create alert contact
contactReq := &nexmonyx.AlertContactCreateRequest{
    Name:  "John Doe",
    Email: "john@example.com",
    Phone: "+1234567890",
    Type:  "primary",
}

contact, err := client.Alerts.CreateContact(ctx, contactReq)

// Create notification channels
slackChannelReq := &nexmonyx.AlertChannelCreateRequest{
    Name: "Slack Notifications",
    Type: "slack",
    Configuration: map[string]interface{}{
        "webhook_url": "https://hooks.slack.com/...",
        "channel":     "#alerts",
        "username":    "Nexmonyx Alerts",
        "icon_emoji":  ":warning:",
    },
}

slackChannel, err := client.Alerts.CreateChannel(ctx, slackChannelReq)

emailChannelReq := &nexmonyx.AlertChannelCreateRequest{
    Name: "Email Notifications",
    Type: "email",
    Configuration: map[string]interface{}{
        "recipients": []string{"alerts@example.com", "oncall@example.com"},
        "subject_template": "[ALERT] {{.Severity}}: {{.Name}}",
    },
}

emailChannel, err := client.Alerts.CreateChannel(ctx, emailChannelReq)

// List active alerts
activeAlerts, _, err := client.Alerts.GetActiveAlerts(ctx, nil)

// Acknowledge alert
err = client.Alerts.AcknowledgeAlert(ctx, "alert-instance-id", &nexmonyx.AlertAcknowledgeRequest{
    Message: "Investigating the issue",
})

// Resolve alert
err = client.Alerts.ResolveAlert(ctx, "alert-instance-id", &nexmonyx.AlertResolveRequest{
    Message: "Issue resolved - CPU usage normalized",
})

// Create silence
silenceReq := &nexmonyx.SilenceCreateRequest{
    Name:        "Maintenance Window",
    Description: "Scheduled maintenance for web servers",
    StartTime:   time.Now().Add(1 * time.Hour).Format(time.RFC3339),
    EndTime:     time.Now().Add(3 * time.Hour).Format(time.RFC3339),
    ServerUUIDs: []string{"server-uuid-1"},
    AlertRules:  []string{"alert-rule-id"},
}

silence, err := client.Alerts.CreateSilence(ctx, silenceReq)

// Get alert metrics and statistics
metrics, err := client.Alerts.GetMetrics(ctx, &nexmonyx.AlertMetricsRequest{
    TimeRange: &nexmonyx.TimeRange{
        Start: time.Now().Add(-7*24*time.Hour).Format(time.RFC3339),
        End:   time.Now().Format(time.RFC3339),
    },
})
```

### Billing

```go
// Get subscription details
subscription, err := client.Billing.GetSubscription(ctx, "org-uuid")

// Get available plans
plans, err := client.Billing.GetPlans(ctx)

// Get plan features
features, err := client.Billing.GetPlanFeatures(ctx)

// Create checkout session
checkoutReq := &nexmonyx.CheckoutSessionRequest{
    PriceID:    "price_1234567890",
    SuccessURL: "https://myapp.com/billing/success",
    CancelURL:  "https://myapp.com/billing/cancel",
    Metadata: map[string]string{
        "upgrade_from": "basic",
        "user_id":      "user-123",
    },
}

session, err := client.Billing.CreateCheckoutSession(ctx, "org-uuid", checkoutReq)

// Update subscription (upgrade/downgrade)
updateReq := &nexmonyx.SubscriptionUpdateRequest{
    PriceID: "price_new_plan",
    ProrationBehavior: "always_invoice",
}

updatedSub, err := client.Billing.UpdateSubscription(ctx, "org-uuid", updateReq)

// Get usage and billing history
usage, err := client.Billing.GetUsage(ctx, "org-uuid", &nexmonyx.UsageRequest{
    StartDate: time.Now().AddDate(0, -1, 0).Format("2006-01-02"),
    EndDate:   time.Now().Format("2006-01-02"),
})

history, _, err := client.Billing.GetBillingHistory(ctx, "org-uuid", nil)

// Create customer portal session
portalReq := &nexmonyx.PortalSessionRequest{
    ReturnURL: "https://myapp.com/billing",
}

portalSession, err := client.Billing.CreatePortalSession(ctx, "org-uuid", portalReq)

// Cancel subscription
cancelReq := &nexmonyx.SubscriptionCancelRequest{
    Reason: "cost_optimization",
    Feedback: "Switching to smaller plan",
    CancelAtPeriodEnd: true,
}

err = client.Billing.CancelSubscription(ctx, "org-uuid", cancelReq)
```

### Settings

```go
// Get all public settings
publicSettings, err := client.Settings.ListPublicSettings(ctx)

// Get settings by category
apiSettings, err := client.Settings.GetSettingsByCategory(ctx, "api")
uiSettings, err := client.Settings.GetSettingsByCategory(ctx, "ui")

// Get specific setting
setting, err := client.Settings.GetSetting(ctx, "api.rate_limit")

// Get setting categories
categories, err := client.Settings.GetCategories(ctx)

// Update setting (admin only)
updateReq := &nexmonyx.SettingUpdateRequest{
    Value:       "new-value",
    Description: "Updated description",
}

setting, err := client.Settings.UpdateSetting(ctx, "setting-key", updateReq)

// Bulk update settings (admin only)
updates := map[string]interface{}{
    "api.rate_limit":              1000,
    "ui.theme":                    "dark",
    "agent.default_interval":      60,
    "alerts.max_per_organization": 100,
}

settings, err := client.Settings.BulkUpdate(ctx, updates)

// Create new setting (admin only)
createReq := &nexmonyx.SettingCreateRequest{
    Key:         "custom.feature_flag",
    Value:       "enabled",
    Type:        "string",
    Category:    "custom",
    Permission:  "authenticated",
    Description: "Custom feature flag",
    IsCacheable: true,
}

newSetting, err := client.Settings.CreateSetting(ctx, createReq)
```

### Status Pages

```go
// Create status page
statusPageReq := &nexmonyx.CreateStatusPageRequest{
    Name:        "Service Status",
    Slug:        "service-status",
    Title:       "Our Service Status",
    Description: "Real-time status of our services",
    Theme: nexmonyx.StatusPageTheme{
        PrimaryColor:    "#007bff",
        SecondaryColor:  "#6c757d",
        BackgroundColor: "#ffffff",
        TextColor:       "#212529",
        LogoURL:         "https://example.com/logo.png",
        FaviconURL:      "https://example.com/favicon.ico",
        CustomCSS:       ".custom { font-weight: bold; }",
    },
    Probes:           []string{"probe-uuid-1", "probe-uuid-2"},
    IsPublic:         true,
    ShowDetailedInfo: true,
    ContactInfo: nexmonyx.StatusPageContact{
        Email:   "support@example.com",
        Phone:   "+1234567890",
        Website: "https://example.com/support",
    },
    SocialLinks: nexmonyx.StatusPageSocial{
        Twitter:   "@example",
        Facebook:  "example",
        LinkedIn:  "company/example",
    },
}

statusPage, _, err := client.StatusPages.Create(ctx, statusPageReq)

// List status pages
pages, _, err := client.StatusPages.List(ctx, &nexmonyx.ListOptions{
    Search: "production",
})

// Get status page details
page, _, err := client.StatusPages.Get(ctx, statusPage.ID)

// Update status page
updateReq := &nexmonyx.UpdateStatusPageRequest{
    Title:       "Updated Service Status",
    Description: "Updated description",
    IsPublic:    true,
}

updatedPage, _, err := client.StatusPages.Update(ctx, statusPage.ID, updateReq)

// Get public status page (no authentication required)
publicPage, _, err := client.StatusPages.GetPublic(ctx, "service-status")

// Get status page history
history, _, err := client.StatusPages.GetPublicHistory(ctx, "service-status", &nexmonyx.ListOptions{
    Limit: 50,
})

// Get status page incidents
incidents, _, err := client.StatusPages.GetPublicIncidents(ctx, "service-status", &nexmonyx.ListOptions{
    Limit: 10,
})

// Delete status page
err = client.StatusPages.Delete(ctx, statusPage.ID)
```

### Virtual Machines

```go
// Test cloud provider credentials
testReq := &nexmonyx.TestProviderRequest{
    Type: "aws",
    Credentials: map[string]interface{}{
        "access_key": "AKIA...",
        "secret_key": "xxx",
        "region":     "us-east-1",
    },
}

testResult, _, err := client.VMs.TestProvider(ctx, "org-id", testReq)

// Create cloud provider
providerReq := &nexmonyx.CreateProviderRequest{
    Name:        "AWS Production",
    Type:        "aws",
    Description: "Production AWS account",
    Credentials: map[string]interface{}{
        "access_key": "AKIA...",
        "secret_key": "xxx",
        "region":     "us-east-1",
    },
    Tags: map[string]string{
        "environment": "production",
        "team":        "infrastructure",
    },
}

provider, _, err := client.VMs.CreateProvider(ctx, "org-id", providerReq)

// List cloud providers
providers, _, err := client.VMs.ListProviders(ctx, "org-id", &nexmonyx.ListOptions{
    Search: "aws",
})

// Create virtual machine
vmReq := &nexmonyx.CreateVMRequest{
    Name:         "web-server-01",
    ProviderID:   provider.ID,
    Region:       "us-east-1",
    InstanceType: "t3.micro",
    ImageID:      "ami-0abcdef123456789",
    KeyPairName:  "my-keypair",
    SecurityGroups: []string{"sg-12345", "sg-67890"},
    SubnetID:     "subnet-abc123",
    UserData:     base64.StdEncoding.EncodeToString([]byte("#!/bin/bash\napt-get update")),
    Tags: map[string]string{
        "Environment": "production",
        "Team":        "backend",
        "Project":     "ecommerce",
    },
    MonitoringEnabled: true,
    BackupEnabled:     true,
}

vm, _, err := client.VMs.CreateVM(ctx, "org-id", vmReq)

// List VMs
vms, _, err := client.VMs.ListVMs(ctx, "org-id", &nexmonyx.ListOptions{
    Filters: map[string]string{
        "status":      "running",
        "environment": "production",
    },
})

// Get VM details
vmDetails, _, err := client.VMs.GetVMDetails(ctx, "org-id", vm.ID)

// VM lifecycle management
startResp, _, err := client.VMs.StartVM(ctx, "org-id", vm.ID)
stopResp, _, err := client.VMs.StopVM(ctx, "org-id", vm.ID)
restartResp, _, err := client.VMs.RestartVM(ctx, "org-id", vm.ID)

// Delete VM
err = client.VMs.DeleteVM(ctx, "org-id", vm.ID)
```

### Background Jobs

```go
// Create background job
jobReq := &nexmonyx.CreateJobRequest{
    Type:        "server_backup",
    Description: "Full backup of production server",
    Priority:    "high",
    Metadata: map[string]interface{}{
        "server_id":      "server-123",
        "backup_type":    "full",
        "s3_bucket":      "backups",
        "retention_days": 30,
        "compression":    "gzip",
        "encryption":     true,
    },
    ScheduledFor: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
    Timeout:      3600, // 1 hour
    RetryCount:   3,
}

job, _, err := client.Jobs.Create(ctx, jobReq)

// List jobs with filtering
jobs, _, err := client.Jobs.List(ctx, &nexmonyx.ListOptions{
    Filters: map[string]string{
        "status": "pending",
        "type":   "server_backup",
    },
    Sort:  "created_at",
    Order: "desc",
})

// Monitor job progress
status, _, err := client.Jobs.GetStatus(ctx, job.ID)
fmt.Printf("Job %s: %s (%d%% complete)\n", job.ID, status.Status, status.Progress)

// Update job status (for job processors)
updateReq := &nexmonyx.UpdateJobRequest{
    Status:   "running",
    Progress: 50,
    Message:  "Processing backup...",
    Metadata: map[string]interface{}{
        "files_processed": 1500,
        "bytes_uploaded":  1073741824, // 1GB
    },
}

updatedJob, _, err := client.Jobs.Update(ctx, job.ID, updateReq)

// Get job details
jobDetails, _, err := client.Jobs.Get(ctx, job.ID)

// Delete completed job
err = client.Jobs.Delete(ctx, job.ID)

// Admin operations
// List all jobs with advanced filtering
jobFilters := &nexmonyx.JobFilters{
    Status:       []string{"pending", "running"},
    Type:         []string{"server_backup", "vm_deployment"},
    UserID:       "user-123",
    CreatedAfter: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
    Priority:     []string{"high", "critical"},
}

allJobs, _, err := client.Jobs.AdminListJobs(ctx, jobFilters, &nexmonyx.ListOptions{
    Page:    1,
    PerPage: 25,
})

// Get job statistics
stats, _, err := client.Jobs.AdminGetStatistics(ctx)
fmt.Printf("Total jobs: %d, Running: %d, Failed: %d\n", 
    stats.Total, stats.ByStatus["running"], stats.ByStatus["failed"])

// Get detailed job information (admin)
adminJobDetails, _, err := client.Jobs.AdminGetJobDetails(ctx, "job-id")

// Cancel or retry jobs (admin)
cancelResp, _, err := client.Jobs.AdminCancel(ctx, "job-id")
retryResp, _, err := client.Jobs.AdminRetry(ctx, "failed-job-id")
```

### API Keys

```go
// Create API key with specific scopes
keyReq := &nexmonyx.CreateAPIKeyRequest{
    Name:        "CI/CD Pipeline",
    Description: "Key for automated deployments and monitoring",
    Scopes: []string{
        "servers:read",
        "servers:write", 
        "metrics:write",
        "jobs:create",
        "jobs:read",
    },
    ExpiresAt:    time.Now().AddDate(1, 0, 0).Format(time.RFC3339), // 1 year
    IPWhitelist:  []string{"192.168.1.0/24", "10.0.0.0/8"},
    RateLimitRPM: 1000, // requests per minute
    Tags: map[string]string{
        "environment": "production",
        "team":        "devops",
        "purpose":     "automation",
    },
}

apiKey, _, err := client.APIKeys.Create(ctx, keyReq)

// IMPORTANT: Store these securely - secrets won't be shown again
fmt.Printf("API Key: %s\n", apiKey.Key)
fmt.Printf("Secret: %s\n", apiKey.Secret)

// List API keys (secrets not included)
keys, _, err := client.APIKeys.List(ctx, &nexmonyx.ListOptions{
    Search: "CI/CD",
    Filters: map[string]string{
        "is_active": "true",
        "scope":     "servers:read",
    },
})

// Get API key details
keyDetails, _, err := client.APIKeys.Get(ctx, apiKey.ID)

// Update API key
updateReq := &nexmonyx.UpdateAPIKeyRequest{
    Description: "Updated: Key for automated deployments and monitoring",
    Scopes:      []string{"servers:read", "metrics:write", "jobs:read"},
    IsActive:    &[]bool{true}[0], // Pointer to bool
    IPWhitelist: []string{"192.168.1.0/24"},
    RateLimitRPM: 500,
}

updatedKey, _, err := client.APIKeys.Update(ctx, apiKey.ID, updateReq)

// Rotate API key (generates new secret)
rotatedKey, _, err := client.APIKeys.Rotate(ctx, apiKey.ID)

// Deactivate API key
isActive := false
deactivatedKey, _, err := client.APIKeys.Update(ctx, apiKey.ID, &nexmonyx.UpdateAPIKeyRequest{
    IsActive: &isActive,
})

// Get API key usage statistics
usage, _, err := client.APIKeys.GetUsage(ctx, apiKey.ID, &nexmonyx.UsageTimeRange{
    Start: time.Now().Add(-7*24*time.Hour).Format(time.RFC3339),
    End:   time.Now().Format(time.RFC3339),
})

// Delete API key permanently
err = client.APIKeys.Delete(ctx, apiKey.ID)

// Admin operations

// Get API key audit log (admin only)
auditLog, _, err := client.APIKeys.GetAuditLog(ctx, apiKey.ID, &nexmonyx.ListOptions{
    Limit: 100,
})
```

### Monitoring Agent Keys

Monitoring agent keys are specialized API keys used by monitoring agents to authenticate with the API and submit probe results.

```go
// Admin operations - create monitoring agent keys for region enrollment
adminKeyReq := &nexmonyx.CreateMonitoringAgentKeyRequest{
    OrganizationID:  1,
    RemoteClusterID: nil, // Optional cluster restriction
    Description:     "Production monitoring agent key",
    Capabilities:    "probe_execution,heartbeat",
}

agentKeyResp, err := client.MonitoringAgentKeys.CreateAdmin(ctx, adminKeyReq)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created monitoring agent key: %s\n", agentKeyResp.FullToken)

// Customer operations - self-service key creation
customerKeyResp, err := client.MonitoringAgentKeys.Create(ctx, "org-uuid", "Development environment monitoring")
if err != nil {
    log.Fatal(err)
}

// List organization's monitoring agent keys
keys, pagination, err := client.MonitoringAgentKeys.List(ctx, "org-uuid", &nexmonyx.ListMonitoringAgentKeysOptions{
    Page:      1,
    Limit:     50,
    Namespace: "production",
})

for _, key := range keys {
    fmt.Printf("Key: %s, Status: %s, Description: %s\n", 
        key.KeyPrefix, key.Status, key.Description)
}

// Revoke a monitoring agent key
err = client.MonitoringAgentKeys.Revoke(ctx, "org-uuid", "mag_abc123")
if err != nil {
    log.Fatal(err)
}
```

### System Health and Information

```go
// Check API health (no authentication required)
health, _, err := client.System.GetHealth(ctx)
if health.Status == "healthy" {
    fmt.Println("API is healthy")
}

// Check API readiness (no authentication required)
readiness, _, err := client.System.GetReadiness(ctx)
if readiness.Status == "ready" {
    fmt.Println("API is ready to serve requests")
}

// Get API version information (no authentication required)
version, _, err := client.System.GetVersion(ctx)
fmt.Printf("API Version: %s, Build: %s\n", version.Version, version.Build)

// Get comprehensive system status
systemStatus, _, err := client.System.GetSystemStatus(ctx)
fmt.Printf("Database: %s, Cache: %s, Queue: %s\n", 
    systemStatus.Database.Status,
    systemStatus.Cache.Status, 
    systemStatus.Queue.Status)

// Get system metrics
sysMetrics, _, err := client.System.GetSystemMetrics(ctx)
fmt.Printf("CPU: %.1f%%, Memory: %.1f%%, Disk: %.1f%%\n",
    sysMetrics.CPU.UsagePercent,
    sysMetrics.Memory.UsagePercent,
    sysMetrics.Disk.UsagePercent)
```

### Terms of Service

```go
// Accept terms of service
acceptReq := &nexmonyx.TermsAcceptanceRequest{
    TermsVersion: "v2.1",
    TermsType:    "terms_of_service",
    IPAddress:    "192.168.1.100",
    UserAgent:    "Mozilla/5.0...",
    AcceptedAt:   time.Now().Format(time.RFC3339),
}

acceptance, _, err := client.Terms.AcceptTerms(ctx, acceptReq)

// Get all terms acceptances for current user
acceptances, _, err := client.Terms.GetAcceptances(ctx)

// Check if user has accepted specific terms
hasAccepted, err := client.Terms.HasAcceptedTerms(ctx, "v2.1", "terms_of_service")
if hasAccepted {
    fmt.Println("User has accepted the latest terms")
}

// Get latest terms acceptance
latest, _, err := client.Terms.GetLatestAcceptance(ctx)
if latest != nil {
    fmt.Printf("Latest acceptance: %s %s at %s\n", 
        latest.TermsType, latest.TermsVersion, latest.AcceptedAt)
}

// Check acceptance for privacy policy
privacyAccepted, err := client.Terms.HasAcceptedTerms(ctx, "v1.5", "privacy_policy")

// Get acceptance history
history, _, err := client.Terms.GetAcceptanceHistory(ctx, &nexmonyx.ListOptions{
    Sort:  "accepted_at",
    Order: "desc",
})
```

### Email Queue Management (Admin)

```go
// Get email queue statistics (admin only)
stats, _, err := client.EmailQueue.GetStats(ctx)
fmt.Printf("Total: %d, Pending: %d, Sent: %d, Failed: %d\n",
    stats.TotalEmails, stats.PendingEmails, stats.SentEmails, stats.FailedEmails)

// List emails with filters
emails, _, err := client.EmailQueue.List(ctx, &nexmonyx.EmailFilters{
    Status:      []string{"pending", "failed"},
    Type:        []string{"invitation", "alert"},
    RecipientEmail: "user@example.com",
    CreatedAfter: time.Now().Add(-24*time.Hour).Format(time.RFC3339),
}, &nexmonyx.ListOptions{
    Page:  1,
    Limit: 50,
    Sort:  "created_at",
    Order: "desc",
})

// Get email details
email, _, err := client.EmailQueue.GetEmailDetails(ctx, "email-id")

// Resend specific email
resendResp, _, err := client.EmailQueue.ResendEmail(ctx, "email-id")

// Update email status
updateReq := &nexmonyx.EmailUpdateRequest{
    Status:  "pending",
    Message: "Retrying delivery",
}

updatedEmail, _, err := client.EmailQueue.UpdateEmail(ctx, "email-id", updateReq)

// Get pending emails
pendingEmails, _, err := client.EmailQueue.GetPendingEmails(ctx, &nexmonyx.ListOptions{
    Limit: 100,
})

// Get failed emails with error details
failedEmails, _, err := client.EmailQueue.GetFailedEmails(ctx, &nexmonyx.ListOptions{
    Limit: 50,
})

// Retry all failed emails
retryResult, _, err := client.EmailQueue.RetryFailedEmails(ctx)
fmt.Printf("Retried %d failed emails\n", retryResult.Count)

// Delete email from queue
err = client.EmailQueue.DeleteEmail(ctx, "email-id")

// Bulk operations
bulkReq := &nexmonyx.BulkEmailOperationRequest{
    EmailIDs:  []string{"email-1", "email-2", "email-3"},
    Operation: "retry",
}

bulkResult, _, err := client.EmailQueue.BulkOperation(ctx, bulkReq)
```

### Public Endpoints

```go
// Get public platform statistics (no authentication required)
stats, _, err := client.Public.GetStats(ctx)
fmt.Printf("Servers: %d, Organizations: %d, Uptime: %.2f%%\n",
    stats.TotalServers, stats.TotalOrganizations, stats.TotalUptime)

// Get customer testimonials (no authentication required)
testimonials, _, err := client.Public.GetTestimonials(ctx)
fmt.Printf("Found %d testimonials\n", len(testimonials))

// Get featured testimonials
featured, _, err := client.Public.GetFeaturedTestimonials(ctx)

// Newsletter signup
signupReq := &nexmonyx.NewsletterSignupRequest{
    Email:     "user@example.com",
    FirstName: "John",
    LastName:  "Doe",
    Company:   "Example Corp",
    Source:    "website_footer",
    Interests: []string{"product_updates", "best_practices"},
}

signupResp, _, err := client.Public.SignupNewsletter(ctx, signupReq)

// Contact form submission
contactReq := &nexmonyx.ContactFormRequest{
    Name:     "John Doe",
    Email:    "john@example.com",
    Company:  "Example Corp",
    Subject:  "Enterprise Inquiry",
    Message:  "I'm interested in your enterprise features...",
    Type:     "sales",
    Source:   "contact_page",
    Phone:    "+1234567890",
}

contactResp, _, err := client.Public.SubmitContactForm(ctx, contactReq)

// Get platform announcements
announcements, _, err := client.Public.GetAnnouncements(ctx, &nexmonyx.ListOptions{
    Limit: 5,
})

// Get pricing information
pricing, _, err := client.Public.GetPricingInfo(ctx)
```

### OS Distributions

```go
// List all OS distributions
distros, _, err := client.Distros.List(ctx, &nexmonyx.ListOptions{
    Search: "ubuntu",
    Sort:   "name",
    Order:  "asc",
})

// Get popular distributions
popular, _, err := client.Distros.GetPopular(ctx)

// Search distributions
searchResults, _, err := client.Distros.Search(ctx, "centos")

// Get distribution by name
distro, _, err := client.Distros.GetByName(ctx, "ubuntu")

// Get distribution by ID
distroByID, _, err := client.Distros.Get(ctx, "distro-id")

// Get distributions by category
categoryDistros, _, err := client.Distros.GetByCategory(ctx, "enterprise")

// Get distribution categories
categories, _, err := client.Distros.GetCategories(ctx)

// Get distribution statistics
distroStats, _, err := client.Distros.GetStatistics(ctx)

// Create distribution (admin with API key)
createReq := &nexmonyx.CreateDistroRequest{
    Name:        "custom-linux",
    DisplayName: "Custom Linux Distribution",
    Category:    "custom",
    IconURL:     "https://example.com/icon.png",
    Website:     "https://customlinux.org",
    Description: "A custom Linux distribution for specialized use cases",
    Tags:        []string{"custom", "specialized", "enterprise"},
    IsActive:    true,
}

newDistro, _, err := client.Distros.Create(ctx, createReq)
```

### Agent Download

```go
// Download latest agent binary (no authentication required)
agentResp, _, err := client.AgentDownload.DownloadLatestAgent(ctx)
fmt.Printf("Downloaded: %s (%d bytes)\n", agentResp.Filename, agentResp.Size)

// Download latest AMD64 agent
amd64Resp, _, err := client.AgentDownload.DownloadLatestAgentAMD64(ctx)

// Download specific version
versionResp, _, err := client.AgentDownload.DownloadAgent(ctx, "v1.2.3")

// Download for specific platform
platformResp, _, err := client.AgentDownload.DownloadAgentForPlatform(ctx, "latest", "linux", "amd64", true)

// Get agent version information (requires server credentials)
// Create client with server credentials first
serverClient, err := nexmonyx.NewClient(&nexmonyx.Config{
    BaseURL: "https://api.nexmonyx.com",
    Auth: nexmonyx.AuthConfig{
        ServerUUID:   "server-uuid",
        ServerSecret: "server-secret",
    },
})

versionInfo, _, err := serverClient.AgentDownload.GetVersion(ctx)
fmt.Printf("Version: %s, Platform: %s/%s\n", 
    versionInfo.Version, versionInfo.Platform, versionInfo.Architecture)
```

### Controllers and Microservices

```go
// Submit controller heartbeat (from microservice)
heartbeatReq := &nexmonyx.ControllerHeartbeatRequest{
    ControllerName: "billing-controller",
    Status:         "healthy",
    Version:        "v1.2.3",
    LastSeen:       time.Now().Format(time.RFC3339),
    Metadata: map[string]interface{}{
        "cpu_usage":     25.5,
        "memory_usage":  45.2,
        "active_jobs":   12,
        "queue_depth":   5,
        "uptime":        3600,
    },
    Health: nexmonyx.ControllerHealth{
        Database:    "healthy",
        Cache:       "healthy", 
        Queue:       "healthy",
        ExternalAPI: "healthy",
    },
}

heartbeatResp, _, err := client.Controllers.SubmitHeartbeat(ctx, "billing-controller", heartbeatReq)

// List all controllers (admin/monitoring)
controllers, _, err := client.Controllers.List(ctx, &nexmonyx.ListOptions{
    Filters: map[string]string{
        "status": "healthy",
    },
})

// Get controller summary
summary, _, err := client.Controllers.GetSummary(ctx)
fmt.Printf("Total: %d, Healthy: %d, Unhealthy: %d\n",
    summary.TotalControllers, summary.HealthyControllers, summary.UnhealthyControllers)

// Get specific controller status
status, _, err := client.Controllers.GetStatus(ctx, "billing-controller")
fmt.Printf("Controller: %s, Status: %s, Last Seen: %s\n",
    status.Name, status.Status, status.LastSeen)

// Delete controller record (admin only)
err = client.Controllers.Delete(ctx, "old-controller")
```

## Error Handling

The SDK provides structured error types for different scenarios:

```go
_, err := client.Users.GetMe(ctx)
if err != nil {
    switch e := err.(type) {
    case *nexmonyx.APIError:
        // Standard API error with details
        log.Printf("API Error: %s - %s", e.Error, e.Message)
        if e.Details != "" {
            log.Printf("Details: %s", e.Details)
        }
        
    case *nexmonyx.UnauthorizedError:
        // Authentication failed
        log.Printf("Unauthorized: %s", e.Message)
        // Redirect to login or refresh token
        
    case *nexmonyx.NotFoundError:
        // Resource not found
        log.Printf("Not found: %s %s", e.Resource, e.ID)
        
    case *nexmonyx.RateLimitError:
        // Rate limit exceeded
        log.Printf("Rate limited: %s", e.Message)
        if e.RetryAfter != "" {
            log.Printf("Retry after: %s", e.RetryAfter)
            // Wait before retrying
        }
        
    case *nexmonyx.ForbiddenError:
        // Insufficient permissions
        log.Printf("Forbidden: %s", e.Error())
        
    case *nexmonyx.ValidationError:
        // Request validation failed
        log.Printf("Validation error: %s", e.Message)
        for field, errors := range e.Errors {
            log.Printf("  %s: %v", field, errors)
        }
        
    case *nexmonyx.ConflictError:
        // Resource conflict (e.g., duplicate)
        log.Printf("Conflict: %s", e.Message)
        
    case *nexmonyx.InternalServerError:
        // Server-side error
        log.Printf("Internal server error: %s", e.Message)
        // Implement retry logic with exponential backoff
        
    default:
        // Network or other errors
        log.Printf("Unknown error: %v", err)
    }
}
```

## Configuration Options

The SDK supports extensive configuration options:

```go
config := &nexmonyx.Config{
    // Base URL (default: https://api.nexmonyx.com)
    BaseURL: "https://api-staging.nexmonyx.com",
    
    // Authentication (choose one method)
    Auth: nexmonyx.AuthConfig{
        // JWT Token (for user authentication)
        Token: "jwt-token",
        
        // API Key authentication (for service-to-service)
        APIKey:    "your-api-key",
        APISecret: "your-api-secret",
        
        // Server credentials (for agents)
        ServerUUID:   "server-uuid", 
        ServerSecret: "server-secret",
        
        // Monitoring key (for monitoring agents)
        MonitoringKey: "monitoring-key",
    },
    
    // HTTP client configuration
    HTTPClient: &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:       100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:    90 * time.Second,
        },
    },
    
    // Request timeout (default: 30s)
    Timeout: 60 * time.Second,
    
    // Custom headers for all requests
    Headers: map[string]string{
        "X-Custom-Header":    "value",
        "X-Application-Name": "my-app",
        "X-Version":          "v1.0.0",
    },
    
    // Debug mode (enables request/response logging)
    Debug: true,
    
    // Retry configuration
    RetryCount:    5,                    // Number of retries
    RetryWaitTime: 2 * time.Second,      // Initial wait time
    RetryMaxWait:  60 * time.Second,     // Maximum wait time
}

client, err := nexmonyx.NewClient(config)
```

## Pagination

List operations support comprehensive pagination:

```go
opts := &nexmonyx.ListOptions{
    Page:    1,          // Page number (1-based)
    Limit:   25,         // Items per page (max 100)
    PerPage: 25,         // Alternative to Limit
    Sort:    "created_at", // Sort field
    Order:   "desc",     // Sort order (asc/desc)
    Search:  "web-server", // Search query
    Filters: map[string]string{
        "environment": "production",
        "location":    "us-east-1",
        "status":      "online",
        "team":        "backend",
    },
}

servers, meta, err := client.Servers.List(ctx, opts)
if err != nil {
    log.Fatal(err)
}

// Pagination metadata
log.Printf("Page %d of %d, %d total items", 
    meta.Page, meta.TotalPages, meta.TotalItems)
log.Printf("Showing %d items, %d per page", 
    meta.Count, meta.PerPage)

// Iterate through all pages
for page := 1; page <= meta.TotalPages; page++ {
    opts.Page = page
    servers, meta, err := client.Servers.List(ctx, opts)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, server := range servers {
        log.Printf("Server: %s (%s)", server.Hostname, server.Status)
    }
}

// Helper function for processing all pages
func processAllServers(client *nexmonyx.Client, opts *nexmonyx.ListOptions, 
    processor func(server nexmonyx.Server) error) error {
    
    page := 1
    for {
        opts.Page = page
        servers, meta, err := client.Servers.List(ctx, opts)
        if err != nil {
            return err
        }
        
        for _, server := range servers {
            if err := processor(server); err != nil {
                return err
            }
        }
        
        if page >= meta.TotalPages {
            break
        }
        page++
    }
    
    return nil
}
```

## Context and Cancellation

All operations support context.Context for cancellation and timeouts:

```go
// Timeout context
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Cancellation context
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(5 * time.Second)
    cancel() // Cancel the request after 5 seconds
}()

// Request with context
user, err := client.Users.GetMe(ctx)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Request timed out")
    } else if ctx.Err() == context.Canceled {
        log.Println("Request was cancelled")
    }
}

// Context with values (for request tracing)
reqID := "req-" + time.Now().Format("20060102150405")
ctx = context.WithValue(ctx, "request_id", reqID)

// The SDK will automatically add X-Request-ID header
servers, _, err := client.Servers.List(ctx, nil)
```

## Authentication Switching

You can create new clients with different authentication methods:

```go
// Start with JWT token
jwtClient, _ := nexmonyx.NewClient(&nexmonyx.Config{
    Auth: nexmonyx.AuthConfig{Token: "jwt-token"},
})

// Switch to API key authentication
apiKeyClient := jwtClient.WithAPIKey("api-key", "api-secret")

// Switch to server credentials
serverClient := jwtClient.WithServerCredentials("server-uuid", "server-secret")

// Use different clients for different operations
user, err := jwtClient.Users.GetMe(ctx)                    // User operations
metrics, err := serverClient.Metrics.SubmitComprehensive(ctx, data) // Agent operations
orgs, _, err := apiKeyClient.Organizations.List(ctx, nil)  // Service operations
```

## Best Practices

### 1. Client Management
```go
// ✅ Good: Reuse clients
var client *nexmonyx.Client
func init() {
    var err error
    client, err = nexmonyx.NewClient(&nexmonyx.Config{
        Auth: nexmonyx.AuthConfig{Token: os.Getenv("NEXMONYX_TOKEN")},
    })
    if err != nil {
        log.Fatal(err)
    }
}

// ❌ Bad: Creating new clients for each request
func getUser() (*nexmonyx.User, error) {
    client, _ := nexmonyx.NewClient(&nexmonyx.Config{...}) // Inefficient
    return client.Users.GetMe(ctx)
}
```

### 2. Context Usage
```go
// ✅ Good: Always use context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

servers, _, err := client.Servers.List(ctx, opts)

// ❌ Bad: Using background context without timeout
servers, _, err := client.Servers.List(context.Background(), opts)
```

### 3. Error Handling
```go
// ✅ Good: Handle specific error types
user, err := client.Users.GetMe(ctx)
if err != nil {
    switch e := err.(type) {
    case *nexmonyx.UnauthorizedError:
        // Redirect to login
        return redirectToLogin()
    case *nexmonyx.RateLimitError:
        // Implement backoff
        time.Sleep(time.Duration(e.RetryAfter) * time.Second)
        return retryOperation()
    default:
        return fmt.Errorf("unexpected error: %w", err)
    }
}

// ❌ Bad: Generic error handling
user, err := client.Users.GetMe(ctx)
if err != nil {
    log.Printf("Error: %v", err) // No specific handling
}
```

### 4. Pagination
```go
// ✅ Good: Process pages efficiently
opts := &nexmonyx.ListOptions{Limit: 100} // Use larger page sizes
for page := 1; ; page++ {
    opts.Page = page
    servers, meta, err := client.Servers.List(ctx, opts)
    if err != nil {
        return err
    }
    
    for _, server := range servers {
        if err := processServer(server); err != nil {
            return err
        }
    }
    
    if page >= meta.TotalPages {
        break
    }
}

// ❌ Bad: Loading all data at once
opts := &nexmonyx.ListOptions{Limit: 10000} // May cause timeouts
servers, _, err := client.Servers.List(ctx, opts)
```

### 5. Configuration
```go
// ✅ Good: Production-ready configuration
config := &nexmonyx.Config{
    BaseURL: os.Getenv("NEXMONYX_API_URL"),
    Auth: nexmonyx.AuthConfig{
        Token: os.Getenv("NEXMONYX_TOKEN"),
    },
    Timeout:       30 * time.Second,
    RetryCount:    3,
    RetryWaitTime: 1 * time.Second,
    RetryMaxWait:  30 * time.Second,
    Debug:         os.Getenv("NEXMONYX_DEBUG") == "true",
    Headers: map[string]string{
        "X-Application": "my-app",
        "X-Version":     version.Version,
    },
}

// ❌ Bad: Minimal configuration
config := &nexmonyx.Config{
    Auth: nexmonyx.AuthConfig{Token: "hardcoded-token"}, // Security risk
}
```

## Integration Examples

### Agent Implementation
```go
// Complete agent implementation
type Agent struct {
    client     *nexmonyx.Client
    serverUUID string
    config     *AgentConfig
}

func NewAgent(config *AgentConfig) (*Agent, error) {
    client, err := nexmonyx.NewClient(&nexmonyx.Config{
        BaseURL: config.APIEndpoint,
        Auth: nexmonyx.AuthConfig{
            ServerUUID:   config.ServerUUID,
            ServerSecret: config.ServerSecret,
        },
        Timeout:    30 * time.Second,
        RetryCount: 3,
    })
    
    return &Agent{
        client:     client,
        serverUUID: config.ServerUUID,
        config:     config,
    }, nil
}

func (a *Agent) Start() error {
    // Send initial heartbeat
    if err := a.client.Servers.Heartbeat(context.Background()); err != nil {
        return fmt.Errorf("initial heartbeat failed: %w", err)
    }
    
    // Start metrics collection
    ticker := time.NewTicker(time.Duration(a.config.MetricsInterval) * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := a.collectAndSubmitMetrics(); err != nil {
                log.Printf("Failed to submit metrics: %v", err)
            }
        }
    }
}

func (a *Agent) collectAndSubmitMetrics() error {
    metrics := &nexmonyx.ComprehensiveMetricsRequest{
        ServerUUID:  a.serverUUID,
        CollectedAt: time.Now().Format(time.RFC3339),
        SystemInfo:  a.collectSystemInfo(),
        CPU:         a.collectCPUMetrics(),
        Memory:      a.collectMemoryMetrics(),
        Disk:        a.collectDiskMetrics(),
        Network:     a.collectNetworkMetrics(),
    }
    
    return a.client.Metrics.SubmitComprehensive(context.Background(), metrics)
}
```

### Dashboard Implementation
```go
// Dashboard service
type DashboardService struct {
    client *nexmonyx.Client
}

func (d *DashboardService) GetOverview(ctx context.Context, orgID string) (*Overview, error) {
    // Get servers
    servers, _, err := d.client.Servers.List(ctx, &nexmonyx.ListOptions{
        Filters: map[string]string{"organization_id": orgID},
    })
    if err != nil {
        return nil, err
    }
    
    // Get active alerts
    alerts, _, err := d.client.Alerts.GetActiveAlerts(ctx, nil)
    if err != nil {
        return nil, err
    }
    
    // Get probe status
    probes, _, err := d.client.Monitoring.ListProbes(ctx, nil)
    if err != nil {
        return nil, err
    }
    
    return &Overview{
        TotalServers:   len(servers),
        OnlineServers:  countOnlineServers(servers),
        ActiveAlerts:   len(alerts),
        ProbesHealthy:  countHealthyProbes(probes),
    }, nil
}
```

### Monitoring Agent Implementation
```go
// Monitoring agent for probe execution
type MonitoringAgent struct {
    client *nexmonyx.Client
    config *MonitoringConfig
}

func (m *MonitoringAgent) Run() error {
    for {
        // Get probes to execute
        probes, err := m.client.Monitoring.GetProbesForAgent(context.Background())
        if err != nil {
            log.Printf("Failed to get probes: %v", err)
            time.Sleep(30 * time.Second)
            continue
        }
        
        // Execute probes
        for _, probe := range probes {
            go m.executeProbe(probe)
        }
        
        time.Sleep(time.Duration(m.config.PollingInterval) * time.Second)
    }
}

func (m *MonitoringAgent) executeProbe(probe *nexmonyx.Probe) {
    result := m.runProbeCheck(probe)
    
    // Submit result
    if err := m.client.Monitoring.SubmitProbeResult(context.Background(), result); err != nil {
        log.Printf("Failed to submit probe result: %v", err)
    }
}
```

## Testing

### Unit Testing
```go
func TestUserService(t *testing.T) {
    // Create test client
    client, err := nexmonyx.NewClient(&nexmonyx.Config{
        BaseURL: "http://localhost:8080",
        Auth: nexmonyx.AuthConfig{
            Token: "test-token",
        },
        Debug: true,
    })
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Test get current user
    user, err := client.Users.GetMe(ctx)
    assert.NoError(t, err)
    assert.NotEmpty(t, user.Email)
}
```

### Integration Testing
```go
func TestIntegration(t *testing.T) {
    if os.Getenv("INTEGRATION_TESTS") != "true" {
        t.Skip("Skipping integration tests")
    }
    
    client, err := nexmonyx.NewClient(&nexmonyx.Config{
        BaseURL: os.Getenv("NEXMONYX_API_URL"),
        Auth: nexmonyx.AuthConfig{
            Token: os.Getenv("NEXMONYX_TOKEN"),
        },
    })
    require.NoError(t, err)
    
    // Run integration tests
    testOrganizations(t, client)
    testServers(t, client)
    testMonitoring(t, client)
}
```

## License

This SDK is released under the same license as the Nexmonyx platform.

---

## LLM Usage Guidelines

When using this SDK as an LLM, follow these patterns:

1. **Always check authentication requirements** before suggesting code
2. **Use appropriate error handling** for production code
3. **Include context.Context** in all operations
4. **Consider pagination** for list operations
5. **Use specific error types** for better error handling
6. **Implement retries** for transient failures
7. **Cache client instances** for better performance
8. **Use structured logging** for debugging

This documentation provides comprehensive examples for both human developers and AI assistants to effectively use the Nexmonyx Go SDK.