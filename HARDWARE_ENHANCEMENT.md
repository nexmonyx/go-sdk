# Nexmonyx Go SDK Hardware Enhancement

This document describes the enhanced hardware support in the Nexmonyx Go SDK, which enables detailed hardware information collection, particularly for individual disk metrics.

## Overview

The SDK has been enhanced to support detailed hardware arrays in server update requests, addressing the limitation where only summary hardware fields were previously supported. This enhancement is **backward compatible** and enables individual disk metrics collection.

## Key Features

### 1. Enhanced Hardware Details Support

The `ServerDetailsUpdateRequest` now supports an optional `Hardware` field containing detailed hardware arrays:

```go
type HardwareDetails struct {
    CPU     []ServerCPUInfo              `json:"cpu,omitempty"`
    Memory  *ServerMemoryInfo            `json:"memory,omitempty"`
    Network []ServerNetworkInterfaceInfo `json:"network,omitempty"`
    Disks   []ServerDiskInfo             `json:"disks,omitempty"`  // KEY ADDITION
}
```

### 2. Individual Disk Information

The `ServerDiskInfo` structure provides comprehensive disk details:

```go
type ServerDiskInfo struct {
    Device       string `json:"device,omitempty"`        // e.g., "/dev/sda"
    DiskModel    string `json:"disk_model,omitempty"`    // e.g., "Samsung SSD 980 PRO"
    SerialNumber string `json:"serial_number,omitempty"` // e.g., "S5P2NS0R123456"
    Size         int64  `json:"size,omitempty"`          // Size in bytes
    Type         string `json:"type,omitempty"`          // e.g., "SSD", "HDD", "NVMe"
    Vendor       string `json:"vendor,omitempty"`        // e.g., "Samsung"
}
```

### 3. Fluent Builder Pattern

The SDK provides a fluent builder pattern for easy construction:

```go
req := NewServerDetailsUpdateRequest().
    WithBasicInfo("server-01", "192.168.1.100", "production", "dc1", "web").
    WithSystemInfo("linux", "Ubuntu 22.04", "x86_64", "SN123", "MAC123").
    WithDisks([]ServerDiskInfo{
        {
            Device:       "/dev/sda",
            DiskModel:    "Samsung SSD 980 PRO",
            SerialNumber: "S5P2NS0R123456",
            Size:         1000204886016,
            Type:         "NVMe",
            Vendor:       "Samsung",
        },
    })
```

### 4. Backward Compatibility

Existing code continues to work unchanged. Legacy hardware fields are still supported:

```go
req := NewServerDetailsUpdateRequest().
    WithLegacyHardware("Intel Xeon", 2, 16, 32768, 1000000)  // Still works!
```

## API Compatibility

The enhanced structures generate JSON that matches the API server's expectations:

```json
{
  "hostname": "server-01",
  "main_ip": "192.168.1.100",
  "hardware": {
    "disks": [
      {
        "device": "/dev/sda",
        "disk_model": "Samsung SSD 980 PRO",
        "serial_number": "S5P2NS0R123456",
        "size": 1000204886016,
        "type": "NVMe",
        "vendor": "Samsung"
      }
    ]
  }
}
```

## Usage Examples

### Basic Disk Information Update

```go
// Create client
client, err := NewClient(&Config{
    BaseURL: "https://api.nexmonyx.com",
    Auth: AuthConfig{
        ServerUUID:   "your-server-uuid",
        ServerSecret: "your-server-secret",
    },
})

// Create request with disk information
req := NewServerDetailsUpdateRequest().
    WithDisks([]ServerDiskInfo{
        {
            Device:       "/dev/nvme0n1",
            DiskModel:    "WD Black SN750",
            SerialNumber: "WD-WX12345678901",
            Size:         512110190592,
            Type:         "NVMe",
            Vendor:       "Western Digital",
        },
    })

// Send update
server, err := client.Servers.UpdateDetails(ctx, "server-uuid", req)
```

### Comprehensive Hardware Update

```go
req := NewServerDetailsUpdateRequest().
    WithCPUs([]ServerCPUInfo{
        {
            Manufacturer:  "Intel",
            ModelName:     "Intel Xeon E5-2680 v4",
            PhysicalCores: 14,
            LogicalCores:  28,
            Architecture:  "x86_64",
        },
    }).
    WithMemory(&ServerMemoryInfo{
        TotalSize:    68719476736, // 64GB
        MemoryType:   "DDR4",
        Speed:        2400,
        ECCSupported: true,
    }).
    WithNetworkInterfaces([]ServerNetworkInterfaceInfo{
        {
            Name:         "eth0",
            HardwareAddr: "00:50:56:c0:00:08",
            SpeedMbps:    1000,
            IsUp:         true,
        },
    }).
    WithDisks([]ServerDiskInfo{
        {
            Device:    "/dev/sda",
            DiskModel: "Samsung SSD 980 PRO",
            Size:      1000204886016,
            Type:      "NVMe",
            Vendor:    "Samsung",
        },
    })
```

## Debug Logging

Enable debug logging to see detailed hardware information in the logs:

```go
client, err := NewClient(&Config{
    // ... other config
    Debug: true,  // Enable debug logging
})
```

With debug enabled, you'll see logs like:

```
[DEBUG] UpdateDetails: Enhanced hardware details present
[DEBUG]   Disks: 2
[DEBUG]   Disk[0]: /dev/sda (Samsung SSD 980 PRO) NVMe - 1000204886016 bytes
[DEBUG]   Disk[1]: /dev/sdb (WD Red Plus) HDD - 4000787030016 bytes
```

## Helper Methods

The SDK provides convenient helper methods:

```go
// Check if request has hardware details
if req.HasHardwareDetails() {
    fmt.Println("Request contains enhanced hardware information")
}

// Check specifically for disk information
if req.HasDisks() {
    fmt.Printf("Request contains %d disks\n", len(req.Hardware.Disks))
}
```

## Migration Guide

### From Legacy to Enhanced

**Before** (legacy approach):
```go
req := &ServerDetailsUpdateRequest{
    CPUModel:     "Intel Xeon",
    CPUCores:     16,
    MemoryTotal:  67108864,
    StorageTotal: 2000000000000,
}
```

**After** (enhanced approach):
```go
req := NewServerDetailsUpdateRequest().
    WithLegacyHardware("Intel Xeon", 2, 16, 67108864, 2000000000000).  // Still works
    WithDisks([]ServerDiskInfo{  // Now you can also add individual disks
        {Device: "/dev/sda", DiskModel: "Samsung SSD", Size: 1000000000000, Type: "SSD"},
        {Device: "/dev/sdb", DiskModel: "WD HDD", Size: 1000000000000, Type: "HDD"},
    })
```

## Error Handling

```go
server, err := client.Servers.UpdateDetails(ctx, serverUUID, req)
if err != nil {
    // Check for specific error types
    if strings.Contains(err.Error(), "hardware") {
        fmt.Printf("Hardware-related error: %v\n", err)
    }
    return err
}
```

## Performance Considerations

- The enhanced hardware structures add minimal JSON marshaling overhead
- Debug logging adds some performance cost but provides valuable troubleshooting information
- Large hardware arrays (100+ disks) should be sent in smaller batches if experiencing timeouts

## Compatibility Matrix

| SDK Version | API Compatibility | Legacy Fields | Enhanced Hardware |
|-------------|------------------|---------------|-------------------|
| < v2.1.0    | ✅ Basic          | ✅ Yes         | ❌ No             |
| >= v2.1.0   | ✅ Full           | ✅ Yes         | ✅ Yes            |

## Testing

The SDK includes comprehensive tests for the enhanced functionality:

```bash
# Run hardware-specific tests
go test -v -run TestHardware

# Run API compatibility tests
go test -v -run TestAPICompatibility

# Run all tests
go test -v
```

## Troubleshooting

### Common Issues

1. **"No disk information to send"**
   - Ensure you're using `WithDisks()` with non-empty array
   - Check that `HasDisks()` returns true

2. **"Hardware details not present"**
   - Hardware field is only created when you use `WithCPUs()`, `WithMemory()`, `WithNetworkInterfaces()`, or `WithDisks()`
   - Legacy fields don't create the hardware structure

3. **JSON marshaling errors**
   - Verify all required fields are populated
   - Check that numeric fields use correct types (int64 for size, int for counts)

### Debug Tips

1. Enable debug logging: `client.Config.Debug = true`
2. Use helper methods: `req.HasHardwareDetails()`, `req.HasDisks()`
3. Inspect JSON output:
   ```go
   jsonData, _ := json.MarshalIndent(req, "", "  ")
   fmt.Println(string(jsonData))
   ```

## Summary

The enhanced Nexmonyx Go SDK now supports:

✅ **Individual disk information collection**  
✅ **Detailed CPU, Memory, and Network hardware arrays**  
✅ **Fluent builder pattern for easy construction**  
✅ **Full backward compatibility with existing code**  
✅ **Comprehensive debug logging**  
✅ **API-compatible JSON structure**  

This enhancement removes the blocking issue for individual disk metrics collection while maintaining full compatibility with existing implementations.