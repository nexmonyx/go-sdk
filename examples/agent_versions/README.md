# Agent Versions API Example

This example demonstrates how to use the AgentVersionsService to manage agent versions via the Nexmonyx API.

## Features

The AgentVersionsService provides the following capabilities:

### Standard Methods
- `RegisterVersion(ctx, req)` - Register a new agent version (simple, no response)
- `CreateVersion(ctx, req)` - Create a new agent version and return the created version
- `GetVersion(ctx, version)` - Retrieve a specific agent version by version string
- `ListVersions(ctx, opts)` - List all agent versions with pagination
- `AddBinary(ctx, versionID, req)` - Add a binary for an existing agent version

### Admin Methods
- `AdminCreateVersion(ctx, version, notes)` - Create version using admin endpoint
- `AdminAddBinary(ctx, versionID, req)` - Add binary using admin endpoint

## API Endpoints

The service uses these API endpoints:

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/agent/versions` | Register/create agent version |
| GET | `/api/v1/agent/versions` | List agent versions |
| GET | `/api/v1/agent/versions/{version}` | Get specific version |
| POST | `/api/v1/agent/versions/{id}/binaries` | Add binary to version |
| POST | `/admin/agent-versions` | Admin create version |
| POST | `/admin/agent-versions/{id}/binaries` | Admin add binary |

## Authentication

The service supports multiple authentication methods:

### API Key Authentication
```go
client, err := nexmonyx.NewClient(&nexmonyx.Config{
    Auth: nexmonyx.AuthConfig{
        UnifiedAPIKey: "your-api-key",
    },
})
```

### Key/Secret Authentication
```go
client, err := nexmonyx.NewClient(&nexmonyx.Config{
    Auth: nexmonyx.AuthConfig{
        UnifiedAPIKey: "your-key",
        APIKeySecret:  "your-secret",
    },
})
```

## Usage Examples

### Basic Agent Version Registration

```go
// Create version request
req := &nexmonyx.AgentVersionRequest{
    Version:     "v1.3.4",
    Environment: "production",
    Platform:    "linux",
    Architectures: []string{"amd64", "arm64"},
    DownloadURLs: map[string]string{
        "amd64": "https://cdn.nexmonyx.com/agent/v1.3.4/nexmonyx-agent-amd64",
        "arm64": "https://cdn.nexmonyx.com/agent/v1.3.4/nexmonyx-agent-arm64",
    },
    ReleaseNotes:      "Bug fixes and improvements",
    MinimumAPIVersion: "1.0.0",
}

// Register the version
err := client.AgentVersions.RegisterVersion(ctx, req)
```

### Admin Workflow (as seen in register-version.sh)

```go
// Step 1: Create version
version, err := client.AgentVersions.AdminCreateVersion(ctx, "v1.3.4", "Release notes")
if err != nil {
    return err
}

// Step 2: Add binaries for each architecture
for _, arch := range []string{"amd64", "arm64"} {
    binary := &nexmonyx.AgentBinaryRequest{
        Platform:     "linux",
        Architecture: arch,
        DownloadURL:  fmt.Sprintf("https://cdn.nexmonyx.com/agent/v1.3.4/nexmonyx-agent-%s", arch),
        FileHash:     fmt.Sprintf("sha256:hash-for-%s", arch),
    }
    
    err = client.AgentVersions.AdminAddBinary(ctx, version.ID, binary)
    if err != nil {
        return err
    }
}
```

### Listing and Retrieving Versions

```go
// List versions with pagination
versions, meta, err := client.AgentVersions.ListVersions(ctx, &nexmonyx.ListOptions{
    Limit: 10,
    Page:  1,
})

// Get specific version
version, err := client.AgentVersions.GetVersion(ctx, "v1.3.4")
```

## Data Structures

### AgentVersionRequest
```go
type AgentVersionRequest struct {
    Version           string                 `json:"version"`
    Environment       string                 `json:"environment,omitempty"`
    Platform          string                 `json:"platform"`
    Architectures     []string               `json:"architectures,omitempty"`
    DownloadURLs      map[string]string      `json:"download_urls,omitempty"`
    UpdaterURLs       map[string]string      `json:"updater_urls,omitempty"`
    ReleaseNotes      string                 `json:"release_notes,omitempty"`
    MinimumAPIVersion string                 `json:"minimum_api_version,omitempty"`
    IsStable          *bool                  `json:"is_stable,omitempty"`
    IsPrerelease      *bool                  `json:"is_prerelease,omitempty"`
    Metadata          map[string]interface{} `json:"metadata,omitempty"`
}
```

### AgentBinaryRequest
```go
type AgentBinaryRequest struct {
    Platform     string `json:"platform"`
    Architecture string `json:"architecture"`
    DownloadURL  string `json:"download_url"`
    FileHash     string `json:"file_hash"`
}
```

## Running the Example

```bash
# Set your API key
export NEXMONYX_API_KEY="your-api-key-here"

# Run the example
go run main.go
```

## Error Handling

The service returns standard Nexmonyx API errors:

```go
err := client.AgentVersions.RegisterVersion(ctx, req)
if err != nil {
    switch e := err.(type) {
    case *nexmonyx.UnauthorizedError:
        log.Printf("Authentication failed: %s", e.Message)
    case *nexmonyx.ValidationError:
        log.Printf("Validation failed: %s", e.Message)
    case *nexmonyx.NotFoundError:
        log.Printf("Version not found: %s", e.Message)
    default:
        log.Printf("API error: %v", err)
    }
}
```

## Integration with Linux Agent

This service resolves the missing `/api/v1/agent/versions` endpoint that the Linux agent was trying to call for version registration. The Linux agent can now use:

```go
// In the Linux agent code
err := sdkClient.AgentVersions.RegisterVersion(ctx, &nexmonyx.AgentVersionRequest{
    Version:     agentVersion,
    Platform:    "linux",
    Environment: "production",
    // ... other fields
})
```