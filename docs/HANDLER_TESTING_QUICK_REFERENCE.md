# Handler Testing - Quick Reference Card

**Print this for your desk!** Quick reference for writing handler tests in the Nexmonyx Go SDK.

---

## Minimal Handler Test Template

```go
func TestServiceName_MethodComprehensive(t *testing.T) {
    tests := []struct {
        name       string
        request    *Request
        mockStatus int
        mockBody   interface{}
        wantErr    bool
    }{
        {
            name:       "success - description",
            request:    &Request{Field: "value"},
            mockStatus: http.StatusOK,
            mockBody:   map[string]interface{}{"data": map[string]interface{}{"id": 1}},
            wantErr:    false,
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

            client, _ := NewClient(&Config{
                BaseURL:    server.URL,
                Auth:       AuthConfig{Token: "test"},
                RetryCount: 0,
            })

            result, err := client.Service.Method(context.Background(), tt.request)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
            }
        })
    }
}
```

---

## 🎯 Critical Settings

```go
// ALWAYS disable retries in tests
RetryCount: 0  // ← Critical!

// ALWAYS use fresh server per test case
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        server := httptest.NewServer(...)  // ← Fresh per test
        defer server.Close()
        // ...
    })
}
```

---

## Request Validation Checklist

```go
// Check method
assert.Equal(t, http.MethodPost, r.Method)

// Check path
assert.Equal(t, "/api/v1/resource", r.URL.Path)

// Check headers
assert.NotEmpty(t, r.Header.Get("Authorization"))

// Check body (POST/PUT)
var req MyRequest
json.NewDecoder(r.Body).Decode(&req)
assert.NotEmpty(t, req.Name)
```

---

## Response Construction Checklist

```go
// Set content type
w.Header().Set("Content-Type", "application/json")

// Set status code
w.WriteHeader(tt.mockStatus)

// Encode body
json.NewEncoder(w).Encode(tt.mockBody)
```

---

## Required Test Scenarios

Every method needs tests for:

| Scenario | Status | When |
|----------|--------|------|
| Success | 200/201/204 | Happy path |
| Validation Error | 400 | Invalid input |
| Unauthorized | 401 | Missing/invalid auth |
| Forbidden | 403 | Insufficient permissions |
| Not Found | 404 | Resource doesn't exist |
| Conflict | 409 | Duplicate/constraint violation |
| Server Error | 500 | Backend failure |

---

## Error Test Example

```go
{
    name:       "unauthorized - invalid token",
    request:    &Request{...},
    mockStatus: http.StatusUnauthorized,
    mockBody:   map[string]interface{}{"error": "invalid token"},
    wantErr:    true,
},
```

---

## Common HTTP Methods

```go
http.MethodGet     // GET    - fetch
http.MethodPost    // POST   - create
http.MethodPut     // PUT    - update
http.MethodDelete  // DELETE - delete
http.MethodPatch   // PATCH  - partial update
```

---

## Assert vs Require

```go
// Setup must succeed - use require
require.NoError(t, err)
require.NotNil(t, obj)

// Assertions can fail and continue - use assert
assert.Error(t, err)
assert.Equal(t, expected, actual)
```

---

## Useful Status Codes

```go
http.StatusOK                  // 200 - success
http.StatusCreated             // 201 - created
http.StatusNoContent           // 204 - no content
http.StatusBadRequest          // 400 - validation error
http.StatusUnauthorized        // 401 - auth required
http.StatusForbidden           // 403 - insufficient perms
http.StatusNotFound            // 404 - not found
http.StatusConflict            // 409 - conflict/duplicate
http.StatusInternalServerError // 500 - server error
```

---

## List Operation Test

```go
{
    name: "success - list with pagination",
    opts: &ListOptions{Page: 1, Limit: 25},
    mockStatus: http.StatusOK,
    mockBody: map[string]interface{}{
        "data": []interface{}{
            map[string]interface{}{"id": 1},
            map[string]interface{}{"id": 2},
        },
    },
    wantErr: false,
},
```

---

## Context with Timeout

```go
ctx := context.Background()

// For error scenarios, add timeout
if tt.mockStatus >= 500 {
    var cancel context.CancelFunc
    ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
}
```

---

## Handler Path Patterns

```
GET    /api/v1/organizations       → List
POST   /api/v1/organizations       → Create
GET    /api/v1/organizations/:id   → Get
PUT    /api/v1/organizations/:id   → Update
DELETE /api/v1/organizations/:id   → Delete
```

---

## Common Validations

```go
// Validate method and path
assert.Equal(t, http.MethodPost, r.Method)
assert.Equal(t, "/api/v1/organizations", r.URL.Path)

// Validate auth header
authHeader := r.Header.Get("Authorization")
assert.NotEmpty(t, authHeader)
assert.Contains(t, authHeader, "Bearer ")

// Validate content type
assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

// Parse and validate body
var req CreateRequest
err := json.NewDecoder(r.Body).Decode(&req)
require.NoError(t, err)
assert.NotEmpty(t, req.Name)
```

---

## Response Examples

### Success (GET)
```go
mockStatus: http.StatusOK,
mockBody: map[string]interface{}{
    "data": map[string]interface{}{
        "id":   1,
        "name": "Test",
    },
},
```

### Success (POST)
```go
mockStatus: http.StatusCreated,
mockBody: map[string]interface{}{
    "data": map[string]interface{}{
        "id":   1,
        "name": "Created",
    },
},
```

### Error
```go
mockStatus: http.StatusBadRequest,
mockBody: map[string]interface{}{
    "error": "name is required",
},
```

---

## File Structure

```
services/
├── organizations.go              ← Implementation
├── organizations_test.go         ← Basic tests
└── organizations_comprehensive_test.go  ← Handler tests (NEW)
```

---

## Imports Required

```go
import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

---

## Test Naming Convention

```
TestServiceName_MethodComprehensive
         ↓              ↓
    Service       What it tests
```

Examples:
- `TestOrganizationsService_CreateComprehensive`
- `TestServersService_ListComprehensive`
- `TestUsersService_GetComprehensive`

---

## DO & DON'T Quick Reference

| DO | DON'T |
|---|---|
| ✅ Use table-driven tests | ❌ Use real API calls |
| ✅ Disable retries | ❌ Share servers across tests |
| ✅ Fresh server per case | ❌ Use `time.Sleep()` |
| ✅ Test all status codes | ❌ Test 3rd party libraries |
| ✅ Validate requests | ❌ Ignore setup errors |
| ✅ Use descriptive names | ❌ Duplicate test code |

---

## Coverage Targets

- **Service methods**: 80%+
- **Error handling**: 90%+
- **Critical paths**: 100%

---

## Useful Files to Reference

- `examples/testing/integration/mock_api_test.go` - Example handler patterns
- `TESTING.md` - Comprehensive guide
- `*_comprehensive_test.go` - Existing handler tests in repo
- `docs/HANDLER_TESTING_STANDARDS.md` - Full standards document

---

## Quick Debug Tips

```bash
# Run specific test
go test -v -run TestOrganizationsService_CreateComprehensive

# Run with verbose output
go test -v ./...

# Check coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

**Print this card and keep it handy while writing tests!**

For complete documentation, see: `docs/HANDLER_TESTING_STANDARDS.md`
