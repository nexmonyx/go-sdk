# Nexmonyx Go SDK

The official Go SDK for the Nexmonyx API - a comprehensive server monitoring and management platform.

## Features

- **Multiple Authentication Methods**: JWT tokens, API keys, server credentials, and monitoring keys
- **Complete API Coverage**: Full support for all Nexmonyx API endpoints
- **Enhanced Hardware Support**: Detailed hardware arrays including individual disk information
- **Type Safety**: Comprehensive Go types for all API models and responses
- **Error Handling**: Structured error types with detailed error information
- **Retry Logic**: Built-in retry mechanism with exponential backoff
- **Rate Limiting**: Automatic handling of rate limit responses
- **Pagination**: Easy-to-use pagination support for list operations
- **Context Support**: Full context.Context support for cancellation and timeouts
- **Debug Mode**: Optional request/response logging for debugging
- **Fluent Builder Pattern**: Easy-to-use builders for complex requests
- **WebSocket Support**: Real-time bidirectional communication for agent commands
- **Backward Compatibility**: Legacy API support maintained

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

### Monitoring Agent (MON_ Key Authentication)

```go
// Create monitoring agent client
client, err := nexmonyx.NewMonitoringAgentClient(&nexmonyx.Config{
    BaseURL: "https://api.nexmonyx.com",
    Auth: nexmonyx.AuthConfig{
        MonitoringKey: "MON_your_monitoring_key_here",
    },
    Debug: true, // Enable debug logging
})
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()

// Test authentication
if err := client.HealthCheck(ctx); err != nil {
    log.Fatal("Authentication failed:", err)
}

// Get assigned probes for a region
probes, err := client.Monitoring.GetAssignedProbes(ctx, "us-east-1")
if err != nil {
    log.Fatal(err)
}

// Execute probes and submit results
results := []nexmonyx.ProbeExecutionResult{
    {
        ProbeID:      probes[0].ProbeID,
        ProbeUUID:    probes[0].ProbeUUID,
        ExecutedAt:   time.Now(),
        Region:       "us-east-1",
        Status:       "success",
        ResponseTime: 150, // milliseconds
        StatusCode:   200,
    },
}

err = client.Monitoring.SubmitResults(ctx, results)
if err != nil {
    log.Fatal(err)
}

// Send heartbeat with node information
nodeInfo := nexmonyx.NodeInfo{
    AgentID:        "my-monitoring-agent",
    AgentVersion:   "1.0.0",
    Region:         "us-east-1",
    Status:         "healthy",
    ProbesAssigned: len(probes),
    SupportedTypes: []string{"http", "https", "tcp", "icmp"},
}

err = client.Monitoring.Heartbeat(ctx, nodeInfo)
if err != nil {
    log.Fatal(err)
}
```

For complete monitoring agent examples, see the [examples/monitoring/](./examples/monitoring/) directory.

## API Services

The SDK is organized into service clients for different API domains:

| Service | Description | Authentication | Example Endpoints |
|---------|-------------|----------------|-------------------|
| **Organizations** | Organization management and membership | JWT, API Key | List, Create, Invite, Members |
| **Servers** | Server registration, monitoring, and management | JWT, Server Credentials | List, Register, Metrics, Credentials |
| **Tags** | Tag management, namespaces, inheritance, and automation | JWT | CRUD, Namespaces, History, Bulk Ops, Rules |
| **Analytics** | AI insights, hardware predictions, fleet analytics, correlations | JWT | AI Analysis, Hardware Health, Fleet Overview, Dependencies |
| **ML** | Machine learning tag/group suggestions, model management, training | JWT | Tag Suggestions, Group Suggestions, Models, Training Jobs |
| **VMs** | Virtual machine lifecycle and resource management | JWT | Create, List, Control (Start/Stop/Restart), Delete |
| **Reporting** | Report generation and scheduling | JWT | Generate, List, Download, Schedule, Manage Schedules |
| **ServerGroups** | Server grouping and organization | JWT | Create, List, Add Servers, Get Members |
| **Search** | Comprehensive search across servers, tags, and resources | JWT | Search Servers, Search Tags, Tag Statistics |
| **Audit** | Audit log tracking and compliance reporting | JWT | List Logs, Export, Statistics, User History |
| **Users** | User profile and preference management | JWT | Profile, Preferences, Avatar |
| **Metrics** | Metrics submission and querying | Server Credentials, JWT | Submit, Query, History |
| **Monitoring** | Probes, regions, and monitoring infrastructure | JWT, Monitoring Key | Probes, Results, Regions |
| **Billing** | Subscription and billing management | JWT | Plans, Checkout, Usage |
| **BillingUsage** | Organization usage metrics for billing | JWT, API Key | Current Usage, History, Summary, Admin Overview |
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
   ├── Server Organization/Tagging → Tags service
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
   ├── Building a Dashboard → Users, Organizations, Servers, Monitoring, Tags
   ├── Managing Infrastructure → VMs, Servers, Organizations, Tags
   ├── Organizing Servers → Tags (namespaces, inheritance, bulk operations)
   ├── Handling Notifications → Alerts, EmailQueue
   ├── Public Website → Public, StatusPages, Distros
   └── Administrative Tool → Admin, Settings, Jobs
```

## Enhanced Hardware Support

The SDK provides comprehensive support for detailed hardware information, particularly enabling individual disk metrics collection:

### Key Features

- **Individual Disk Information**: Collect detailed information for each disk device
- **Comprehensive Hardware Arrays**: CPU, Memory, Network, and Disk details
- **Fluent Builder Pattern**: Easy-to-use construction methods
- **Backward Compatibility**: Existing legacy hardware fields continue to work
- **API Compatible**: JSON structure matches server expectations

### Quick Example

```go
// Create enhanced hardware details request
req := NewServerDetailsUpdateRequest().
    WithBasicInfo("server-01", "192.168.1.100", "production", "dc1", "web").
    WithDisks([]ServerDiskInfo{
        {
            Device:       "/dev/sda",
            DiskModel:    "Samsung SSD 980 PRO",
            SerialNumber: "S5P2NS0R123456",
            Size:         1000204886016,
            Type:         "NVMe",
            Vendor:       "Samsung",
        },
        {
            Device:       "/dev/sdb",
            DiskModel:    "WD Red Plus WD40EFPX",
            SerialNumber: "WD-WX12345678901",
            Size:         4000787030016,
            Type:         "HDD",
            Vendor:       "Western Digital",
        },
    })

// Send to API
server, err := client.Servers.UpdateDetails(ctx, "server-uuid", req)
```

### Helper Methods

```go
// Check if request has hardware details
if req.HasHardwareDetails() {
    fmt.Println("Enhanced hardware details present")
}

// Check specifically for disk information
if req.HasDisks() {
    fmt.Printf("Request contains %d disks\n", len(req.Hardware.Disks))
}
```

📖 **For complete hardware enhancement documentation, see [HARDWARE_ENHANCEMENT.md](HARDWARE_ENHANCEMENT.md)**

## WebSocket Support

The SDK provides comprehensive WebSocket support for real-time bidirectional communication with agents. This enables sending commands to agents and receiving responses with proper correlation tracking.

### Features

- **8 System Commands**: Complete support for all agent commands
  - `run_collection` - Trigger metrics collection
  - `force_collection` - Force comprehensive collection
  - `update_agent` - Update agent version
  - `check_updates` - Check for available updates
  - `restart_agent` - Restart agent service
  - `graceful_restart` - Gracefully restart agent
  - `agent_health` - Get agent health status
  - `system_status` - Get system status information

- **Connection Management**: Automatic connection, authentication, and heartbeat handling
- **Response Correlation**: Command responses matched via correlation IDs
- **Timeout Handling**: Configurable timeouts with context support
- **Event Handlers**: Connection, disconnection, and message event callbacks
- **Error Handling**: Comprehensive error handling for all scenarios

### Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/nexmonyx/go-sdk/v2"
)

func main() {
    // Create client with server credentials (required for WebSocket)
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

    // Initialize WebSocket service
    wsService, err := client.NewWebSocketService()
    if err != nil {
        log.Fatal(err)
    }

    // Set up event handlers
    wsService.OnConnect(func() {
        fmt.Println("WebSocket connected successfully")
    })

    wsService.OnDisconnect(func(err error) {
        fmt.Printf("WebSocket disconnected: %v\n", err)
    })

    // Connect to WebSocket
    if err := wsService.Connect(); err != nil {
        log.Fatal(err)
    }
    defer wsService.Disconnect()

    // Send commands to agents
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Check agent health
    response, err := wsService.AgentHealth(ctx, "target-server-uuid")
    if err != nil {
        log.Fatal(err)
    }

    if response.Success {
        fmt.Printf("Agent is healthy: %s\n", string(response.Data))
    } else {
        fmt.Printf("Agent health check failed: %s\n", response.Error)
    }
}
```

### System Commands

#### Metrics Collection

```go
// Run standard collection
collectionReq := &nexmonyx.CollectionRequest{
    CollectorTypes: []string{"cpu", "memory", "network"},
    Comprehensive:  false,
    Timeout:        30,
}
response, err := wsService.RunCollection(ctx, serverUUID, collectionReq)

// Force comprehensive collection
response, err := wsService.ForceCollection(ctx, serverUUID, &nexmonyx.CollectionRequest{
    CollectorTypes: []string{"all"},
    Timeout:        60,
})
```

#### Agent Management

```go
// Check for updates
response, err := wsService.CheckUpdates(ctx, serverUUID)

// Update agent to specific version
updateReq := &nexmonyx.UpdateRequest{
    Version:   "2.1.5",
    Force:     false,
    Immediate: false,
}
response, err := wsService.UpdateAgent(ctx, serverUUID, updateReq)

// Graceful restart with delay
restartReq := &nexmonyx.RestartRequest{
    Delay:  10,
    Reason: "Scheduled maintenance",
}
response, err := wsService.GracefulRestart(ctx, serverUUID, restartReq)
```

#### Status and Health Monitoring

```go
// Get agent health status
response, err := wsService.AgentHealth(ctx, serverUUID)
if response.Success {
    var health map[string]interface{}
    json.Unmarshal(response.Data, &health)
    fmt.Printf("Agent Status: %s, Version: %s\n", 
        health["status"], health["version"])
}

// Get system status
response, err := wsService.SystemStatus(ctx, serverUUID)
if response.Success {
    var status map[string]interface{}
    json.Unmarshal(response.Data, &status)
    fmt.Printf("Load Average: %v\n", status["load_average"])
}
```

### Advanced Usage

#### Custom Event Handling

```go
wsService.OnMessage(func(msg *nexmonyx.WSMessage) {
    switch msg.Type {
    case nexmonyx.WSTypeUpdateProgress:
        fmt.Printf("Update progress: %s\n", string(msg.Payload))
    case nexmonyx.WSTypeError:
        fmt.Printf("WebSocket error: %s\n", string(msg.Payload))
    default:
        fmt.Printf("Received message: type=%s\n", msg.Type)
    }
})
```

#### Batch Operations

```go
// Send commands to multiple servers
servers := []string{"server-1", "server-2", "server-3"}
results := make(map[string]*nexmonyx.WSCommandResponse)

for _, serverUUID := range servers {
    response, err := wsService.AgentHealth(ctx, serverUUID)
    if err != nil {
        fmt.Printf("Health check failed for %s: %v\n", serverUUID, err)
        continue
    }
    results[serverUUID] = response
}

// Process results
for serverUUID, response := range results {
    status := "FAILED"
    if response.Success {
        status = "OK"
    }
    fmt.Printf("Server %s: %s\n", serverUUID, status)
}
```

#### Configuration Options

```go
wsService, err := client.NewWebSocketService()
if err != nil {
    log.Fatal(err)
}

// Configure timeouts and retry behavior
wsService.SetTimeout(60 * time.Second)
wsService.SetReconnectDelay(10 * time.Second)
wsService.SetMaxReconnects(3)
```

### Error Handling

```go
response, err := wsService.AgentHealth(ctx, serverUUID)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "not connected"):
        fmt.Println("WebSocket is not connected")
    case strings.Contains(err.Error(), "timeout"):
        fmt.Println("Command timed out")
    case strings.Contains(err.Error(), "context deadline exceeded"):
        fmt.Println("Context timeout reached")
    default:
        fmt.Printf("Command failed: %v\n", err)
    }
    return
}

if !response.Success {
    fmt.Printf("Agent command failed: %s\n", response.Error)
    return
}

// Process successful response
fmt.Printf("Command executed in %.0fms\n", 
    response.Metadata["execution_time_ms"].(float64))
```

### Authentication Requirements

**Important**: WebSocket functionality requires server authentication credentials (`ServerUUID` and `ServerSecret`). Other authentication methods (JWT tokens, API keys) are not supported for WebSocket connections.

```go
// ✅ Correct - Server credentials
config := &nexmonyx.Config{
    Auth: nexmonyx.AuthConfig{
        ServerUUID:   "your-server-uuid",
        ServerSecret: "your-server-secret",
    },
}

// ❌ Incorrect - JWT token won't work for WebSocket
config := &nexmonyx.Config{
    Auth: nexmonyx.AuthConfig{
        Token: "jwt-token",
    },
}
```

📖 **For complete WebSocket examples, see [websocket_examples.go](websocket_examples.go)**

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

### Tags

The Tags service provides comprehensive tag management for organizing servers, including namespaces, inheritance, history tracking, bulk operations, and automated detection rules.

```go
// Basic Tag Management
// --------------------

// Create a tag
tag, err := client.Tags.Create(ctx, &nexmonyx.TagCreateRequest{
    Namespace: "env",
    Key:       "environment",
    Value:     "production",
    Color:     "#28a745",
})

// List tags with filtering
tags, meta, err := client.Tags.List(ctx, &nexmonyx.TagListOptions{
    Namespace: "env",
    Source:    "manual",      // "manual", "inherited", "auto_detected"
    Page:      1,
    Limit:     50,
})

// Get tags for a specific server
serverTags, err := client.Tags.GetServerTags(ctx, "server-uuid-123")

// Assign tags to a server
err = client.Tags.AssignTagsToServer(ctx, "server-uuid-123", &nexmonyx.ServerTagsAssignRequest{
    TagIDs: []uint{1, 2, 3},
})

// Remove a tag from a server
err = client.Tags.RemoveTagFromServer(ctx, "server-uuid-123", 1)


// Namespace Management
// --------------------

// Create hierarchical namespace
namespace, err := client.Tags.CreateNamespace(ctx, &nexmonyx.TagNamespaceCreateRequest{
    Name:        "infrastructure",
    ParentID:    nil,  // Top-level namespace
    Description: "Infrastructure organization tags",
    Icon:        "🏗️",
    Color:       "#007bff",
    IsSystem:    false,
})

// Create child namespace
childNamespace, err := client.Tags.CreateNamespace(ctx, &nexmonyx.TagNamespaceCreateRequest{
    Name:        "kubernetes",
    ParentID:    &namespace.ID,  // Child of infrastructure
    Description: "Kubernetes cluster tags",
    Icon:        "☸️",
    Color:       "#326ce5",
})

// List all namespaces
namespaces, err := client.Tags.ListNamespaces(ctx)

// Set namespace permissions
err = client.Tags.SetNamespacePermissions(ctx, namespace.ID, &nexmonyx.TagNamespacePermissionRequest{
    CanCreate: true,
    CanRead:   true,
    CanUpdate: false,
    CanDelete: false,
    CanApprove: false,
})


// Tag Inheritance
// ----------------

// Create inheritance rule
rule, err := client.Tags.CreateInheritanceRule(ctx, &nexmonyx.TagInheritanceRuleCreateRequest{
    SourceType:   "organization",  // "organization", "tag", "server_group"
    SourceID:     "org-uuid",
    TargetType:   "server",        // "server", "vm"
    Conditions:   json.RawMessage(`{"environment": "production"}`),
    Priority:     10,
    IsActive:     true,
})

// Set organization-level tags (inherited by all servers)
orgTag, err := client.Tags.SetOrganizationTag(ctx, &nexmonyx.OrganizationTagRequest{
    Namespace: "compliance",
    Key:       "data_classification",
    Value:     "confidential",
    AppliesTo: "all_servers",  // "all_servers", "specific_environments"
})

// List organization tags
orgTags, err := client.Tags.ListOrganizationTags(ctx, &nexmonyx.OrganizationTagListOptions{
    Namespace: "compliance",
    Page:      1,
    Limit:     50,
})

// Remove organization tag
err = client.Tags.RemoveOrganizationTag(ctx, orgTag.ID)

// Create server parent-child relationship for inheritance
relationship, err := client.Tags.CreateServerRelationship(ctx, &nexmonyx.ServerRelationshipRequest{
    ParentServerID:   "parent-uuid-100",
    ChildServerID:    "child-uuid-200",
    RelationshipType: "vm_host",  // "vm_host", "container_host", "cluster_member"
})

// List server relationships
relationships, err := client.Tags.ListServerRelationships(ctx, &nexmonyx.ServerRelationshipListOptions{
    ServerID:         "parent-uuid-100",
    RelationshipType: "vm_host",
})

// Delete relationship
err = client.Tags.DeleteServerRelationship(ctx, relationship.ID)


// Tag History and Audit
// ----------------------

// Get tag change history for a server
history, err := client.Tags.GetTagHistory(ctx, "server-uuid-123", &nexmonyx.TagHistoryOptions{
    StartDate: time.Now().AddDate(0, -1, 0),  // Last month
    EndDate:   time.Now(),
    Action:    "assigned",  // "assigned", "removed", "modified"
    Page:      1,
    Limit:     50,
})

// Process history entries
for _, entry := range history {
    fmt.Printf("Action: %s, User: %s, Time: %s\n",
        entry.Action, entry.UserEmail, entry.Timestamp)
    fmt.Printf("  Tag: %s:%s = %s\n",
        entry.Tag.Namespace, entry.Tag.Key, entry.Tag.Value)
}

// Get tag usage summary
summary, err := client.Tags.GetTagHistorySummary(ctx, &nexmonyx.TagHistorySummaryRequest{
    StartDate: time.Now().AddDate(0, -1, 0),
    EndDate:   time.Now(),
    GroupBy:   "namespace",  // "namespace", "key", "user"
})


// Bulk Operations
// ----------------

// Bulk create multiple tags
results, err := client.Tags.BulkCreateTags(ctx, &nexmonyx.BulkTagCreateRequest{
    Tags: []nexmonyx.TagCreateRequest{
        {Namespace: "env", Key: "environment", Value: "production"},
        {Namespace: "env", Key: "environment", Value: "staging"},
        {Namespace: "team", Key: "owner", Value: "backend"},
        {Namespace: "team", Key: "owner", Value: "frontend"},
    },
})

// Bulk assign tags to multiple servers
err = client.Tags.BulkAssignTags(ctx, &nexmonyx.BulkTagAssignRequest{
    ServerIDs: []string{"server-1", "server-2", "server-3"},
    TagIDs:    []uint{1, 2, 3},
})

// Assign servers to groups based on tag criteria
groupResults, err := client.Tags.AssignTagsToGroups(ctx, &nexmonyx.TagGroupAssignmentRequest{
    Groups: []nexmonyx.TagGroupCriteria{
        {
            Name:      "Production Web Servers",
            Criteria:  json.RawMessage(`{"environment": "production", "role": "web"}`),
            TagsToAdd: []uint{10, 11},
        },
        {
            Name:      "Development Databases",
            Criteria:  json.RawMessage(`{"environment": "dev", "role": "database"}`),
            TagsToAdd: []uint{20, 21},
        },
    },
})


// Tag Detection Rules
// --------------------

// List tag detection rules
rules, totalRules, err := client.Tags.ListTagDetectionRules(ctx, &nexmonyx.TagDetectionRuleListOptions{
    Enabled:   &enabled,  // Filter by enabled status
    Namespace: "auto",
    Page:      1,
    Limit:     50,
})

// Create default detection rules
result, err := client.Tags.CreateDefaultRules(ctx)
fmt.Printf("Created %d rules\n", result.RulesCreated)

// Evaluate rules for automatic tagging
evalResult, err := client.Tags.EvaluateRules(ctx, &nexmonyx.EvaluateRulesRequest{
    ServerIDs:      []string{"server-1", "server-2"},
    RuleIDs:        []uint{1, 2, 3},  // Optional: specific rules to evaluate
    AutoApply:      true,              // Automatically apply high-confidence matches
    MinConfidence:  0.8,               // Minimum confidence threshold
})

// Process evaluation results
for _, match := range evalResult.Matches {
    fmt.Printf("Server: %s, Rule: %s, Confidence: %.2f\n",
        match.ServerID, match.RuleName, match.Confidence)
    if match.Applied {
        fmt.Printf("  ✓ Tag automatically applied: %s:%s = %s\n",
            match.Namespace, match.TagKey, match.TagValue)
    }
}


// Error Handling
// --------------

tags, meta, err := client.Tags.List(ctx, &nexmonyx.TagListOptions{
    Namespace: "env",
})
if err != nil {
    switch {
    case errors.Is(err, nexmonyx.ErrUnauthorized):
        // Handle authentication error
        log.Fatal("Authentication required")
    case errors.Is(err, nexmonyx.ErrNotFound):
        // No tags found
        log.Println("No tags found in namespace 'env'")
    case errors.Is(err, nexmonyx.ErrRateLimitExceeded):
        // Rate limited
        log.Println("Rate limit exceeded, retry after:", err.RetryAfter)
    default:
        log.Printf("Error listing tags: %v", err)
    }
}


// Pagination
// ----------

// Iterate through all tags with pagination
page := 1
limit := 50
for {
    tags, meta, err := client.Tags.List(ctx, &nexmonyx.TagListOptions{
        Page:  page,
        Limit: limit,
    })
    if err != nil {
        return err
    }

    // Process tags
    for _, tag := range tags {
        fmt.Printf("Tag: %s:%s = %s (Source: %s)\n",
            tag.Namespace, tag.Key, tag.Value, tag.Source)
    }

    // Check if there are more pages
    if meta == nil || page >= meta.TotalPages {
        break
    }
    page++
}
```

### Analytics

The Analytics service provides AI-powered insights, hardware predictions, fleet statistics, and advanced correlation analysis using the `/v2/analytics` endpoints.

```go
// AI Analytics
// ------------

// Get available AI capabilities
capabilities, err := client.Analytics.GetCapabilities(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Anomaly Detection: %v\n", capabilities.AnomalyDetection)
fmt.Printf("Predictive Analytics: %v\n", capabilities.PredictiveAnalytics)

// Analyze metrics using AI
analysisReq := &nexmonyx.AIAnalysisRequest{
    ServerUUIDs:  []string{"server-uuid-123"},
    MetricTypes:  []string{"cpu", "memory", "disk"},
    AnalysisType: "anomaly",  // "anomaly", "prediction", "root_cause", "capacity"
    TimeRange: nexmonyx.TimeRange{
        Start: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
        End:   time.Now().Format(time.RFC3339),
    },
}

result, err := client.Analytics.AnalyzeMetrics(ctx, analysisReq)
if err != nil {
    log.Fatal(err)
}

// Process AI insights
for _, insight := range result.Insights {
    fmt.Printf("[%s] %s: %s (Confidence: %.2f)\n",
        insight.Severity, insight.Title, insight.Description, insight.Confidence)
}

// Check AI service status
status, err := client.Analytics.GetServiceStatus(ctx)
fmt.Printf("AI Service Status: %s (Uptime: %.2f%%)\n", status.Status, status.Uptime)


// Hardware Analytics
// ------------------

// Get hardware trends for a server
trends, err := client.Analytics.GetHardwareTrends(
    ctx,
    "server-uuid-123",
    time.Now().Add(-7*24*time.Hour).Format(time.RFC3339),  // Last 7 days
    time.Now().Format(time.RFC3339),
    "cpu,memory",  // Optional: specific metrics
)

fmt.Printf("CPU Average: %.2f%%, Growth: %.2f%%\n",
    trends.CPUTrend.Average, trends.CPUTrend.Growth)

// Get current hardware health score
health, err := client.Analytics.GetHardwareHealth(ctx, "server-uuid-123")
fmt.Printf("Overall Health Score: %d/100\n", health.OverallScore)
fmt.Printf("CPU Health: %d, Memory Health: %d\n",
    health.ComponentScores.CPU, health.ComponentScores.Memory)

// Review health issues
for _, issue := range health.Issues {
    fmt.Printf("[%s] %s: %s\n", issue.Severity, issue.Component, issue.Description)
}

// Get hardware failure predictions
predictions, err := client.Analytics.GetHardwarePredictions(ctx, "server-uuid-123", 30)
fmt.Printf("Failure Probability (30 days): %.2f%%\n", predictions.FailureProbability*100)

for _, component := range predictions.ComponentPredictions {
    if component.WarningLevel != "none" {
        fmt.Printf("⚠️  %s: %.2f%% failure risk (%s)\n",
            component.Component,
            component.FailureProbability*100,
            component.WarningLevel)
    }
}


// Fleet Analytics
// ----------------

// Get organization-wide fleet overview
overview, err := client.Analytics.GetFleetOverview(ctx)
fmt.Printf("Total Servers: %d (Active: %d, Inactive: %d)\n",
    overview.TotalServers, overview.ActiveServers, overview.InactiveServers)
fmt.Printf("Health Distribution - Healthy: %d, Warning: %d, Critical: %d\n",
    overview.HealthDistribution.Healthy,
    overview.HealthDistribution.Warning,
    overview.HealthDistribution.Critical)

// Get comprehensive dashboard data
dashboard, err := client.Analytics.GetOrganizationDashboard(ctx)

// Review recent alerts
fmt.Println("\nRecent Alerts:")
for _, alert := range dashboard.RecentAlerts {
    fmt.Printf("[%s] %s - %s\n", alert.Severity, alert.Title, alert.ServerName)
}

// Review trending metrics
fmt.Println("\nTrending Metrics:")
for _, metric := range dashboard.TrendingMetrics {
    trendSymbol := "→"
    if metric.Trend == "up" {
        trendSymbol = "↑"
    } else if metric.Trend == "down" {
        trendSymbol = "↓"
    }
    fmt.Printf("%s %s: %.2f (%s %.2f%%)\n",
        trendSymbol, metric.MetricType, metric.Value, trendSymbol, metric.Change)
}

// Check capacity forecasts
for _, forecast := range dashboard.CapacityForecasts {
    fmt.Printf("⚠️  %s: %d days until exhaustion (%.2f%% used)\n",
        forecast.ResourceType,
        forecast.DaysUntilExhaustion,
        forecast.CurrentUtilization)
}


// Advanced Analytics
// ------------------

// Analyze metric correlations
correlationReq := &nexmonyx.CorrelationAnalysisRequest{
    MetricTypes: []string{"cpu", "memory", "disk", "network"},
    TimeRange: nexmonyx.TimeRange{
        Start: time.Now().Add(-7*24*time.Hour).Format(time.RFC3339),
        End:   time.Now().Format(time.RFC3339),
    },
    Method: "pearson",  // "pearson", "spearman", "kendall"
}

correlations, err := client.Analytics.AnalyzeCorrelations(ctx, correlationReq)

fmt.Println("\nMetric Correlations:")
for _, corr := range correlations.Correlations {
    if math.Abs(corr.Coefficient) > 0.5 {  // Show significant correlations
        fmt.Printf("%s ↔ %s: %.3f (%s)\n",
            corr.Metric1, corr.Metric2, corr.Coefficient, corr.Strength)
    }
}

// Build infrastructure dependency graph
graph, err := client.Analytics.BuildDependencyGraph(ctx)

fmt.Printf("\nInfrastructure Dependency Graph:\n")
fmt.Printf("  Nodes: %d, Edges: %d\n", len(graph.Nodes), len(graph.Edges))

// Identify critical infrastructure
for _, path := range graph.CriticalPaths {
    fmt.Printf("  Critical Path: %s\n", strings.Join(path, " → "))
}

// Analyze dependencies
for _, edge := range graph.Edges {
    fmt.Printf("  %s %s %s\n", edge.From, edge.Type, edge.To)
}
```

### ML (Machine Learning)

The ML service provides AI-powered tag suggestions, server grouping recommendations, model management, and training job orchestration using `/v1/ml` and `/v1/groups/suggestions` endpoints.

```go
// Tag Suggestions
// ---------------

// Get ML-generated tag suggestions for a server
suggestions, err := client.ML.GetTagSuggestions(ctx, "server-uuid-123")
if err != nil {
    log.Fatal(err)
}

// Review suggestions
for _, suggestion := range suggestions {
    fmt.Printf("[%.0f%% confidence] %s=%s: %s\n",
        suggestion.Confidence*100,
        suggestion.TagKey,
        suggestion.TagValue,
        suggestion.Reason)
}

// Apply a specific tag suggestion
tagsApplied, err := client.ML.ApplyTagSuggestion(ctx, "server-uuid-123", "prediction-id-001")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Applied %d tags\n", tagsApplied)

// Reject a suggestion with feedback
err = client.ML.RejectTagSuggestion(ctx, "server-uuid-123", "prediction-id-002", "Server role is incorrect")
if err != nil {
    log.Fatal(err)
}


// Group Suggestions
// -----------------

// Get ML-generated server grouping suggestions
groupSuggestions, meta, err := client.ML.GetGroupSuggestions(ctx, &nexmonyx.PaginationOptions{
    Page:  1,
    Limit: 10,
})
if err != nil {
    log.Fatal(err)
}

// Review grouping suggestions
for _, suggestion := range groupSuggestions {
    fmt.Printf("\nGroup: %s (%.0f%% confidence)\n", suggestion.GroupName, suggestion.Confidence*100)
    fmt.Printf("Servers: %v\n", suggestion.ServerUUIDs)
    fmt.Printf("Reason: %s\n", suggestion.Reason)
    fmt.Printf("Criteria: %v\n", suggestion.Criteria)
    fmt.Printf("Benefit: %s\n", suggestion.EstimatedBenefit)
}

// Accept a group suggestion (creates the group)
acceptedGroup, err := client.ML.AcceptGroupSuggestion(ctx, 1)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created group ID: %d\n", *acceptedGroup.CreatedGroupID)

// Reject a group suggestion
err = client.ML.RejectGroupSuggestion(ctx, 2)
if err != nil {
    log.Fatal(err)
}


// Model Management
// ----------------

// List available ML models
models, meta, err := client.ML.ListModels(ctx, &nexmonyx.PaginationOptions{
    Page:  1,
    Limit: 20,
})
if err != nil {
    log.Fatal(err)
}

// View model details
for _, model := range models {
    fmt.Printf("\n%s (v%s) - %s\n", model.Name, model.Version, model.Status)
    fmt.Printf("  Type: %s\n", model.ModelType)
    fmt.Printf("  Enabled: %v\n", model.Enabled)
    if model.Accuracy > 0 {
        fmt.Printf("  Accuracy: %.2f%%\n", model.Accuracy*100)
        fmt.Printf("  F1 Score: %.2f\n", model.F1Score)
    }
}

// Get detailed model performance
performance, err := client.ML.GetModelPerformance(ctx, 1)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Accuracy: %.2f%%\n", performance.Accuracy*100)
fmt.Printf("Predictions: %d (Correct: %d, Incorrect: %d)\n",
    performance.PredictionsCount,
    performance.CorrectCount,
    performance.IncorrectCount)

// Toggle model enabled/disabled state
updatedModel, err := client.ML.ToggleModel(ctx, 1)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Model %s is now %s\n", updatedModel.Name,
    map[bool]string{true: "enabled", false: "disabled"}[updatedModel.Enabled])


// Model Training
// --------------

// Train a specific model type
job, err := client.ML.TrainModel(ctx, "tag_prediction", map[string]interface{}{
    "epochs":          100,
    "batch_size":      32,
    "learning_rate":   0.001,
    "validation_split": 0.2,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Training job %d started: %s\n", job.ID, job.Status)

// Trigger batch training for all models
jobs, err := client.ML.TriggerModelTraining(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Started %d training jobs\n", len(jobs))

// Monitor training jobs
trainingJobs, meta, err := client.ML.GetTrainingJobs(ctx,
    &nexmonyx.PaginationOptions{Page: 1, Limit: 10},
    "running") // Filter by status: pending, running, completed, failed
if err != nil {
    log.Fatal(err)
}

for _, job := range trainingJobs {
    fmt.Printf("Job %d: %s (%d%% complete)\n", job.ID, job.ModelType, job.Progress)
    if job.Status == "failed" {
        fmt.Printf("  Error: %s\n", job.ErrorMessage)
    }
    if job.Duration > 0 {
        fmt.Printf("  Duration: %d seconds\n", job.Duration)
    }
}

// Get aggregated performance across all models
aggregatedPerf, err := client.ML.GetAggregatedModelPerformance(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Overall ML System Performance:\n")
fmt.Printf("  Accuracy: %.2f%%\n", aggregatedPerf.Accuracy*100)
fmt.Printf("  Total Predictions: %d\n", aggregatedPerf.PredictionsCount)
fmt.Printf("  Average Confidence: %.2f%%\n", aggregatedPerf.AverageConfidence*100)
```

### VMs (Virtual Machines)

The VMs service provides virtual machine lifecycle management including creation, resource control, and monitoring using `/api/v1/vms` and `/api/v2/organizations/{orgId}/virtual-machines` endpoints.

```go
// VM Lifecycle Management
// -----------------------

// List all virtual machines
vms, meta, err := client.VMs.List(ctx, &nexmonyx.PaginationOptions{
    Page:  1,
    Limit: 20,
})
if err != nil {
    log.Fatal(err)
}

// Display VM information
for _, vm := range vms {
    fmt.Printf("\n%s (ID: %d) - %s\n", vm.Name, vm.ID, vm.Status)
    fmt.Printf("  Resources: %d CPU, %d MB RAM, %d GB Storage\n",
        vm.CPUCores, vm.MemoryMB, vm.StorageGB)
    fmt.Printf("  OS: %s %s\n", vm.OSType, vm.OSVersion)
    if vm.IPAddress != "" {
        fmt.Printf("  IP: %s\n", vm.IPAddress)
    }
}


// Create Virtual Machine
// ----------------------

config := &nexmonyx.VMConfiguration{
    Name:        "web-server-02",
    Description: "Development web server",

    // Resource allocation
    CPUCores:  4,
    MemoryMB:  8192,
    StorageGB: 100,

    // Operating system
    OSType:    "linux",
    OSVersion: "Ubuntu 22.04",

    // Optional: specify host server
    HostServerUUID: "server-uuid-123",

    // Tags and metadata
    Tags: []string{"environment:development", "role:webserver"},
    Metadata: map[string]interface{}{
        "project": "acme-app",
        "owner":   "devops-team",
    },
}

newVM, err := client.VMs.Create(ctx, config)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created VM: %s (ID: %d)\n", newVM.Name, newVM.ID)


// Get VM Details
// --------------

vm, err := client.VMs.Get(ctx, 1)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("VM: %s\n", vm.Name)
fmt.Printf("Status: %s\n", vm.Status)
fmt.Printf("Resources:\n")
fmt.Printf("  CPU: %d cores\n", vm.CPUCores)
fmt.Printf("  Memory: %d MB\n", vm.MemoryMB)
fmt.Printf("  Storage: %d GB\n", vm.StorageGB)

if vm.IPAddress != "" {
    fmt.Printf("Network:\n")
    fmt.Printf("  IP: %s\n", vm.IPAddress)
    fmt.Printf("  MAC: %s\n", vm.MACAddress)
}


// VM Control Operations
// ---------------------

orgID := uint(10)
vmID := uint(1)

// Start a stopped VM
startOp, err := client.VMs.Start(ctx, orgID, vmID)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Starting VM: %s (Progress: %d%%)\n", startOp.Message, startOp.Progress)

// Stop a running VM (graceful shutdown)
stopOp, err := client.VMs.Stop(ctx, orgID, vmID, false)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Stopping VM gracefully: %s\n", stopOp.Message)

// Force stop (immediate)
forceStopOp, err := client.VMs.Stop(ctx, orgID, vmID, true)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Force stopping VM: %s\n", forceStopOp.Message)

// Restart VM (graceful)
restartOp, err := client.VMs.Restart(ctx, orgID, vmID, false)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Restarting VM: %s (Operation ID: %d)\n", restartOp.Message, restartOp.ID)

// Monitor operation progress
fmt.Printf("Operation Status: %s\n", restartOp.Status)
fmt.Printf("Progress: %d%%\n", restartOp.Progress)


// Delete Virtual Machine
// -----------------------

err = client.VMs.Delete(ctx, orgID, vmID)
if err != nil {
    log.Fatal(err)
}
fmt.Println("VM deleted successfully")


// VM Operation Status
// -------------------

// VMOperation objects returned from control operations provide:
// - OperationType: "start", "stop", "restart", "delete", "create"
// - Status: "pending", "in_progress", "completed", "failed"
// - Progress: 0-100 percentage
// - Timestamps: CreatedAt, StartedAt, CompletedAt
// - Error details if operation failed

if startOp.Status == "failed" {
    fmt.Printf("Operation failed: %s\n", startOp.ErrorDetails)
} else {
    fmt.Printf("Operation %s: %d%% complete\n", startOp.Status, startOp.Progress)
}
```

### Reporting

The Reporting service provides comprehensive report generation and scheduling capabilities for usage, performance, compliance, and billing data.

#### Generate Reports

```go
// Generate a usage report for the last 30 days
config := &nexmonyx.ReportConfiguration{
    ReportType: "usage",
    Format:     "pdf",
    Name:       "Monthly Usage Report",
    Description: "Usage summary for December 2024",
    TimeRange: &nexmonyx.ReportTimeRange{
        StartDate: "2024-12-01T00:00:00Z",
        EndDate:   "2024-12-31T23:59:59Z",
    },
    Filters: &nexmonyx.ReportFilter{
        Locations:    []string{"us-east-1", "us-west-2"},
        Environments: []string{"production"},
        IncludeInactive: false,
    },
    Delivery: &nexmonyx.ReportDeliveryOptions{
        EmailRecipients: []string{"admin@example.com"},
        EmailSubject:    "Monthly Usage Report",
        RetentionDays:   30,
    },
}

report, err := client.Reporting.GenerateReport(ctx, config)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Report %d: %s (Status: %s)\n",
    report.ID, report.Name, report.Status)

// Generate performance report with specific servers
perfConfig := &nexmonyx.ReportConfiguration{
    ReportType: "performance",
    Format:     "csv",
    Name:       "Server Performance Analysis",
    TimeRange: &nexmonyx.ReportTimeRange{
        Preset: "last_7_days",
    },
    Filters: &nexmonyx.ReportFilter{
        ServerUUIDs: []string{"uuid-1", "uuid-2", "uuid-3"},
        MetricTypes: []string{"cpu", "memory", "disk_io"},
    },
    IncludeSections: []string{"summary", "trends", "anomalies"},
}

perfReport, err := client.Reporting.GenerateReport(ctx, perfConfig)
```

#### List and Retrieve Reports

```go
// List all completed reports
reports, meta, err := client.Reporting.ListReports(ctx,
    &nexmonyx.PaginationOptions{Page: 1, Limit: 20},
    "completed")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d completed reports\n", meta.TotalItems)
for _, report := range reports {
    fmt.Printf("- %s (%s) - %s [%d bytes]\n",
        report.Name, report.Format, report.Status, report.FileSize)
    if report.CompletedAt != nil {
        fmt.Printf("  Completed: %s\n", report.CompletedAt.Time)
    }
}

// Get specific report details
report, err := client.Reporting.GetReport(ctx, 123)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Report: %s\n", report.Name)
fmt.Printf("Type: %s, Format: %s\n", report.ReportType, report.Format)
fmt.Printf("Status: %s\n", report.Status)

if report.Status == "failed" {
    fmt.Printf("Error: %s\n", report.ErrorMessage)
}
```

#### Download Reports

```go
// Download completed report
reportID := uint(123)

// Check if report is ready
report, err := client.Reporting.GetReport(ctx, reportID)
if err != nil {
    log.Fatal(err)
}

if report.Status != "completed" {
    fmt.Printf("Report not ready yet (Status: %s)\n", report.Status)
    return
}

// Download the file
content, err := client.Reporting.DownloadReport(ctx, reportID)
if err != nil {
    log.Fatal(err)
}

// Save to file
filename := fmt.Sprintf("report_%d.%s", reportID, report.Format)
err = os.WriteFile(filename, content, 0644)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Report saved to %s (%d bytes)\n", filename, len(content))
```

#### Schedule Recurring Reports

```go
// Schedule weekly usage report
schedule := &nexmonyx.ReportSchedule{
    Name:        "Weekly Usage Report",
    Description: "Automated weekly usage summary",
    Configuration: &nexmonyx.ReportConfiguration{
        ReportType: "usage",
        Format:     "pdf",
        TimeRange: &nexmonyx.ReportTimeRange{
            Preset: "last_7_days",
        },
        Delivery: &nexmonyx.ReportDeliveryOptions{
            EmailRecipients: []string{"team@example.com"},
            EmailSubject:    "Weekly Usage Report",
            AutoDelete:      true,
            RetentionDays:   14,
        },
    },
    Schedule: "0 9 * * MON", // Every Monday at 9 AM
    Enabled:  true,
}

created, err := client.Reporting.ScheduleReport(ctx, schedule)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Scheduled report %d: %s\n", created.ID, created.Name)
if created.NextRunAt != nil {
    fmt.Printf("Next run: %s\n", created.NextRunAt.Time)
}

// Schedule daily performance report with custom cron
dailySchedule := &nexmonyx.ReportSchedule{
    Name:        "Daily Performance Summary",
    Description: "Performance metrics for all production servers",
    Configuration: &nexmonyx.ReportConfiguration{
        ReportType: "performance",
        Format:     "csv",
        Filters: &nexmonyx.ReportFilter{
            Environments: []string{"production"},
            MetricTypes:  []string{"cpu", "memory", "network"},
        },
        Delivery: &nexmonyx.ReportDeliveryOptions{
            WebhookURL: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
        },
    },
    Schedule: "0 6 * * *", // Every day at 6 AM
    Enabled:  true,
}

daily, err := client.Reporting.ScheduleReport(ctx, dailySchedule)
```

#### Manage Schedules

```go
// List all enabled schedules
enabled := true
schedules, meta, err := client.Reporting.ListSchedules(ctx,
    &nexmonyx.PaginationOptions{Page: 1, Limit: 50},
    &enabled)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Active schedules: %d\n", meta.TotalItems)
for _, sched := range schedules {
    fmt.Printf("- %s (%s)\n", sched.Name, sched.Schedule)
    if sched.LastRunAt != nil {
        fmt.Printf("  Last run: %s\n", sched.LastRunAt.Time)
    }
    if sched.NextRunAt != nil {
        fmt.Printf("  Next run: %s\n", sched.NextRunAt.Time)
    }
    if sched.LastReportID != nil {
        fmt.Printf("  Last report: %d\n", *sched.LastReportID)
    }
}

// List all schedules (enabled and disabled)
allSchedules, _, err := client.Reporting.ListSchedules(ctx, nil, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total schedules: %d\n", len(allSchedules))

// Delete a schedule
scheduleID := uint(456)
err = client.Reporting.DeleteSchedule(ctx, scheduleID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Schedule %d deleted\n", scheduleID)
```

#### Report Types and Formats

**Supported Report Types:**
- `usage` - Server usage metrics, resource consumption
- `performance` - CPU, memory, disk I/O, network performance
- `compliance` - Security compliance, audit logs
- `billing` - Cost analysis, usage-based billing

**Supported Formats:**
- `pdf` - Professional PDF documents
- `csv` - Comma-separated values for spreadsheets
- `json` - Structured JSON data
- `html` - HTML documents for web display

**Common Cron Expressions:**
- `0 0 * * *` - Daily at midnight
- `0 9 * * MON` - Every Monday at 9 AM
- `0 0 1 * *` - First day of every month at midnight
- `0 */6 * * *` - Every 6 hours
- `0 0 * * 0` - Every Sunday at midnight

### Server Groups

Server Groups enable logical organization of servers for batch operations, monitoring, and access control.

#### Create and List Groups

```go
// Create a server group
group, err := client.ServerGroups.CreateGroup(ctx,
	"Production Servers",
	"All production environment servers",
	[]string{"production", "critical"})
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Created group %d: %s\n", group.ID, group.Name)
fmt.Printf("Server count: %d\n", group.ServerCount)

// Create group for specific environment
devGroup, err := client.ServerGroups.CreateGroup(ctx,
	"Development Servers",
	"Development and testing environment",
	[]string{"development", "non-production"})

// List all groups
groups, meta, err := client.ServerGroups.ListGroups(ctx,
	&nexmonyx.PaginationOptions{Page: 1, Limit: 50},
	"",  // No name filter
	nil) // No tag filter
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Total groups: %d\n", meta.TotalItems)
for _, group := range groups {
	fmt.Printf("- %s (%d servers)\n", group.Name, group.ServerCount)
	if len(group.Tags) > 0 {
		fmt.Printf("  Tags: %v\n", group.Tags)
	}
}

// List groups with name filter
prodGroups, _, err := client.ServerGroups.ListGroups(ctx,
	&nexmonyx.PaginationOptions{Page: 1, Limit: 20},
	"prod", // Filter by name containing "prod"
	nil)

// List groups by tags
criticalGroups, _, err := client.ServerGroups.ListGroups(ctx,
	nil,
	"",
	[]string{"critical", "production"}) // Filter by tags
```

#### Add Servers to Groups

```go
// Add servers by UUIDs
groupID := uint(1)
serverUUIDs := []string{
	"server-uuid-1",
	"server-uuid-2",
	"server-uuid-3",
	"server-uuid-4",
}

count, err := client.ServerGroups.AddServersToGroup(ctx,
	groupID,
	nil,        // No server IDs
	serverUUIDs)
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Added %d servers to group\n", count)

// Add servers by IDs
serverIDs := []uint{101, 102, 103}
count, err = client.ServerGroups.AddServersToGroup(ctx,
	groupID,
	serverIDs,
	nil) // No UUIDs

// Add servers using both IDs and UUIDs
count, err = client.ServerGroups.AddServersToGroup(ctx,
	groupID,
	[]uint{104, 105},
	[]string{"uuid-6", "uuid-7"})
```

#### Get Group Members

```go
// Get all servers in a group
groupID := uint(1)
members, meta, err := client.ServerGroups.GetGroupServers(ctx,
	groupID,
	&nexmonyx.PaginationOptions{Page: 1, Limit: 100},
	"",  // No status filter
	nil) // No tag filter
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Group has %d servers:\n", meta.TotalItems)
for _, member := range members {
	fmt.Printf("- %s (%s) - Status: %s\n",
		member.ServerName,
		member.ServerUUID,
		member.ServerStatus)
	fmt.Printf("  Added: %s\n", member.AddedAt.Time)
}

// Get only online servers in group
onlineMembers, _, err := client.ServerGroups.GetGroupServers(ctx,
	groupID,
	nil,
	"online", // Filter by status
	nil)

fmt.Printf("Online servers: %d\n", len(onlineMembers))

// Get servers by tags within group
taggedMembers, _, err := client.ServerGroups.GetGroupServers(ctx,
	groupID,
	&nexmonyx.PaginationOptions{Page: 1, Limit: 50},
	"",                               // No status filter
	[]string{"database", "primary"}) // Filter by tags

// Pagination through large groups
page := 1
limit := 50
for {
	members, meta, err := client.ServerGroups.GetGroupServers(ctx,
		groupID,
		&nexmonyx.PaginationOptions{Page: page, Limit: limit},
		"", nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, member := range members {
		fmt.Printf("Server: %s\n", member.ServerName)
	}

	if page >= meta.TotalPages {
		break
	}
	page++
}
```

#### Common Use Cases

**Batch Operations by Group:**
```go
// Get all servers in production group
groupID := uint(1)
members, _, err := client.ServerGroups.GetGroupServers(ctx, groupID, nil, "", nil)
if err != nil {
	log.Fatal(err)
}

// Perform operation on each server
for _, member := range members {
	// Example: Get metrics for each server
	metrics, err := client.Metrics.GetLatest(ctx, member.ServerUUID)
	if err != nil {
		log.Printf("Error getting metrics for %s: %v\n", member.ServerName, err)
		continue
	}
	// Process metrics...
}
```

**Monitoring Setup:**
```go
// Create monitoring groups by role
roles := []string{"web-servers", "database-servers", "cache-servers"}
for _, role := range roles {
	group, err := client.ServerGroups.CreateGroup(ctx,
		fmt.Sprintf("Production %s", role),
		fmt.Sprintf("All production %s", role),
		[]string{"production", role})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created group: %s (ID: %d)\n", group.Name, group.ID)
}
```

**Access Control by Group:**
```go
// List all groups to show available server collections
groups, _, err := client.ServerGroups.ListGroups(ctx, nil, "", nil)
if err != nil {
	log.Fatal(err)
}

fmt.Println("Available Server Groups:")
for _, group := range groups {
	fmt.Printf("- %s: %d servers", group.Name, group.ServerCount)
	if len(group.Tags) > 0 {
		fmt.Printf(" [%s]", strings.Join(group.Tags, ", "))
	}
	fmt.Println()
}
```

### Search

The Search service provides comprehensive search capabilities across servers, tags, and other resources with advanced filtering, relevance scoring, and statistics.

#### Search Servers

```go
// Basic server search
results, meta, err := client.Search.SearchServers(ctx,
	"web",                                   // Search query
	&nexmonyx.PaginationOptions{Page: 1, Limit: 20},
	nil)                                     // No filters
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Found %d servers:\n", meta.TotalItems)
for _, result := range results {
	fmt.Printf("- %s (%s) - Score: %.2f\n",
		result.ServerName,
		result.ServerUUID,
		result.RelevanceScore)
	fmt.Printf("  Matched fields: %s\n", strings.Join(result.MatchedFields, ", "))
	if len(result.Tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(result.Tags, ", "))
	}
}

// Advanced server search with filters
results, meta, err := client.Search.SearchServers(ctx,
	"database",
	&nexmonyx.PaginationOptions{Page: 1, Limit: 50},
	map[string]interface{}{
		"location":       "us-east-1",
		"environment":    "production",
		"status":         "online",
		"classification": "critical",
	})
if err != nil {
	log.Fatal(err)
}

// Search by IP address or UUID
results, meta, err := client.Search.SearchServers(ctx,
	"10.0.1.10",  // Searches across name, hostname, IP addresses, UUID
	nil,
	nil)
```

#### Search Tags

```go
// Basic tag search
tags, meta, err := client.Search.SearchTags(ctx,
	"prod",                                  // Search query
	&nexmonyx.PaginationOptions{Page: 1, Limit: 50},
	nil)                                     // No filters
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Found %d tags:\n", meta.TotalItems)
for _, tag := range tags {
	fmt.Printf("- %s (%s) - Score: %.2f\n",
		tag.TagName,
		tag.TagType,
		tag.RelevanceScore)
	fmt.Printf("  Usage: %d resources, %d servers\n",
		tag.UsageCount,
		tag.ServerCount)
	if tag.Description != "" {
		fmt.Printf("  Description: %s\n", tag.Description)
	}
}

// Filter tags by type and scope
tags, meta, err := client.Search.SearchTags(ctx,
	"system",
	nil,
	map[string]interface{}{
		"tag_type": "auto",           // manual, auto, system
		"scope":    "server",         // organization, user, server
	})
if err != nil {
	log.Fatal(err)
}

// Find unused tags (usage_count = 0)
tags, meta, err := client.Search.SearchTags(ctx,
	"",  // Empty query to get all
	&nexmonyx.PaginationOptions{Page: 1, Limit: 100},
	nil)

unusedTags := []nexmonyx.TagSearchResult{}
for _, tag := range tags {
	if tag.UsageCount == 0 {
		unusedTags = append(unusedTags, tag)
	}
}
fmt.Printf("Found %d unused tags\n", len(unusedTags))
```

#### Tag Statistics

```go
// Get comprehensive tag usage statistics
stats, err := client.Search.GetTagStatistics(ctx, "", "")
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Tag Statistics:\n")
fmt.Printf("Total Tags: %d\n", stats.TotalTags)
fmt.Printf("- Manual: %d\n", stats.ManualTags)
fmt.Printf("- Auto: %d\n", stats.AutoTags)
fmt.Printf("- System: %d\n", stats.SystemTags)
fmt.Printf("Average tags per server: %.2f\n", stats.AveragePerServer)
fmt.Printf("Unused tags: %d\n", stats.UnusedTags)

// Tags by scope
fmt.Println("\nTags by Scope:")
for scope, count := range stats.TagsByScope {
	fmt.Printf("- %s: %d tags\n", scope, count)
}

// Most used tags
fmt.Println("\nMost Used Tags:")
for i, tag := range stats.MostUsedTags {
	fmt.Printf("%d. %s (%s) - %d uses across %d servers\n",
		i+1,
		tag.TagName,
		tag.TagType,
		tag.UsageCount,
		tag.ServerCount)
}

// Recently created tags
fmt.Println("\nRecently Created Tags:")
for _, tag := range stats.RecentlyCreated {
	fmt.Printf("- %s (created %s)\n",
		tag.TagName,
		tag.CreatedAt.Format("2006-01-02 15:04:05"))
}

// Filter statistics by tag type and scope
stats, err := client.Search.GetTagStatistics(ctx, "manual", "organization")
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Organization-level manual tags: %d\n", stats.TotalTags)
```

#### Common Search Use Cases

```go
// Find all production servers in a specific location
prodServers, _, err := client.Search.SearchServers(ctx,
	"",  // Empty query to get all
	&nexmonyx.PaginationOptions{Page: 1, Limit: 100},
	map[string]interface{}{
		"environment": "production",
		"location":    "us-west-2",
		"status":      "online",
	})

// Find servers with specific tags
webServers, _, err := client.Search.SearchServers(ctx,
	"web",  // Searches in tags field
	nil,
	map[string]interface{}{
		"environment": "production",
	})

// Identify underutilized tags
stats, err := client.Search.GetTagStatistics(ctx, "", "")
if err != nil {
	log.Fatal(err)
}

fmt.Println("Tags with low usage:")
for _, tag := range stats.MostUsedTags {
	if tag.ServerCount < 5 {
		fmt.Printf("- %s: only %d servers\n", tag.TagName, tag.ServerCount)
	}
}

// Find duplicate or similar tag names
allTags, _, err := client.Search.SearchTags(ctx,
	"prod",  // Will find "prod", "production", "prod-critical", etc.
	&nexmonyx.PaginationOptions{Page: 1, Limit: 100},
	nil)

tagNames := make(map[string]int)
for _, tag := range allTags {
	tagNames[tag.TagName]++
}
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

### BillingUsage

The BillingUsage service provides organization usage metrics for billing and analytics purposes. It supports both self-service endpoints (for users to view their own organization's usage) and admin endpoints (for platform administrators to view all organizations).

#### Self-Service Usage Endpoints (JWT Token)

```go
// Get current usage metrics for your organization
currentUsage, err := client.BillingUsage.GetMyCurrentUsage(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Active Agents: %d\n", currentUsage.ActiveAgentCount)
fmt.Printf("Storage Used: %.2f GB\n", currentUsage.StorageUsedGB)
fmt.Printf("Retention Days: %d\n", currentUsage.RetentionDays)

// Get usage history over a time period
startDate := time.Now().AddDate(0, 0, -30) // 30 days ago
endDate := time.Now()
history, err := client.BillingUsage.GetMyUsageHistory(ctx, startDate, endDate, "daily")
if err != nil {
    log.Fatal(err)
}

for _, record := range history {
    fmt.Printf("Date: %v, Agents: %d, Storage: %.2f GB\n",
        record.CollectedAt, record.ActiveAgentCount, record.StorageUsedGB)
}

// Get aggregated usage summary
summary, err := client.BillingUsage.GetMyUsageSummary(ctx, startDate, endDate)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Average Agents: %.1f\n", summary.AverageAgentCount)
fmt.Printf("Max Agents: %d\n", summary.MaxAgentCount)
fmt.Printf("Average Storage: %.2f GB\n", summary.AverageStorageGB)
fmt.Printf("Max Storage: %.2f GB\n", summary.MaxStorageGB)
if summary.BillingRecommendation != "" {
    fmt.Printf("Recommendation: %s\n", summary.BillingRecommendation)
}
```

#### Admin Usage Endpoints (Admin JWT Token or API Key)

```go
// Get current usage for a specific organization
orgID := uint(100)
orgUsage, err := client.BillingUsage.GetOrgCurrentUsage(ctx, orgID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Org %d: %d agents, %.2f GB storage\n",
    orgUsage.OrganizationID, orgUsage.ActiveAgentCount, orgUsage.StorageUsedGB)

// Get usage history for a specific organization
startDate := time.Now().AddDate(0, -3, 0) // 3 months ago
endDate := time.Now()
orgHistory, err := client.BillingUsage.GetOrgUsageHistory(ctx, orgID, startDate, endDate, "monthly")
if err != nil {
    log.Fatal(err)
}

// Get usage summary for a specific organization
orgSummary, err := client.BillingUsage.GetOrgUsageSummary(ctx, orgID, startDate, endDate)
if err != nil {
    log.Fatal(err)
}

// Get usage overview for all organizations (paginated)
opts := &nexmonyx.ListOptions{
    Page:  1,
    Limit: 50,
}

overview, meta, err := client.BillingUsage.GetAllUsageOverview(ctx, opts)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total Organizations: %d\n", overview.TotalOrganizations)
fmt.Printf("Total Active Agents: %d\n", overview.TotalActiveAgents)
fmt.Printf("Total Storage: %.2f GB\n", overview.TotalStorageGB)

for _, org := range overview.Organizations {
    fmt.Printf("  Org %d: %d agents, %.2f GB\n",
        org.OrganizationID, org.ActiveAgentCount, org.StorageUsedGB)
}

fmt.Printf("Page %d of %d (Total: %d organizations)\n",
    meta.Page, meta.TotalPages, meta.TotalItems)
```

#### Usage Intervals

The `GetMyUsageHistory` and `GetOrgUsageHistory` methods support three aggregation intervals:

- `"hourly"` - Hourly usage data points
- `"daily"` - Daily aggregated usage (default)
- `"monthly"` - Monthly aggregated usage

#### Authentication Requirements

- **Self-Service Endpoints**: Require JWT token authentication (user must be authenticated)
- **Admin Endpoints**: Require either:
  - Admin JWT token (user with admin privileges)
  - API Key authentication with admin scope

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

Monitoring agent keys are specialized API keys used by monitoring agents to authenticate with the API and submit probe results. Keys can be created for two types of agents:

- **Public Agents**: Nexmonyx-managed agents that can only execute public probes and require a region code
- **Private Agents**: Customer-managed agents that can execute both public and private probes

```go
// Create a private monitoring agent key (customer-managed)
privateKeyReq := nexmonyx.NewPrivateAgentKeyRequest(
    "My Private Monitoring Agent",
    "private-agent-1",
    "NYC3", // Region code is optional for private agents
)

privateKey, err := client.MonitoringAgentKeys.Create(ctx, "114", privateKeyReq)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Private Agent Key: %s\n", privateKey.FullToken)
fmt.Printf("Allowed Scopes: %v\n", privateKey.AllowedProbeScopes) // ["public", "private"]

// Create a public monitoring agent key (Nexmonyx-managed)
publicKeyReq := nexmonyx.NewPublicAgentKeyRequest(
    "NYC3 Public Monitoring Agent",
    "public-agent-nyc3",
    "NYC3", // Region code is REQUIRED for public agents
)

publicKey, err := client.MonitoringAgentKeys.Create(ctx, "114", publicKeyReq)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Public Agent Key: %s\n", publicKey.FullToken)
fmt.Printf("Allowed Scopes: %v\n", publicKey.AllowedProbeScopes) // ["public"]

// Custom key creation with full control
customKeyReq := &nexmonyx.CreateMonitoringAgentKeyRequest{
    Description:        "Custom monitoring agent",
    NamespaceName:      "custom-agent-1",
    AgentType:          "private",
    RegionCode:         "NYC3",
    AllowedProbeScopes: []string{"public", "private"},
    Capabilities:       `["probe:read","probe:write","node:register"]`,
}

customKey, err := client.MonitoringAgentKeys.Create(ctx, "114", customKeyReq)
if err != nil {
    log.Fatal(err)
}

// Admin operations - create monitoring agent keys for any organization
adminKeyReq := &nexmonyx.CreateMonitoringAgentKeyRequest{
    OrganizationID:     114,
    Description:        "Admin-created monitoring agent",
    NamespaceName:      "admin-agent-1",
    AgentType:          "private",
    RegionCode:         "NYC3",
    AllowedProbeScopes: []string{"public", "private"},
}

agentKeyResp, err := client.MonitoringAgentKeys.CreateAdmin(ctx, adminKeyReq)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created monitoring agent key: %s\n", agentKeyResp.FullToken)

// List organization's monitoring agent keys
keys, pagination, err := client.MonitoringAgentKeys.List(ctx, "114", &nexmonyx.ListMonitoringAgentKeysOptions{
    Page:      1,
    Limit:     50,
    Namespace: "production",
})

for _, key := range keys {
    fmt.Printf("Key: %s (%s), Type: %s, Region: %s, Status: %s\n", 
        key.KeyID, key.Description, key.AgentType, key.RegionCode, key.Status)
    
    if key.IsPublic() {
        fmt.Println("  This is a public agent key")
    } else if key.IsPrivate() {
        fmt.Println("  This is a private agent key")
    }
}

// Revoke a monitoring agent key
err = client.MonitoringAgentKeys.Revoke(ctx, "114", "mag_abc123")
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