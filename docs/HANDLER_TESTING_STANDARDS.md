# Handler Testing Standards - Nexmonyx Go SDK

**Document Status**: Definitive Testing Guide for SDK Handlers
**Last Updated**: 2025-10-17
**Scope**: HTTP mock server handlers used in SDK unit tests

---

## 📋 Table of Contents

- [Overview](#overview)
- [What Are Handlers?](#what-are-handlers)
- [Handler Anatomy](#handler-anatomy)
- [Testing Patterns](#testing-patterns)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)
- [Advanced Techniques](#advanced-techniques)
- [Checklist](#checklist)

---

## Overview

This document defines standards for testing HTTP handlers in the Nexmonyx Go SDK. Handlers are mock HTTP servers that simulate the Nexmonyx API during unit testing.

### Purpose
- Ensure consistent, reliable unit tests
- Define testing patterns for all services
- Provide templates for new handler tests
- Establish quality standards

### Scope
- HTTP mock servers created with `httptest`
- Service method testing
- Request/response validation
- Error handling

---

## What Are Handlers?

### Definition
**Handlers are mock HTTP servers that simulate the Nexmonyx API** for testing SDK service methods without making real API calls.

### Architecture
```
Test Function
    ↓
Mock HTTP Server (Handler)
    ↓
Service Method Call
    ↓
Handler Response
    ↓
Assertion
```

### Key Characteristics
- Created with `httptest.NewServer(http.HandlerFunc(...))`
- Simulate API endpoints and responses
- Used exclusively in unit tests
- Defined inline in test files
- Validate request structure and return mock responses

### Location
- **102+ test files** throughout the SDK
- **In test files only** (`*_test.go`)
- **No separate handlers package** exists
- Handlers are inline with test implementations

---

## Handler Anatomy

### Basic Structure

Every handler follows a three-phase pattern:

#### Phase 1: Request Validation
```go
// Validate HTTP method
if r.Method != http.MethodGet {
    http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    return
}

// Validate path
if r.URL.Path != "/api/v1/organizations" {
    http.Error(w, "not found", http.StatusNotFound)
    return
}

// Validate authentication
authHeader := r.Header.Get("Authorization")
if authHeader == "" {
    http.Error(w, "unauthorized", http.StatusUnauthorized)
    return
}
```

#### Phase 2: Request Parsing
```go
// Decode request body (for POST/PUT)
var req CreateOrganizationRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, "bad request", http.StatusBadRequest)
    return
}

// Validate request fields
if req.Name == "" {
    http.Error(w, "name required", http.StatusBadRequest)
    return
}
```

#### Phase 3: Response Construction
```go
// Build response
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(map[string]interface{}{
    "data": map[string]interface{}{
        "id":   1,
        "name": req.Name,
    },
})
```

---

## Testing Patterns

### Pattern 1: Simple Handler (Single Scenario)

Use for simple tests with one request/response:

```go
func TestOrganizationsService_GetBasic(t *testing.T) {
    // Setup mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validate request
        assert.Equal(t, http.MethodGet, r.Method)
        assert.Equal(t, "/api/v1/organizations/1", r.URL.Path)

        // Return response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "data": map[string]interface{}{
                "id":   1,
                "name": "Test Org",
            },
        })
    }))
    defer server.Close()

    // Create client
    client, err := nexmonyx.NewClient(&nexmonyx.Config{
        BaseURL:   server.URL,
        Auth:      nexmonyx.AuthConfig{Token: "test-token"},
        RetryCount: 0, // Critical: disable retries in tests
    })
    require.NoError(t, err)

    // Call service method
    ctx := context.Background()
    org, err := client.Organizations.Get(ctx, uint(1))

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, uint(1), org.ID)
    assert.Equal(t, "Test Org", org.Name)
}
```

**When to use:**
- Simple, single-scenario tests
- Basic happy path validation
- Quick smoke tests

### Pattern 2: Table-Driven Handler (Multiple Scenarios)

Use for testing multiple scenarios in one test:

```go
func TestOrganizationsService_CreateComprehensive(t *testing.T) {
    tests := []struct {
        name           string
        request        *nexmonyx.CreateOrganizationRequest
        mockStatus     int
        mockBody       interface{}
        wantErr        bool
        validateFunc   func(*testing.T, *nexmonyx.Organization)
    }{
        {
            name: "success - create organization",
            request: &nexmonyx.CreateOrganizationRequest{
                Name:        "New Org",
                Description: "Test organization",
            },
            mockStatus: http.StatusCreated,
            mockBody: map[string]interface{}{
                "data": map[string]interface{}{
                    "id":          1,
                    "name":        "New Org",
                    "description": "Test organization",
                },
            },
            wantErr: false,
            validateFunc: func(t *testing.T, org *nexmonyx.Organization) {
                assert.Equal(t, "New Org", org.Name)
                assert.Equal(t, "Test organization", org.Description)
            },
        },
        {
            name: "validation error - missing name",
            request: &nexmonyx.CreateOrganizationRequest{
                Description: "No name",
            },
            mockStatus: http.StatusBadRequest,
            mockBody:   map[string]interface{}{"error": "name required"},
            wantErr:    true,
        },
        {
            name: "unauthorized - invalid token",
            request: &nexmonyx.CreateOrganizationRequest{
                Name: "Org",
            },
            mockStatus: http.StatusUnauthorized,
            mockBody:   map[string]interface{}{"error": "invalid token"},
            wantErr:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup handler for this scenario
            server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                // Validate common aspects
                assert.Equal(t, http.MethodPost, r.Method)

                // Return scenario-specific response
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(tt.mockStatus)
                json.NewEncoder(w).Encode(tt.mockBody)
            }))
            defer server.Close()

            // Create client
            client, err := nexmonyx.NewClient(&nexmonyx.Config{
                BaseURL:    server.URL,
                Auth:       nexmonyx.AuthConfig{Token: "test-token"},
                RetryCount: 0,
            })
            require.NoError(t, err)

            // Execute
            ctx := context.Background()
            org, err := client.Organizations.Create(ctx, tt.request)

            // Assert error handling
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                if tt.validateFunc != nil {
                    tt.validateFunc(t, org)
                }
            }
        })
    }
}
```

**When to use:**
- Multiple scenarios (success, errors, edge cases)
- Comprehensive coverage of one method
- Related test cases
- Recommended for all handler tests

### Pattern 3: Context-Aware Handler (Error Scenarios)

Use for testing error handling and timeouts:

```go
func TestOrganizationsService_GetWithTimeout(t *testing.T) {
    tests := []struct {
        name           string
        mockStatus     int
        mockBody       interface{}
        useTimeout     bool
        expectError    bool
    }{
        {
            name:       "success - normal response",
            mockStatus: http.StatusOK,
            mockBody: map[string]interface{}{
                "data": map[string]interface{}{"id": 1},
            },
            useTimeout: false,
            expectError: false,
        },
        {
            name:        "server error - with timeout",
            mockStatus:  http.StatusInternalServerError,
            mockBody:    map[string]interface{}{"error": "server error"},
            useTimeout:  true, // Add timeout for error scenarios
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(tt.mockStatus)
                json.NewEncoder(w).Encode(tt.mockBody)
            }))
            defer server.Close()

            client, err := nexmonyx.NewClient(&nexmonyx.Config{
                BaseURL:    server.URL,
                Auth:       nexmonyx.AuthConfig{Token: "test-token"},
                RetryCount: 0,
            })
            require.NoError(t, err)

            // Create context with timeout for error scenarios
            ctx := context.Background()
            if tt.useTimeout {
                var cancel context.CancelFunc
                ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
                defer cancel()
            }

            org, err := client.Organizations.Get(ctx, uint(1))

            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, org)
            }
        })
    }
}
```

**When to use:**
- Error scenarios and edge cases
- Timeout testing
- Context cancellation
- Complex error paths

---

## Best Practices

### ✅ DO

1. **Always disable retries in tests**
   ```go
   RetryCount: 0, // Critical!
   ```
   Reason: Retries add delays and unpredictability

2. **Use fresh mock servers per test case**
   ```go
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) {
           server := httptest.NewServer(...)
           defer server.Close()
           // ...
       })
   }
   ```
   Reason: Prevents test interference and port conflicts

3. **Validate all request aspects**
   ```go
   assert.Equal(t, http.MethodPost, r.Method)
   assert.Equal(t, "/api/v1/organizations", r.URL.Path)
   authHeader := r.Header.Get("Authorization")
   ```
   Reason: Ensures SDK sends correct requests

4. **Use table-driven tests**
   ```go
   tests := []struct {
       name       string
       request    *Request
       mockStatus int
       wantErr    bool
   }{...}
   ```
   Reason: Scales to many scenarios without duplication

5. **Test error scenarios explicitly**
   ```go
   // Include 400, 401, 403, 404, 409, 500 scenarios
   // Not just 200/201
   ```
   Reason: Error handling is critical for production

6. **Use `require.NoError` for setup, `assert.Error` for validation**
   ```go
   require.NoError(t, err)  // Setup must succeed
   assert.Error(t, err)     // Assertion can fail and continue
   ```
   Reason: Distinguishes setup failures from test failures

### ❌ DON'T

1. **Don't use real API calls in tests**
   ```go
   // ❌ BAD
   client, _ := nexmonyx.NewClient(&nexmonyx.Config{
       BaseURL: "https://api.nexmonyx.com",
   })
   ```

2. **Don't share mock servers across test cases**
   ```go
   // ❌ BAD
   server := httptest.NewServer(...)
   for _, tt := range tests {
       // reuse server for all tests
   }
   ```

3. **Don't rely on test execution order**
   ```go
   // ❌ BAD
   // Tests must be independent
   ```

4. **Don't ignore errors**
   ```go
   // ❌ BAD
   client, _ := nexmonyx.NewClient(...)
   ```

5. **Don't test third-party code**
   ```go
   // ❌ BAD - Don't test JSON marshaling or HTTP library
   ```

6. **Don't use `time.Sleep` in tests**
   ```go
   // ❌ BAD
   time.Sleep(100 * time.Millisecond)

   // ✅ GOOD
   ctx, cancel := context.WithTimeout(...)
   ```

---

## Common Patterns

### Testing List Operations (Pagination)

```go
func TestOrganizationsService_ListComprehensive(t *testing.T) {
    tests := []struct {
        name       string
        opts       *nexmonyx.ListOptions
        mockStatus int
        mockBody   interface{}
        wantErr    bool
    }{
        {
            name: "success - list with pagination",
            opts: &nexmonyx.ListOptions{
                Page:  1,
                Limit: 25,
            },
            mockStatus: http.StatusOK,
            mockBody: map[string]interface{}{
                "data": []interface{}{
                    map[string]interface{}{"id": 1, "name": "Org1"},
                    map[string]interface{}{"id": 2, "name": "Org2"},
                },
                "pagination": map[string]interface{}{
                    "page":  1,
                    "limit": 25,
                    "total": 2,
                },
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                // Validate query parameters
                query := r.URL.Query()
                if tt.opts.Page > 0 {
                    assert.Equal(t, "1", query.Get("page"))
                }
                if tt.opts.Limit > 0 {
                    assert.Equal(t, "25", query.Get("limit"))
                }

                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(tt.mockStatus)
                json.NewEncoder(w).Encode(tt.mockBody)
            }))
            defer server.Close()

            client, _ := nexmonyx.NewClient(&nexmonyx.Config{
                BaseURL:    server.URL,
                Auth:       nexmonyx.AuthConfig{Token: "test"},
                RetryCount: 0,
            })

            ctx := context.Background()
            orgs, meta, err := client.Organizations.List(ctx, tt.opts)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Len(t, orgs, 2)
                assert.NotNil(t, meta)
            }
        })
    }
}
```

### Testing CRUD Operations

```go
func TestOrganizationsService_CRUDComprehensive(t *testing.T) {
    // Test all four operations: Create, Read, Update, Delete

    // Test 1: CREATE
    t.Run("create", func(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            assert.Equal(t, http.MethodPost, r.Method)
            w.WriteHeader(http.StatusCreated)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "data": map[string]interface{}{"id": 1, "name": "Org"},
            })
        }))
        defer server.Close()
        // ... test code
    })

    // Test 2: READ
    t.Run("read", func(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            assert.Equal(t, http.MethodGet, r.Method)
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "data": map[string]interface{}{"id": 1, "name": "Org"},
            })
        }))
        defer server.Close()
        // ... test code
    })

    // Test 3: UPDATE
    t.Run("update", func(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            assert.Equal(t, http.MethodPut, r.Method)
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "data": map[string]interface{}{"id": 1, "name": "Updated"},
            })
        }))
        defer server.Close()
        // ... test code
    })

    // Test 4: DELETE
    t.Run("delete", func(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            assert.Equal(t, http.MethodDelete, r.Method)
            w.WriteHeader(http.StatusNoContent)
        }))
        defer server.Close()
        // ... test code
    })
}
```

### Testing Error Responses

```go
func TestOrganizationsService_ErrorHandling(t *testing.T) {
    tests := []struct {
        name           string
        mockStatus     int
        errorResponse  string
        expectedError  string
    }{
        {
            name:          "validation error 400",
            mockStatus:    http.StatusBadRequest,
            errorResponse: `{"error":"name required"}`,
            expectedError: "name required",
        },
        {
            name:          "unauthorized 401",
            mockStatus:    http.StatusUnauthorized,
            errorResponse: `{"error":"invalid token"}`,
            expectedError: "invalid token",
        },
        {
            name:          "forbidden 403",
            mockStatus:    http.StatusForbidden,
            errorResponse: `{"error":"insufficient permissions"}`,
            expectedError: "insufficient permissions",
        },
        {
            name:          "not found 404",
            mockStatus:    http.StatusNotFound,
            errorResponse: `{"error":"organization not found"}`,
            expectedError: "organization not found",
        },
        {
            name:          "conflict 409",
            mockStatus:    http.StatusConflict,
            errorResponse: `{"error":"name already exists"}`,
            expectedError: "name already exists",
        },
        {
            name:          "server error 500",
            mockStatus:    http.StatusInternalServerError,
            errorResponse: `{"error":"internal server error"}`,
            expectedError: "internal server error",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(tt.mockStatus)
                w.Write([]byte(tt.errorResponse))
            }))
            defer server.Close()

            client, _ := nexmonyx.NewClient(&nexmonyx.Config{
                BaseURL:    server.URL,
                Auth:       nexmonyx.AuthConfig{Token: "test"},
                RetryCount: 0,
            })

            ctx := context.Background()
            _, err := client.Organizations.Get(ctx, uint(1))

            assert.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectedError)
        })
    }
}
```

---

## Advanced Techniques

### Request Body Validation

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Decode and validate request body
    var req nexmonyx.CreateOrganizationRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    require.NoError(t, err)

    // Validate fields
    assert.NotEmpty(t, req.Name)
    assert.NotEmpty(t, req.Email)

    // Return response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "data": map[string]interface{}{
            "id":    1,
            "name":  req.Name,
            "email": req.Email,
        },
    })
}))
```

### Header Validation

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Validate content type
    assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

    // Validate authorization
    authHeader := r.Header.Get("Authorization")
    assert.NotEmpty(t, authHeader)
    assert.True(t, strings.HasPrefix(authHeader, "Bearer "))

    // Return response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
}))
```

### Response Headers

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-ID", "req-123")
    w.Header().Set("X-RateLimit-Limit", "1000")
    w.Header().Set("X-RateLimit-Remaining", "999")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(data)
}))
```

---

## Checklist

### Before Writing Handler Tests

- [ ] Understand the service method being tested
- [ ] Identify all HTTP methods used (GET, POST, PUT, DELETE, PATCH)
- [ ] List all possible status codes (200, 201, 204, 400, 401, 403, 404, 409, 500)
- [ ] Identify request validation rules
- [ ] Plan response structure

### During Test Development

- [ ] Use table-driven test pattern
- [ ] Disable retries: `RetryCount: 0`
- [ ] Create fresh server per test case
- [ ] Validate all request aspects (method, path, headers, body)
- [ ] Test success scenario first
- [ ] Add error scenarios (at least 400, 401, 403, 404, 500)
- [ ] Use context with timeout for error cases
- [ ] Use `require` for setup, `assert` for validation
- [ ] Include descriptive test names

### Code Review

- [ ] All tests pass: `go test ./...`
- [ ] No `time.Sleep()` calls
- [ ] No real API calls
- [ ] No shared state between tests
- [ ] All assertions have helpful messages
- [ ] Error messages explain what failed
- [ ] Documentation comments for complex tests
- [ ] Coverage > 80% for the service

### Post-Merge

- [ ] Tests run in CI/CD
- [ ] No flaky tests
- [ ] Coverage maintained or improved
- [ ] Other tests still pass

---

## Example: Complete Handler Test

```go
package nexmonyx

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestOrganizationsService_CreateComprehensive tests Create method with comprehensive coverage
func TestOrganizationsService_CreateComprehensive(t *testing.T) {
    tests := []struct {
        name         string
        request      *CreateOrganizationRequest
        mockStatus   int
        mockBody     interface{}
        wantErr      bool
        validateFunc func(*testing.T, *Organization)
    }{
        {
            name: "success - create organization",
            request: &CreateOrganizationRequest{
                Name:        "Test Org",
                Description: "Test Description",
            },
            mockStatus: http.StatusCreated,
            mockBody: map[string]interface{}{
                "data": map[string]interface{}{
                    "id":          1,
                    "name":        "Test Org",
                    "description": "Test Description",
                },
            },
            wantErr: false,
            validateFunc: func(t *testing.T, org *Organization) {
                assert.Equal(t, uint(1), org.ID)
                assert.Equal(t, "Test Org", org.Name)
                assert.Equal(t, "Test Description", org.Description)
            },
        },
        {
            name: "validation error - missing name",
            request: &CreateOrganizationRequest{
                Description: "No Name",
            },
            mockStatus: http.StatusBadRequest,
            mockBody:   map[string]interface{}{"error": "name required"},
            wantErr:    true,
        },
        {
            name: "unauthorized",
            request: &CreateOrganizationRequest{
                Name: "Org",
            },
            mockStatus: http.StatusUnauthorized,
            mockBody:   map[string]interface{}{"error": "invalid token"},
            wantErr:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock server
            server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                // Validate request
                assert.Equal(t, http.MethodPost, r.Method, "expected POST method")
                assert.Equal(t, "/api/v1/organizations", r.URL.Path, "expected correct path")

                // Decode and validate body
                var req CreateOrganizationRequest
                err := json.NewDecoder(r.Body).Decode(&req)
                if err == nil {
                    assert.NotEmpty(t, req.Name, "name should not be empty")
                }

                // Return mock response
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(tt.mockStatus)
                json.NewEncoder(w).Encode(tt.mockBody)
            }))
            defer server.Close()

            // Create client
            client, err := NewClient(&Config{
                BaseURL:    server.URL,
                Auth:       AuthConfig{Token: "test-token"},
                RetryCount: 0, // Critical: disable retries
            })
            require.NoError(t, err, "failed to create client")

            // Execute service method
            ctx := context.Background()
            org, err := client.Organizations.Create(ctx, tt.request)

            // Validate results
            if tt.wantErr {
                assert.Error(t, err, "expected error for %s", tt.name)
            } else {
                assert.NoError(t, err, "unexpected error for %s", tt.name)
                require.NotNil(t, org, "expected organization response")
                if tt.validateFunc != nil {
                    tt.validateFunc(t, org)
                }
            }
        })
    }
}
```

---

## References

- [Go Testing Package](https://golang.org/pkg/testing/)
- [httptest Documentation](https://golang.org/pkg/net/http/httptest/)
- [Testify Library](https://github.com/stretchr/testify)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [SDK Testing Guide](./TESTING.md)
- [Example Tests](../examples/testing/)

---

## Questions?

For questions about handler testing patterns:
1. Review this document section on common patterns
2. Check existing test files: `*_comprehensive_test.go`
3. See example tests: `examples/testing/`
4. Consult the SDK testing guide: `TESTING.md`

**Maintainer**: Nexmonyx Development Team
**Last Review**: 2025-10-17
