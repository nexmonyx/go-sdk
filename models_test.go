package nexmonyx

import (
	"encoding/json"
	"testing"
	"time"
)

// ============================================================================
// CustomTime Additional Tests (beyond what's in client_test.go)
// ============================================================================

// TestCustomTime_MarshalJSON tests CustomTime marshaling
func TestCustomTime_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		time     CustomTime
		wantNull bool
	}{
		{
			name:     "zero time returns null",
			time:     CustomTime{},
			wantNull: true,
		},
		{
			name:     "valid time returns RFC3339",
			time:     CustomTime{Time: time.Date(2023, 10, 14, 12, 30, 45, 0, time.UTC)},
			wantNull: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.time.MarshalJSON()
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			if tt.wantNull {
				if string(data) != "null" {
					t.Errorf("expected null, got %s", string(data))
				}
			} else {
				if string(data) == "null" {
					t.Error("expected non-null value, got null")
				}
				// Verify it's valid JSON string
				var s string
				if err := json.Unmarshal(data, &s); err != nil {
					t.Errorf("marshaled value is not valid JSON string: %v", err)
				}
			}
		})
	}
}

// TestCustomTime_IsZero tests the IsZero behavior
func TestCustomTime_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		time     CustomTime
		expected bool
	}{
		{
			name:     "zero time",
			time:     CustomTime{},
			expected: true,
		},
		{
			name:     "non-zero time",
			time:     CustomTime{Time: time.Now()},
			expected: false,
		},
		{
			name:     "explicitly zero",
			time:     CustomTime{Time: time.Time{}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.time.IsZero(); got != tt.expected {
				t.Errorf("IsZero() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestCustomTime_RoundTrip tests full marshal/unmarshal cycle
func TestCustomTime_RoundTrip(t *testing.T) {
	original := CustomTime{Time: time.Date(2023, 10, 14, 12, 30, 45, 0, time.UTC)}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal
	var decoded CustomTime
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Compare (truncate to seconds as JSON doesn't preserve nanoseconds)
	if !original.Time.Truncate(time.Second).Equal(decoded.Time.Truncate(time.Second)) {
		t.Errorf("times don't match: original=%v, decoded=%v", original.Time, decoded.Time)
	}
}

// ============================================================================
// Base Model Tests
// ============================================================================

// TestGormModel_JSON tests GormModel JSON serialization
func TestGormModel_JSON(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	model := GormModel{
		ID:        123,
		CreatedAt: &now,
		UpdatedAt: &now,
		DeletedAt: nil,
	}

	// Marshal
	data, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("failed to marshal GormModel: %v", err)
	}

	// Unmarshal
	var decoded GormModel
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal GormModel: %v", err)
	}

	// Verify fields
	if decoded.ID != model.ID {
		t.Errorf("ID = %d, want %d", decoded.ID, model.ID)
	}
	if decoded.CreatedAt == nil {
		t.Error("CreatedAt should not be nil")
	}
	if decoded.DeletedAt != nil {
		t.Error("DeletedAt should be nil")
	}
}

// TestBaseModel_JSON tests BaseModel JSON serialization
func TestBaseModel_JSON(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	model := BaseModel{
		UUID:      "test-uuid-123",
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	// Marshal
	data, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("failed to marshal BaseModel: %v", err)
	}

	// Unmarshal
	var decoded BaseModel
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal BaseModel: %v", err)
	}

	// Verify fields
	if decoded.UUID != model.UUID {
		t.Errorf("UUID = %s, want %s", decoded.UUID, model.UUID)
	}
	if decoded.CreatedAt == nil {
		t.Error("CreatedAt should not be nil")
	}
}

// TestPaginationOptions_JSON tests PaginationOptions serialization
func TestPaginationOptions_JSON(t *testing.T) {
	opts := PaginationOptions{
		Page:  2,
		Limit: 50,
	}

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("failed to marshal PaginationOptions: %v", err)
	}

	var decoded PaginationOptions
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal PaginationOptions: %v", err)
	}

	if decoded.Page != opts.Page {
		t.Errorf("Page = %d, want %d", decoded.Page, opts.Page)
	}
	if decoded.Limit != opts.Limit {
		t.Errorf("Limit = %d, want %d", decoded.Limit, opts.Limit)
	}
}

// ============================================================================
// Domain Model Tests
// ============================================================================

// TestOrganization_JSON tests Organization model serialization
func TestOrganization_JSON(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	org := Organization{
		GormModel: GormModel{
			ID:        1,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
		UUID:              "org-uuid-123",
		Name:              "Test Organization",
		Description:       "Test description",
		Industry:          "Technology",
		Website:           "https://example.com",
		Size:              "50-200",
		Country:           "US",
		TimeZone:          "America/New_York",
		MaxServers:        100,
		MaxUsers:          50,
		MaxProbes:         200,
		AlertsEnabled:     true,
		MonitoringEnabled: true,
		Tags:              []string{"test", "demo"},
	}

	// Marshal
	data, err := json.Marshal(org)
	if err != nil {
		t.Fatalf("failed to marshal Organization: %v", err)
	}

	// Unmarshal
	var decoded Organization
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Organization: %v", err)
	}

	// Verify key fields
	if decoded.UUID != org.UUID {
		t.Errorf("UUID = %s, want %s", decoded.UUID, org.UUID)
	}
	if decoded.Name != org.Name {
		t.Errorf("Name = %s, want %s", decoded.Name, org.Name)
	}
	if decoded.MaxServers != org.MaxServers {
		t.Errorf("MaxServers = %d, want %d", decoded.MaxServers, org.MaxServers)
	}
	if len(decoded.Tags) != len(org.Tags) {
		t.Errorf("Tags length = %d, want %d", len(decoded.Tags), len(org.Tags))
	}
}

// TestUser_JSON tests User model serialization
func TestUser_JSON(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	user := User{
		GormModel: GormModel{
			ID:        1,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
		Email:            "test@example.com",
		FirstName:        "John",
		LastName:         "Doe",
		DisplayName:      "John Doe",
		PhoneNumber:      "+1234567890",
		AuthProvider:     "native",
		LastLogin:        &now,
		EmailVerified:    true,
		TwoFactorEnabled: false,
		OrganizationID:   1,
		Role:             "admin",
		Permissions:      []string{"read", "write", "delete"},
		IsActive:         true,
		IsAdmin:          true,
		Timezone:         "America/New_York",
		Language:         "en",
	}

	// Marshal
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal User: %v", err)
	}

	// Unmarshal
	var decoded User
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal User: %v", err)
	}

	// Verify key fields
	if decoded.Email != user.Email {
		t.Errorf("Email = %s, want %s", decoded.Email, user.Email)
	}
	if decoded.FirstName != user.FirstName {
		t.Errorf("FirstName = %s, want %s", decoded.FirstName, user.FirstName)
	}
	if decoded.IsActive != user.IsActive {
		t.Errorf("IsActive = %v, want %v", decoded.IsActive, user.IsActive)
	}
	if len(decoded.Permissions) != len(user.Permissions) {
		t.Errorf("Permissions length = %d, want %d", len(decoded.Permissions), len(user.Permissions))
	}
}

// TestServer_JSON tests Server model serialization
func TestServer_JSON(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	server := Server{
		GormModel: GormModel{
			ID:        1,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
		ServerUUID:        "server-uuid-123",
		Hostname:          "web-server-01",
		FQDN:              "web-server-01.example.com",
		OrganizationID:    1,
		OS:                "Ubuntu",
		OSVersion:         "22.04",
		OSArch:            "x86_64",
		KernelVersion:     "5.15.0",
		CPUArchitecture:   "x86_64",
		CPUModel:          "Intel Xeon",
		CPUCores:          8,
		TotalMemoryGB:     32.0,
		TotalDiskGB:       500.0,
		MainIP:            "192.168.1.100",
		IPv6Address:       "fe80::1",
		Environment:       "production",
		Location:          "US-East",
		DataCenter:        "DC1",
		LastHeartbeat:     &now,
		Status:            "active",
		AgentVersion:      "1.2.3",
		MonitoringEnabled: true,
		AlertsEnabled:     true,
		Provider:          "AWS",
		InstanceType:      "t3.medium",
		Region:            "us-east-1",
		Tags:              []string{"web", "production"},
	}

	// Marshal
	data, err := json.Marshal(server)
	if err != nil {
		t.Fatalf("failed to marshal Server: %v", err)
	}

	// Unmarshal
	var decoded Server
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Server: %v", err)
	}

	// Verify key fields
	if decoded.ServerUUID != server.ServerUUID {
		t.Errorf("ServerUUID = %s, want %s", decoded.ServerUUID, server.ServerUUID)
	}
	if decoded.Hostname != server.Hostname {
		t.Errorf("Hostname = %s, want %s", decoded.Hostname, server.Hostname)
	}
	if decoded.CPUCores != server.CPUCores {
		t.Errorf("CPUCores = %d, want %d", decoded.CPUCores, server.CPUCores)
	}
	if decoded.TotalMemoryGB != server.TotalMemoryGB {
		t.Errorf("TotalMemoryGB = %f, want %f", decoded.TotalMemoryGB, server.TotalMemoryGB)
	}
}

// TestEmptyModel_JSON tests JSON serialization of empty models
func TestEmptyModel_JSON(t *testing.T) {
	tests := []struct {
		name  string
		model interface{}
	}{
		{"empty Organization", &Organization{}},
		{"empty User", &User{}},
		{"empty Server", &Server{}},
		{"empty GormModel", &GormModel{}},
		{"empty BaseModel", &BaseModel{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should be able to marshal empty models
			data, err := json.Marshal(tt.model)
			if err != nil {
				t.Errorf("failed to marshal empty model: %v", err)
			}

			// Should be able to unmarshal back
			if err := json.Unmarshal(data, tt.model); err != nil {
				t.Errorf("failed to unmarshal empty model: %v", err)
			}
		})
	}
}

// TestNilPointerFields tests models with nil pointer fields
func TestNilPointerFields(t *testing.T) {
	org := Organization{
		UUID: "test-uuid",
		Name: "Test Org",
		// All optional pointer fields are nil
		TrialEndsAt:      nil,
		BillingContact:   nil,
		TechnicalContact: nil,
	}

	data, err := json.Marshal(org)
	if err != nil {
		t.Fatalf("failed to marshal with nil pointers: %v", err)
	}

	var decoded Organization
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal with nil pointers: %v", err)
	}

	if decoded.TrialEndsAt != nil {
		t.Error("TrialEndsAt should be nil")
	}
	if decoded.BillingContact != nil {
		t.Error("BillingContact should be nil")
	}
}

// TestMapFields tests models with map[string]interface{} fields
func TestMapFields(t *testing.T) {
	org := Organization{
		UUID: "test-uuid",
		Name: "Test Org",
		Settings: map[string]interface{}{
			"theme":         "dark",
			"notifications": true,
			"max_retries":   3,
		},
		Metadata: map[string]interface{}{
			"created_by": "admin",
			"version":    "1.0",
		},
	}

	data, err := json.Marshal(org)
	if err != nil {
		t.Fatalf("failed to marshal with map fields: %v", err)
	}

	var decoded Organization
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal with map fields: %v", err)
	}

	if len(decoded.Settings) != len(org.Settings) {
		t.Errorf("Settings length = %d, want %d", len(decoded.Settings), len(org.Settings))
	}
	if len(decoded.Metadata) != len(org.Metadata) {
		t.Errorf("Metadata length = %d, want %d", len(decoded.Metadata), len(org.Metadata))
	}
}

// TestSliceFields tests models with slice fields
func TestSliceFields(t *testing.T) {
	org := Organization{
		UUID: "test-uuid",
		Name: "Test Org",
		Tags: []string{"production", "critical", "monitored"},
	}

	data, err := json.Marshal(org)
	if err != nil {
		t.Fatalf("failed to marshal with slice fields: %v", err)
	}

	var decoded Organization
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal with slice fields: %v", err)
	}

	if len(decoded.Tags) != len(org.Tags) {
		t.Errorf("Tags length = %d, want %d", len(decoded.Tags), len(org.Tags))
	}
	for i, tag := range decoded.Tags {
		if tag != org.Tags[i] {
			t.Errorf("Tags[%d] = %s, want %s", i, tag, org.Tags[i])
		}
	}
}

// TestOmitemptyFields tests that omitempty works correctly
func TestOmitemptyFields(t *testing.T) {
	// Create model with only required fields
	org := Organization{
		UUID: "test-uuid",
		Name: "Test Org",
		// All omitempty fields are zero values
	}

	data, err := json.Marshal(org)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Verify certain fields are not in JSON
	jsonStr := string(data)
	if contains(jsonStr, "description") {
		t.Error("empty description should be omitted")
	}
	if contains(jsonStr, "website") {
		t.Error("empty website should be omitted")
	}
}

// TestNestedStructs tests models with nested struct fields
func TestNestedStructs(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	org := Organization{
		GormModel: GormModel{
			ID:        1,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
		UUID: "org-uuid",
		Name: "Test Org",
		BillingContact: &User{
			GormModel: GormModel{
				ID: 10,
			},
			Email:     "billing@example.com",
			FirstName: "Billing",
			LastName:  "User",
		},
	}

	data, err := json.Marshal(org)
	if err != nil {
		t.Fatalf("failed to marshal nested structs: %v", err)
	}

	var decoded Organization
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal nested structs: %v", err)
	}

	if decoded.BillingContact == nil {
		t.Fatal("BillingContact should not be nil")
	}
	if decoded.BillingContact.Email != "billing@example.com" {
		t.Errorf("BillingContact.Email = %s, want billing@example.com", decoded.BillingContact.Email)
	}
}

// ============================================================================
// UnifiedAPIKey Method Tests
// ============================================================================

// TestUnifiedAPIKey_IsActive tests the IsActive method
func TestUnifiedAPIKey_IsActive(t *testing.T) {
	futureTime := CustomTime{Time: time.Now().Add(24 * time.Hour)}
	pastTime := CustomTime{Time: time.Now().Add(-24 * time.Hour)}

	tests := []struct {
		name     string
		key      *UnifiedAPIKey
		expected bool
	}{
		{
			name: "active key without expiration",
			key: &UnifiedAPIKey{
				Status:    APIKeyStatusActive,
				ExpiresAt: nil,
			},
			expected: true,
		},
		{
			name: "active key not expired",
			key: &UnifiedAPIKey{
				Status:    APIKeyStatusActive,
				ExpiresAt: &futureTime,
			},
			expected: true,
		},
		{
			name: "active key but expired",
			key: &UnifiedAPIKey{
				Status:    APIKeyStatusActive,
				ExpiresAt: &pastTime,
			},
			expected: false,
		},
		{
			name: "revoked key",
			key: &UnifiedAPIKey{
				Status:    APIKeyStatusRevoked,
				ExpiresAt: nil,
			},
			expected: false,
		},
		{
			name: "expired status",
			key: &UnifiedAPIKey{
				Status:    APIKeyStatusExpired,
				ExpiresAt: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.key.IsActive(); got != tt.expected {
				t.Errorf("IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_IsExpired tests the IsExpired method
func TestUnifiedAPIKey_IsExpired(t *testing.T) {
	futureTime := CustomTime{Time: time.Now().Add(24 * time.Hour)}
	pastTime := CustomTime{Time: time.Now().Add(-24 * time.Hour)}

	tests := []struct {
		name     string
		key      *UnifiedAPIKey
		expected bool
	}{
		{
			name: "not expired - future expiration",
			key: &UnifiedAPIKey{
				ExpiresAt: &futureTime,
			},
			expected: false,
		},
		{
			name: "expired - past expiration",
			key: &UnifiedAPIKey{
				ExpiresAt: &pastTime,
			},
			expected: true,
		},
		{
			name: "no expiration set",
			key: &UnifiedAPIKey{
				ExpiresAt: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.key.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_IsRevoked tests the IsRevoked method
func TestUnifiedAPIKey_IsRevoked(t *testing.T) {
	tests := []struct {
		name     string
		status   APIKeyStatus
		expected bool
	}{
		{"active key", APIKeyStatusActive, false},
		{"revoked key", APIKeyStatusRevoked, true},
		{"expired key", APIKeyStatusExpired, false},
		{"pending key", APIKeyStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &UnifiedAPIKey{Status: tt.status}
			if got := key.IsRevoked(); got != tt.expected {
				t.Errorf("IsRevoked() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_HasCapability tests the HasCapability method
func TestUnifiedAPIKey_HasCapability(t *testing.T) {
	tests := []struct {
		name         string
		capabilities []string
		check        string
		expected     bool
	}{
		{
			name:         "has specific capability",
			capabilities: []string{"servers:read", "servers:write"},
			check:        "servers:read",
			expected:     true,
		},
		{
			name:         "does not have capability",
			capabilities: []string{"servers:read"},
			check:        "servers:write",
			expected:     false,
		},
		{
			name:         "has wildcard capability",
			capabilities: []string{"*"},
			check:        "anything",
			expected:     true,
		},
		{
			name:         "wildcard in list",
			capabilities: []string{"servers:read", "*"},
			check:        "admin:delete",
			expected:     true,
		},
		{
			name:         "empty capabilities",
			capabilities: []string{},
			check:        "servers:read",
			expected:     false,
		},
		{
			name:         "nil capabilities",
			capabilities: nil,
			check:        "servers:read",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &UnifiedAPIKey{Capabilities: tt.capabilities}
			if got := key.HasCapability(tt.check); got != tt.expected {
				t.Errorf("HasCapability(%s) = %v, want %v", tt.check, got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_HasScope tests the HasScope method
func TestUnifiedAPIKey_HasScope(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		check    string
		expected bool
	}{
		{
			name:     "has specific scope",
			scopes:   []string{"read:servers", "write:servers"},
			check:    "read:servers",
			expected: true,
		},
		{
			name:     "does not have scope",
			scopes:   []string{"read:servers"},
			check:    "write:servers",
			expected: false,
		},
		{
			name:     "has wildcard scope",
			scopes:   []string{"*"},
			check:    "anything",
			expected: true,
		},
		{
			name:     "empty scopes",
			scopes:   []string{},
			check:    "read:servers",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &UnifiedAPIKey{Scopes: tt.scopes}
			if got := key.HasScope(tt.check); got != tt.expected {
				t.Errorf("HasScope(%s) = %v, want %v", tt.check, got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_IsMonitoringAgent tests the IsMonitoringAgent method
func TestUnifiedAPIKey_IsMonitoringAgent(t *testing.T) {
	tests := []struct {
		name     string
		keyType  APIKeyType
		expected bool
	}{
		{"monitoring agent key", APIKeyTypeMonitoringAgent, true},
		{"public agent key", APIKeyTypePublicAgent, true},
		{"user key", APIKeyTypeUser, false},
		{"admin key", APIKeyTypeAdmin, false},
		{"registration key", APIKeyTypeRegistration, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &UnifiedAPIKey{Type: tt.keyType}
			if got := key.IsMonitoringAgent(); got != tt.expected {
				t.Errorf("IsMonitoringAgent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_IsRegistrationKey tests the IsRegistrationKey method
func TestUnifiedAPIKey_IsRegistrationKey(t *testing.T) {
	tests := []struct {
		name     string
		keyType  APIKeyType
		expected bool
	}{
		{"registration key", APIKeyTypeRegistration, true},
		{"user key", APIKeyTypeUser, false},
		{"monitoring agent key", APIKeyTypeMonitoringAgent, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &UnifiedAPIKey{Type: tt.keyType}
			if got := key.IsRegistrationKey(); got != tt.expected {
				t.Errorf("IsRegistrationKey() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_CanRegisterServers tests the CanRegisterServers method
func TestUnifiedAPIKey_CanRegisterServers(t *testing.T) {
	tests := []struct {
		name         string
		keyType      APIKeyType
		capabilities []string
		expected     bool
	}{
		{
			name:     "registration key",
			keyType:  APIKeyTypeRegistration,
			expected: true,
		},
		{
			name:         "has servers:register capability",
			keyType:      APIKeyTypeUser,
			capabilities: []string{"servers:register"},
			expected:     true,
		},
		{
			name:         "has servers:* capability",
			keyType:      APIKeyTypeUser,
			capabilities: []string{"servers:*"},
			expected:     true,
		},
		{
			name:         "has wildcard capability",
			keyType:      APIKeyTypeUser,
			capabilities: []string{"*"},
			expected:     true,
		},
		{
			name:         "no register capability",
			keyType:      APIKeyTypeUser,
			capabilities: []string{"servers:read"},
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &UnifiedAPIKey{
				Type:         tt.keyType,
				Capabilities: tt.capabilities,
			}
			if got := key.CanRegisterServers(); got != tt.expected {
				t.Errorf("CanRegisterServers() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_CanAccessOrganization tests the CanAccessOrganization method
func TestUnifiedAPIKey_CanAccessOrganization(t *testing.T) {
	tests := []struct {
		name     string
		keyType  APIKeyType
		keyOrgID uint
		checkOrg uint
		expected bool
	}{
		{
			name:     "same organization",
			keyType:  APIKeyTypeUser,
			keyOrgID: 1,
			checkOrg: 1,
			expected: true,
		},
		{
			name:     "different organization",
			keyType:  APIKeyTypeUser,
			keyOrgID: 1,
			checkOrg: 2,
			expected: false,
		},
		{
			name:     "system key can access any org",
			keyType:  APIKeyTypeSystem,
			keyOrgID: 1,
			checkOrg: 2,
			expected: true,
		},
		{
			name:     "admin key can access any org",
			keyType:  APIKeyTypeAdmin,
			keyOrgID: 1,
			checkOrg: 2,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &UnifiedAPIKey{
				Type:           tt.keyType,
				OrganizationID: tt.keyOrgID,
			}
			if got := key.CanAccessOrganization(tt.checkOrg); got != tt.expected {
				t.Errorf("CanAccessOrganization(%d) = %v, want %v", tt.checkOrg, got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_IsPublicAgent tests the IsPublicAgent method
func TestUnifiedAPIKey_IsPublicAgent(t *testing.T) {
	tests := []struct {
		name      string
		keyType   APIKeyType
		agentType string
		expected  bool
	}{
		{
			name:     "public agent key type",
			keyType:  APIKeyTypePublicAgent,
			expected: true,
		},
		{
			name:      "monitoring agent with public type",
			keyType:   APIKeyTypeMonitoringAgent,
			agentType: "public",
			expected:  true,
		},
		{
			name:      "monitoring agent with private type",
			keyType:   APIKeyTypeMonitoringAgent,
			agentType: "private",
			expected:  false,
		},
		{
			name:     "user key",
			keyType:  APIKeyTypeUser,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &UnifiedAPIKey{
				Type:      tt.keyType,
				AgentType: tt.agentType,
			}
			if got := key.IsPublicAgent(); got != tt.expected {
				t.Errorf("IsPublicAgent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_IsPrivateAgent tests the IsPrivateAgent method
func TestUnifiedAPIKey_IsPrivateAgent(t *testing.T) {
	tests := []struct {
		name      string
		keyType   APIKeyType
		agentType string
		expected  bool
	}{
		{
			name:      "monitoring agent with private type",
			keyType:   APIKeyTypeMonitoringAgent,
			agentType: "private",
			expected:  true,
		},
		{
			name:      "monitoring agent with public type",
			keyType:   APIKeyTypeMonitoringAgent,
			agentType: "public",
			expected:  false,
		},
		{
			name:     "public agent key",
			keyType:  APIKeyTypePublicAgent,
			expected: false,
		},
		{
			name:     "user key",
			keyType:  APIKeyTypeUser,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &UnifiedAPIKey{
				Type:      tt.keyType,
				AgentType: tt.agentType,
			}
			if got := key.IsPrivateAgent(); got != tt.expected {
				t.Errorf("IsPrivateAgent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_GetAuthenticationMethod tests the GetAuthenticationMethod method
func TestUnifiedAPIKey_GetAuthenticationMethod(t *testing.T) {
	tests := []struct {
		name     string
		keyType  APIKeyType
		expected string
	}{
		{
			name:     "monitoring agent uses bearer",
			keyType:  APIKeyTypeMonitoringAgent,
			expected: "bearer",
		},
		{
			name:     "public agent uses bearer",
			keyType:  APIKeyTypePublicAgent,
			expected: "bearer",
		},
		{
			name:     "registration uses headers",
			keyType:  APIKeyTypeRegistration,
			expected: "headers",
		},
		{
			name:     "user uses headers",
			keyType:  APIKeyTypeUser,
			expected: "headers",
		},
		{
			name:     "admin uses headers",
			keyType:  APIKeyTypeAdmin,
			expected: "headers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &UnifiedAPIKey{Type: tt.keyType}
			if got := key.GetAuthenticationMethod(); got != tt.expected {
				t.Errorf("GetAuthenticationMethod() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestUnifiedAPIKey_JSON tests UnifiedAPIKey JSON serialization
func TestUnifiedAPIKey_JSON(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	key := UnifiedAPIKey{
		GormModel: GormModel{
			ID:        1,
			CreatedAt: &now,
		},
		KeyID:          "key-123",
		Name:           "Test Key",
		Description:    "Test Description",
		Type:           APIKeyTypeUser,
		Capabilities:   []string{"servers:read", "servers:write"},
		OrganizationID: 1,
		Status:         APIKeyStatusActive,
		UsageCount:     42,
		Tags:           []string{"test", "development"},
	}

	// Marshal
	data, err := json.Marshal(key)
	if err != nil {
		t.Fatalf("failed to marshal UnifiedAPIKey: %v", err)
	}

	// Unmarshal
	var decoded UnifiedAPIKey
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal UnifiedAPIKey: %v", err)
	}

	// Verify key fields
	if decoded.KeyID != key.KeyID {
		t.Errorf("KeyID = %s, want %s", decoded.KeyID, key.KeyID)
	}
	if decoded.Name != key.Name {
		t.Errorf("Name = %s, want %s", decoded.Name, key.Name)
	}
	if decoded.Type != key.Type {
		t.Errorf("Type = %s, want %s", decoded.Type, key.Type)
	}
	if decoded.Status != key.Status {
		t.Errorf("Status = %s, want %s", decoded.Status, key.Status)
	}
	if decoded.UsageCount != key.UsageCount {
		t.Errorf("UsageCount = %d, want %d", decoded.UsageCount, key.UsageCount)
	}
}

// TestAPIKeyStatus_String tests APIKeyStatus constants
func TestAPIKeyStatus_String(t *testing.T) {
	statuses := []APIKeyStatus{
		APIKeyStatusActive,
		APIKeyStatusRevoked,
		APIKeyStatusExpired,
		APIKeyStatusPending,
	}

	for _, status := range statuses {
		// Should be able to convert to string
		s := string(status)
		if s == "" {
			t.Errorf("status %v converted to empty string", status)
		}

		// Should be able to use in comparisons
		if status != APIKeyStatus(s) {
			t.Errorf("status comparison failed for %v", status)
		}
	}
}

// TestAPIKeyType_String tests APIKeyType constants
func TestAPIKeyType_String(t *testing.T) {
	types := []APIKeyType{
		APIKeyTypeUser,
		APIKeyTypeAdmin,
		APIKeyTypeMonitoringAgent,
		APIKeyTypeSystem,
		APIKeyTypePublicAgent,
		APIKeyTypeRegistration,
		APIKeyTypeOrgMonitoring,
	}

	for _, keyType := range types {
		// Should be able to convert to string
		s := string(keyType)
		if s == "" {
			t.Errorf("keyType %v converted to empty string", keyType)
		}

		// Should be able to use in comparisons
		if keyType != APIKeyType(s) {
			t.Errorf("keyType comparison failed for %v", keyType)
		}
	}
}

// ============================================================================
// ServerDetailsUpdateRequest Method Tests
// ============================================================================

// TestServerDetailsUpdateRequest_WithBasicInfo tests WithBasicInfo method
func TestServerDetailsUpdateRequest_WithBasicInfo(t *testing.T) {
	req := &ServerDetailsUpdateRequest{}
	result := req.WithBasicInfo("test-host", "192.168.1.1", "production", "US-East", "web")

	if result.Hostname != "test-host" {
		t.Errorf("Hostname = %s, want test-host", result.Hostname)
	}
	if result.MainIP != "192.168.1.1" {
		t.Errorf("MainIP = %s, want 192.168.1.1", result.MainIP)
	}
	if result.Environment != "production" {
		t.Errorf("Environment = %s, want production", result.Environment)
	}
	if result.Location != "US-East" {
		t.Errorf("Location = %s, want US-East", result.Location)
	}
	if result.Classification != "web" {
		t.Errorf("Classification = %s, want web", result.Classification)
	}

	// Verify it returns self for chaining
	if result != req {
		t.Error("WithBasicInfo should return self for chaining")
	}
}

// TestServerDetailsUpdateRequest_WithSystemInfo tests WithSystemInfo method
func TestServerDetailsUpdateRequest_WithSystemInfo(t *testing.T) {
	req := &ServerDetailsUpdateRequest{}
	result := req.WithSystemInfo("Ubuntu", "22.04", "x86_64", "SN123", "00:11:22:33:44:55")

	if result.OS != "Ubuntu" {
		t.Errorf("OS = %s, want Ubuntu", result.OS)
	}
	if result.OSVersion != "22.04" {
		t.Errorf("OSVersion = %s, want 22.04", result.OSVersion)
	}
	if result.OSArch != "x86_64" {
		t.Errorf("OSArch = %s, want x86_64", result.OSArch)
	}
	if result.SerialNumber != "SN123" {
		t.Errorf("SerialNumber = %s, want SN123", result.SerialNumber)
	}
	if result.MacAddress != "00:11:22:33:44:55" {
		t.Errorf("MacAddress = %s, want 00:11:22:33:44:55", result.MacAddress)
	}
}

// TestServerDetailsUpdateRequest_WithLegacyHardware tests WithLegacyHardware method
func TestServerDetailsUpdateRequest_WithLegacyHardware(t *testing.T) {
	req := &ServerDetailsUpdateRequest{}
	result := req.WithLegacyHardware("Intel Xeon", 2, 16, 32768, 512000)

	if result.CPUModel != "Intel Xeon" {
		t.Errorf("CPUModel = %s, want Intel Xeon", result.CPUModel)
	}
	if result.CPUCount != 2 {
		t.Errorf("CPUCount = %d, want 2", result.CPUCount)
	}
	if result.CPUCores != 16 {
		t.Errorf("CPUCores = %d, want 16", result.CPUCores)
	}
	if result.MemoryTotal != 32768 {
		t.Errorf("MemoryTotal = %d, want 32768", result.MemoryTotal)
	}
	if result.StorageTotal != 512000 {
		t.Errorf("StorageTotal = %d, want 512000", result.StorageTotal)
	}
}

// TestServerDetailsUpdateRequest_WithHardwareDetails tests WithHardwareDetails method
func TestServerDetailsUpdateRequest_WithHardwareDetails(t *testing.T) {
	req := &ServerDetailsUpdateRequest{}
	hardware := &HardwareDetails{
		CPU: []ServerCPUInfo{
			{ModelName: "Intel Xeon", PhysicalCores: 8},
		},
	}

	result := req.WithHardwareDetails(hardware)

	if result.Hardware == nil {
		t.Fatal("Hardware should not be nil")
	}
	if len(result.Hardware.CPU) != 1 {
		t.Errorf("CPU count = %d, want 1", len(result.Hardware.CPU))
	}
}

// TestServerDetailsUpdateRequest_WithCPUs tests WithCPUs method
func TestServerDetailsUpdateRequest_WithCPUs(t *testing.T) {
	req := &ServerDetailsUpdateRequest{}
	cpus := []ServerCPUInfo{
		{ModelName: "Intel Xeon E5", PhysicalCores: 8},
		{ModelName: "Intel Xeon E7", PhysicalCores: 12},
	}

	result := req.WithCPUs(cpus)

	if result.Hardware == nil {
		t.Fatal("Hardware should not be nil")
	}
	if len(result.Hardware.CPU) != 2 {
		t.Errorf("CPU count = %d, want 2", len(result.Hardware.CPU))
	}
}

// TestServerDetailsUpdateRequest_WithMemory tests WithMemory method
func TestServerDetailsUpdateRequest_WithMemory(t *testing.T) {
	req := &ServerDetailsUpdateRequest{}
	memory := &ServerMemoryInfo{
		TotalSize:  32768,
		MemoryType: "DDR4",
		Speed:      2400,
	}

	result := req.WithMemory(memory)

	if result.Hardware == nil {
		t.Fatal("Hardware should not be nil")
	}
	if result.Hardware.Memory == nil {
		t.Fatal("Memory should not be nil")
	}
	if result.Hardware.Memory.TotalSize != 32768 {
		t.Errorf("TotalSize = %d, want 32768", result.Hardware.Memory.TotalSize)
	}
}

// TestServerDetailsUpdateRequest_WithNetworkInterfaces tests WithNetworkInterfaces method
func TestServerDetailsUpdateRequest_WithNetworkInterfaces(t *testing.T) {
	req := &ServerDetailsUpdateRequest{}
	interfaces := []ServerNetworkInterfaceInfo{
		{Name: "eth0", SpeedMbps: 1000},
		{Name: "eth1", SpeedMbps: 10000},
	}

	result := req.WithNetworkInterfaces(interfaces)

	if result.Hardware == nil {
		t.Fatal("Hardware should not be nil")
	}
	if len(result.Hardware.Network) != 2 {
		t.Errorf("Network interface count = %d, want 2", len(result.Hardware.Network))
	}
}

// TestServerDetailsUpdateRequest_WithDisks tests WithDisks method
func TestServerDetailsUpdateRequest_WithDisks(t *testing.T) {
	req := &ServerDetailsUpdateRequest{}
	disks := []ServerDiskInfo{
		{Device: "/dev/sda", Size: 512000},
		{Device: "/dev/sdb", Size: 1024000},
	}

	result := req.WithDisks(disks)

	if result.Hardware == nil {
		t.Fatal("Hardware should not be nil")
	}
	if len(result.Hardware.Disks) != 2 {
		t.Errorf("Disk count = %d, want 2", len(result.Hardware.Disks))
	}
}

// TestServerDetailsUpdateRequest_HasHardwareDetails tests HasHardwareDetails method
func TestServerDetailsUpdateRequest_HasHardwareDetails(t *testing.T) {
	tests := []struct {
		name     string
		req      *ServerDetailsUpdateRequest
		expected bool
	}{
		{
			name:     "nil hardware",
			req:      &ServerDetailsUpdateRequest{},
			expected: false,
		},
		{
			name: "with hardware",
			req: &ServerDetailsUpdateRequest{
				Hardware: &HardwareDetails{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.HasHardwareDetails(); got != tt.expected {
				t.Errorf("HasHardwareDetails() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestServerDetailsUpdateRequest_HasDisks tests HasDisks method
func TestServerDetailsUpdateRequest_HasDisks(t *testing.T) {
	tests := []struct {
		name     string
		req      *ServerDetailsUpdateRequest
		expected bool
	}{
		{
			name:     "nil hardware",
			req:      &ServerDetailsUpdateRequest{},
			expected: false,
		},
		{
			name: "empty disks",
			req: &ServerDetailsUpdateRequest{
				Hardware: &HardwareDetails{},
			},
			expected: false,
		},
		{
			name: "with disks",
			req: &ServerDetailsUpdateRequest{
				Hardware: &HardwareDetails{
					Disks: []ServerDiskInfo{
						{Device: "/dev/sda"},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.HasDisks(); got != tt.expected {
				t.Errorf("HasDisks() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestServerDetailsUpdateRequest_Chaining tests method chaining
func TestServerDetailsUpdateRequest_Chaining(t *testing.T) {
	req := &ServerDetailsUpdateRequest{}

	result := req.
		WithBasicInfo("test-host", "192.168.1.1", "prod", "US", "web").
		WithSystemInfo("Ubuntu", "22.04", "x86_64", "SN123", "00:11:22:33:44:55").
		WithLegacyHardware("Intel Xeon", 2, 16, 32768, 512000).
		WithCPUs([]ServerCPUInfo{{ModelName: "Intel"}}).
		WithMemory(&ServerMemoryInfo{TotalSize: 16384}).
		WithDisks([]ServerDiskInfo{{Device: "/dev/sda"}})

	if result.Hostname != "test-host" {
		t.Error("chaining failed for basic info")
	}
	if result.OS != "Ubuntu" {
		t.Error("chaining failed for system info")
	}
	if result.CPUModel != "Intel Xeon" {
		t.Error("chaining failed for legacy hardware")
	}
	if !result.HasHardwareDetails() {
		t.Error("chaining failed for hardware details")
	}
	if !result.HasDisks() {
		t.Error("chaining failed for disks")
	}
}

// ============================================================================
// ToQuery Method Tests
// ============================================================================

// TestHardwareInventoryListOptions_ToQuery tests HardwareInventoryListOptions ToQuery method
func TestHardwareInventoryListOptions_ToQuery(t *testing.T) {
	startTime := time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 10, 31, 23, 59, 59, 0, time.UTC)

	opts := &HardwareInventoryListOptions{
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	query := opts.ToQuery()

	if query["start_time"] == "" {
		t.Error("start_time should be present in query")
	}
	if query["end_time"] == "" {
		t.Error("end_time should be present in query")
	}

	// Test with nil times
	opts2 := &HardwareInventoryListOptions{}
	query2 := opts2.ToQuery()

	if _, exists := query2["start_time"]; exists {
		t.Error("start_time should not be present when nil")
	}
	if _, exists := query2["end_time"]; exists {
		t.Error("end_time should not be present when nil")
	}
}

// TestTagNamespaceListOptions_ToQuery tests TagNamespaceListOptions ToQuery method
func TestTagNamespaceListOptions_ToQuery(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name     string
		opts     TagNamespaceListOptions
		wantKeys []string
	}{
		{
			name:     "all fields",
			opts:     TagNamespaceListOptions{Type: "custom", Parent: "root", Active: &trueVal, Search: "test"},
			wantKeys: []string{"type", "parent", "active", "search"},
		},
		{
			name:     "active false",
			opts:     TagNamespaceListOptions{Active: &falseVal},
			wantKeys: []string{"active"},
		},
		{
			name:     "empty options",
			opts:     TagNamespaceListOptions{},
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.opts.ToQuery()

			for _, key := range tt.wantKeys {
				if _, exists := query[key]; !exists {
					t.Errorf("expected key %s to exist in query", key)
				}
			}

			// Check active value if set
			if tt.opts.Active != nil {
				expectedValue := "false"
				if *tt.opts.Active {
					expectedValue = "true"
				}
				if query["active"] != expectedValue {
					t.Errorf("active = %s, want %s", query["active"], expectedValue)
				}
			}
		})
	}
}

// ============================================================================
// NotificationPriority Tests
// ============================================================================

// TestNotificationPriority_String tests NotificationPriority String method
func TestNotificationPriority_String(t *testing.T) {
	tests := []struct {
		priority NotificationPriority
		expected string
	}{
		{NotificationPriorityLow, "low"},
		{NotificationPriorityNormal, "normal"},
		{NotificationPriorityHigh, "high"},
		{NotificationPriorityCritical, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.priority.String(); got != tt.expected {
				t.Errorf("String() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// Additional Model Tests
// ============================================================================

// TestAlert_JSON tests Alert model serialization
func TestAlert_JSON(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	alert := Alert{
		GormModel: GormModel{
			ID:        1,
			CreatedAt: &now,
		},
		Name:        "CPU Alert",
		Description: "CPU usage too high",
		Type:        "metric",
		MetricName:  "cpu_usage",
		Condition:   "greater_than",
		Threshold:   80.0,
		Duration:    300,
		Enabled:     true,
		Status:      "active",
		Severity:    "critical",
		Channels:    []string{"email", "slack"},
		Recipients:  []string{"admin@example.com"},
		Tags:        []string{"production", "critical"},
	}

	data, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("failed to marshal Alert: %v", err)
	}

	var decoded Alert
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Alert: %v", err)
	}

	if decoded.Name != alert.Name {
		t.Errorf("Name = %s, want %s", decoded.Name, alert.Name)
	}
	if decoded.Threshold != alert.Threshold {
		t.Errorf("Threshold = %f, want %f", decoded.Threshold, alert.Threshold)
	}
}

// TestMetric_JSON tests Metric model serialization
func TestMetric_JSON(t *testing.T) {
	metric := Metric{
		ServerUUID: "server-123",
		Timestamp:  time.Now().UTC(),
		Name:       "cpu_usage",
		Value:      75.5,
		Unit:       "percent",
		Tags: map[string]string{
			"host": "web-01",
			"env":  "production",
		},
	}

	data, err := json.Marshal(metric)
	if err != nil {
		t.Fatalf("failed to marshal Metric: %v", err)
	}

	var decoded Metric
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Metric: %v", err)
	}

	if decoded.Name != metric.Name {
		t.Errorf("Name = %s, want %s", decoded.Name, metric.Name)
	}
	if decoded.Value != metric.Value {
		t.Errorf("Value = %f, want %f", decoded.Value, metric.Value)
	}
}

// TestServerCreateRequest_JSON tests ServerCreateRequest serialization
func TestServerCreateRequest_JSON(t *testing.T) {
	req := ServerCreateRequest{
		Hostname:       "test-server",
		MainIP:         "192.168.1.100",
		OS:             "Ubuntu",
		OSVersion:      "22.04",
		OSArch:         "x86_64",
		SerialNumber:   "SN123456",
		MacAddress:     "00:11:22:33:44:55",
		Environment:    "production",
		Location:       "US-East",
		Classification: "web",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal ServerCreateRequest: %v", err)
	}

	var decoded ServerCreateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ServerCreateRequest: %v", err)
	}

	if decoded.Hostname != req.Hostname {
		t.Errorf("Hostname = %s, want %s", decoded.Hostname, req.Hostname)
	}
	if decoded.MainIP != req.MainIP {
		t.Errorf("MainIP = %s, want %s", decoded.MainIP, req.MainIP)
	}
}

// TestMonitoringAgent_JSON tests MonitoringAgent serialization
func TestMonitoringAgent_JSON(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	agent := MonitoringAgent{
		GormModel: GormModel{
			ID:        1,
			CreatedAt: &now,
		},
		UUID:           "agent-uuid-123",
		Name:           "Test Agent",
		Status:         "active",
		Version:        "1.0.0",
		OrganizationID: 1,
		ServerUUID:     "server-uuid-123",
		Configuration: map[string]interface{}{
			"interval": 60,
			"enabled":  true,
		},
	}

	data, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("failed to marshal MonitoringAgent: %v", err)
	}

	var decoded MonitoringAgent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal MonitoringAgent: %v", err)
	}

	if decoded.UUID != agent.UUID {
		t.Errorf("UUID = %s, want %s", decoded.UUID, agent.UUID)
	}
	if decoded.Status != agent.Status {
		t.Errorf("Status = %s, want %s", decoded.Status, agent.Status)
	}
}

// ============================================================================
// Constructor Function Tests
// ============================================================================

// TestConstructor_NewUserAPIKey tests the NewUserAPIKey constructor
func TestConstructor_NewUserAPIKey(t *testing.T) {
	req := NewUserAPIKey("Test Key", "Test Description", []string{"read", "write"})

	if req.Name != "Test Key" {
		t.Errorf("Name = %s, want Test Key", req.Name)
	}
	if req.Description != "Test Description" {
		t.Errorf("Description = %s, want Test Description", req.Description)
	}
	if req.Type != APIKeyTypeUser {
		t.Errorf("Type = %s, want %s", req.Type, APIKeyTypeUser)
	}
	if len(req.Capabilities) != 2 {
		t.Errorf("Capabilities length = %d, want 2", len(req.Capabilities))
	}
}

// TestConstructor_NewAdminAPIKey tests the NewAdminAPIKey constructor
func TestConstructor_NewAdminAPIKey(t *testing.T) {
	req := NewAdminAPIKey("Admin Key", "Admin Description", []string{"*"}, 123)

	if req.Name != "Admin Key" {
		t.Errorf("Name = %s, want Admin Key", req.Name)
	}
	if req.Type != APIKeyTypeAdmin {
		t.Errorf("Type = %s, want %s", req.Type, APIKeyTypeAdmin)
	}
	if req.OrganizationID != 123 {
		t.Errorf("OrganizationID = %d, want 123", req.OrganizationID)
	}
	if len(req.Capabilities) != 1 || req.Capabilities[0] != "*" {
		t.Errorf("Capabilities = %v, want [*]", req.Capabilities)
	}
}

// TestConstructor_NewControllerKey tests the NewControllerKey constructor
func TestConstructor_NewControllerKey(t *testing.T) {
	req := NewControllerKey("Controller Key", "Controller Desc", []string{"controller:read"}, 456)

	if req.Name != "Controller Key" {
		t.Errorf("Name = %s, want Controller Key", req.Name)
	}
	if req.Type != APIKeyTypeController {
		t.Errorf("Type = %s, want %s", req.Type, APIKeyTypeController)
	}
	if req.OrganizationID != 456 {
		t.Errorf("OrganizationID = %d, want 456", req.OrganizationID)
	}
}

// TestConstructor_NewMonitoringAgentKey tests the NewMonitoringAgentKey constructor
func TestConstructor_NewMonitoringAgentKey(t *testing.T) {
	req := NewMonitoringAgentKey("Agent Key", "Agent Desc", "test-namespace", "public", "us-east-1", []string{"scope1", "scope2"})

	if req.Name != "Agent Key" {
		t.Errorf("Name = %s, want Agent Key", req.Name)
	}
	if req.Type != APIKeyTypeMonitoringAgent {
		t.Errorf("Type = %s, want %s", req.Type, APIKeyTypeMonitoringAgent)
	}
	if req.NamespaceName != "test-namespace" {
		t.Errorf("NamespaceName = %s, want test-namespace", req.NamespaceName)
	}
	if req.AgentType != "public" {
		t.Errorf("AgentType = %s, want public", req.AgentType)
	}
	if req.RegionCode != "us-east-1" {
		t.Errorf("RegionCode = %s, want us-east-1", req.RegionCode)
	}
	if len(req.AllowedProbeScopes) != 2 {
		t.Errorf("AllowedProbeScopes length = %d, want 2", len(req.AllowedProbeScopes))
	}
	// Verify default capabilities are set
	hasMonitoring := false
	hasProbes := false
	for _, cap := range req.Capabilities {
		if cap == "monitoring:execute" {
			hasMonitoring = true
		}
		if cap == "probes:execute" {
			hasProbes = true
		}
	}
	if !hasMonitoring || !hasProbes {
		t.Errorf("Capabilities = %v, missing monitoring:execute or probes:execute", req.Capabilities)
	}
}

// TestConstructor_NewRegistrationKey tests the NewRegistrationKey constructor
func TestConstructor_NewRegistrationKey(t *testing.T) {
	req := NewRegistrationKey("Registration Key", "Registration Desc", 789)

	if req.Name != "Registration Key" {
		t.Errorf("Name = %s, want Registration Key", req.Name)
	}
	if req.Type != APIKeyTypeRegistration {
		t.Errorf("Type = %s, want %s", req.Type, APIKeyTypeRegistration)
	}
	if req.OrganizationID != 789 {
		t.Errorf("OrganizationID = %d, want 789", req.OrganizationID)
	}
	// Verify default capabilities
	hasRegister := false
	hasUpdate := false
	for _, cap := range req.Capabilities {
		if cap == "servers:register" {
			hasRegister = true
		}
		if cap == "servers:update" {
			hasUpdate = true
		}
	}
	if !hasRegister || !hasUpdate {
		t.Errorf("Capabilities = %v, missing servers:register or servers:update", req.Capabilities)
	}
}

// ============================================================================
// ToQuery Method Tests (Note: Tag-related ToQuery tests are in tags_test.go and model_methods_coverage_test.go)
// ============================================================================

// Note: The following ToQuery tests are already covered in other test files:
// - TestTagListOptions_ToQuery -> tags_test.go
// - TestOrganizationTagListOptions_ToQuery -> model_methods_coverage_test.go
// - TestServerRelationshipListOptions_ToQuery -> model_methods_coverage_test.go
// - TestTagHistoryQueryParams_ToQuery -> tags_test.go
// - TestTagDetectionRuleListOptions_ToQuery -> tags_test.go

// ============================================================================
// JSON Validation Edge Cases
// ============================================================================

// TestJSON_MalformedInput tests unmarshaling malformed JSON
func TestJSON_MalformedInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		model interface{}
	}{
		{"missing closing brace", `{"name":"test"`, &Organization{}},
		{"invalid comma", `{"name":"test",,}`, &User{}},
		{"trailing comma", `{"name":"test",}`, &Server{}},
		{"single quotes", `{'name':'test'}`, &Organization{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tt.input), tt.model)
			if err == nil {
				t.Error("expected error for malformed JSON, got nil")
			}
		})
	}
}

// TestJSON_TypeMismatches tests type mismatch handling
func TestJSON_TypeMismatches(t *testing.T) {
	tests := []struct {
		name  string
		input string
		model interface{}
	}{
		{
			name:  "string for int",
			input: `{"id":"not-a-number","name":"test"}`,
			model: &Organization{},
		},
		{
			name:  "int for string",
			input: `{"uuid":12345,"name":"test"}`,
			model: &Organization{},
		},
		{
			name:  "string for bool",
			input: `{"alerts_enabled":"yes"}`,
			model: &Organization{},
		},
		{
			name:  "int for bool",
			input: `{"is_active":1}`,
			model: &User{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tt.input), tt.model)
			if err == nil {
				t.Error("expected error for type mismatch, got nil")
			}
		})
	}
}

// TestJSON_EmptyAndNullArrays tests empty vs nil array handling
func TestJSON_EmptyAndNullArrays(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectNil   bool
		expectEmpty bool
	}{
		{
			name:        "null array",
			input:       `{"uuid":"test","name":"Test","tags":null}`,
			expectNil:   true,
			expectEmpty: false,
		},
		{
			name:        "empty array",
			input:       `{"uuid":"test","name":"Test","tags":[]}`,
			expectNil:   false,
			expectEmpty: true,
		},
		{
			name:        "missing array field",
			input:       `{"uuid":"test","name":"Test"}`,
			expectNil:   true,
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var org Organization
			if err := json.Unmarshal([]byte(tt.input), &org); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if tt.expectNil {
				if org.Tags != nil {
					t.Errorf("expected nil Tags, got %v", org.Tags)
				}
			}

			if tt.expectEmpty {
				if org.Tags == nil {
					t.Error("expected non-nil empty Tags, got nil")
				} else if len(org.Tags) != 0 {
					t.Errorf("expected empty Tags, got length %d", len(org.Tags))
				}
			}
		})
	}
}

// TestJSON_EmptyAndNullMaps tests empty vs nil map handling
func TestJSON_EmptyAndNullMaps(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectNil   bool
		expectEmpty bool
	}{
		{
			name:        "null map",
			input:       `{"uuid":"test","name":"Test","settings":null}`,
			expectNil:   true,
			expectEmpty: false,
		},
		{
			name:        "empty map",
			input:       `{"uuid":"test","name":"Test","settings":{}}`,
			expectNil:   false,
			expectEmpty: true,
		},
		{
			name:        "missing map field",
			input:       `{"uuid":"test","name":"Test"}`,
			expectNil:   true,
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var org Organization
			if err := json.Unmarshal([]byte(tt.input), &org); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if tt.expectNil {
				if org.Settings != nil {
					t.Errorf("expected nil Settings, got %v", org.Settings)
				}
			}

			if tt.expectEmpty {
				if org.Settings == nil {
					t.Error("expected non-nil empty Settings, got nil")
				} else if len(org.Settings) != 0 {
					t.Errorf("expected empty Settings, got length %d", len(org.Settings))
				}
			}
		})
	}
}

// TestJSON_NegativeValuesForUint tests negative values for uint fields
func TestJSON_NegativeValuesForUint(t *testing.T) {
	input := `{"uuid":"test","name":"Test","max_servers":-10}`

	var org Organization
	err := json.Unmarshal([]byte(input), &org)
	if err == nil {
		t.Error("expected error for negative value in uint field, got nil")
	}
}

// ============================================================================
// Additional Domain Model Tests
// ============================================================================

// TestProbe_JSON tests Probe model serialization
func TestProbe_JSON(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	probe := Probe{
		GormModel: GormModel{
			ID:        1,
			CreatedAt: &now,
		},
		ProbeUUID:      "probe-uuid-123",
		Name:           "HTTP Health Check",
		Description:    "Checks HTTP endpoint health",
		Type:           "http",
		Target:         "https://example.com",
		Interval:       60,
		Timeout:        10,
		Enabled:        true,
		OrganizationID: 1,
		Regions:        []string{"us-east-1", "us-west-2"},
	}

	data, err := json.Marshal(probe)
	if err != nil {
		t.Fatalf("failed to marshal Probe: %v", err)
	}

	var decoded Probe
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Probe: %v", err)
	}

	if decoded.ProbeUUID != probe.ProbeUUID {
		t.Errorf("ProbeUUID = %s, want %s", decoded.ProbeUUID, probe.ProbeUUID)
	}
	if decoded.Name != probe.Name {
		t.Errorf("Name = %s, want %s", decoded.Name, probe.Name)
	}
	if decoded.Type != probe.Type {
		t.Errorf("Type = %s, want %s", decoded.Type, probe.Type)
	}
	if decoded.Interval != probe.Interval {
		t.Errorf("Interval = %d, want %d", decoded.Interval, probe.Interval)
	}
}

// TestCPUMetrics_JSON tests CPUMetrics model serialization
func TestCPUMetrics_JSON(t *testing.T) {
	metrics := CPUMetrics{
		UsagePercent:  75.5,
		UserPercent:   45.2,
		SystemPercent: 30.3,
		IdlePercent:   24.5,
		IOWaitPercent: 5.0,
		StealPercent:  0.0,
		LoadAverage1:  1.5,
		LoadAverage5:  2.0,
		LoadAverage15: 2.5,
		CoreCount:     8,
		ThreadCount:   16,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("failed to marshal CPUMetrics: %v", err)
	}

	var decoded CPUMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CPUMetrics: %v", err)
	}

	if decoded.UsagePercent != metrics.UsagePercent {
		t.Errorf("UsagePercent = %f, want %f", decoded.UsagePercent, metrics.UsagePercent)
	}
	if decoded.CoreCount != metrics.CoreCount {
		t.Errorf("CoreCount = %d, want %d", decoded.CoreCount, metrics.CoreCount)
	}
}

// TestMemoryMetrics_JSON tests MemoryMetrics model serialization
func TestMemoryMetrics_JSON(t *testing.T) {
	metrics := MemoryMetrics{
		TotalBytes:       34359738368,
		UsedBytes:        17179869184,
		FreeBytes:        17179869184,
		UsagePercent:     50.0,
		AvailableBytes:   20000000000,
		BuffersBytes:     1000000000,
		CachedBytes:      2000000000,
		SwapTotalBytes:   8589934592,
		SwapUsedBytes:    4294967296,
		SwapFreeBytes:    4294967296,
		SwapUsagePercent: 50.0,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("failed to marshal MemoryMetrics: %v", err)
	}

	var decoded MemoryMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal MemoryMetrics: %v", err)
	}

	if decoded.TotalBytes != metrics.TotalBytes {
		t.Errorf("TotalBytes = %d, want %d", decoded.TotalBytes, metrics.TotalBytes)
	}
	if decoded.UsagePercent != metrics.UsagePercent {
		t.Errorf("UsagePercent = %f, want %f", decoded.UsagePercent, metrics.UsagePercent)
	}
}

// TestDiskMetrics_JSON tests DiskMetrics model serialization
func TestDiskMetrics_JSON(t *testing.T) {
	metrics := DiskMetrics{
		Device:             "/dev/sda1",
		Mountpoint:         "/",
		Filesystem:         "ext4",
		TotalBytes:         1099511627776,
		UsedBytes:          549755813888,
		FreeBytes:          549755813888,
		UsagePercent:       50.0,
		InodesTotal:        65536000,
		InodesUsed:         32768000,
		InodesFree:         32768000,
		InodesUsagePercent: 50.0,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("failed to marshal DiskMetrics: %v", err)
	}

	var decoded DiskMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal DiskMetrics: %v", err)
	}

	if decoded.Device != metrics.Device {
		t.Errorf("Device = %s, want %s", decoded.Device, metrics.Device)
	}
	if decoded.Filesystem != metrics.Filesystem {
		t.Errorf("Filesystem = %s, want %s", decoded.Filesystem, metrics.Filesystem)
	}
	if decoded.UsagePercent != metrics.UsagePercent {
		t.Errorf("UsagePercent = %f, want %f", decoded.UsagePercent, metrics.UsagePercent)
	}
}

// TestNetworkMetrics_JSON tests NetworkMetrics model serialization
func TestNetworkMetrics_JSON(t *testing.T) {
	metrics := NetworkMetrics{
		Interface:   "eth0",
		BytesSent:   1073741824,
		BytesRecv:   2147483648,
		PacketsSent: 1000000,
		PacketsRecv: 2000000,
		ErrorsIn:    100,
		ErrorsOut:   50,
		DropsIn:     10,
		DropsOut:    5,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("failed to marshal NetworkMetrics: %v", err)
	}

	var decoded NetworkMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal NetworkMetrics: %v", err)
	}

	if decoded.Interface != metrics.Interface {
		t.Errorf("Interface = %s, want %s", decoded.Interface, metrics.Interface)
	}
	if decoded.BytesSent != metrics.BytesSent {
		t.Errorf("BytesSent = %d, want %d", decoded.BytesSent, metrics.BytesSent)
	}
}

// TestProcessMetrics_JSON tests ProcessMetrics model serialization
func TestProcessMetrics_JSON(t *testing.T) {
	metrics := ProcessMetrics{
		PID:           1234,
		Name:          "nginx",
		Username:      "www-data",
		State:         "running",
		CPUPercent:    15.5,
		MemoryPercent: 5.2,
		MemoryRSS:     104857600,
		MemoryVMS:     209715200,
		CreateTime:    1234567890,
		OpenFiles:     100,
		NumThreads:    4,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("failed to marshal ProcessMetrics: %v", err)
	}

	var decoded ProcessMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ProcessMetrics: %v", err)
	}

	if decoded.PID != metrics.PID {
		t.Errorf("PID = %d, want %d", decoded.PID, metrics.PID)
	}
	if decoded.Name != metrics.Name {
		t.Errorf("Name = %s, want %s", decoded.Name, metrics.Name)
	}
}

// TestTemperatureMetrics_JSON tests TemperatureMetrics model serialization
func TestTemperatureMetrics_JSON(t *testing.T) {
	metrics := TemperatureMetrics{
		Sensors: []TemperatureSensorData{
			{
				SensorID:      "cpu_sensor_1",
				SensorName:    "CPU",
				Temperature:   65.5,
				Status:        "ok",
				Type:          "cpu",
				UpperWarning:  80.0,
				UpperCritical: 90.0,
			},
			{
				SensorID:      "gpu_sensor_1",
				SensorName:    "GPU",
				Temperature:   70.2,
				Status:        "ok",
				Type:          "gpu",
				UpperWarning:  85.0,
				UpperCritical: 95.0,
			},
		},
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("failed to marshal TemperatureMetrics: %v", err)
	}

	var decoded TemperatureMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal TemperatureMetrics: %v", err)
	}

	if len(decoded.Sensors) != len(metrics.Sensors) {
		t.Errorf("Sensors length = %d, want %d", len(decoded.Sensors), len(metrics.Sensors))
	}
	if len(decoded.Sensors) > 0 {
		if decoded.Sensors[0].SensorName != metrics.Sensors[0].SensorName {
			t.Errorf("Sensors[0].SensorName = %s, want %s", decoded.Sensors[0].SensorName, metrics.Sensors[0].SensorName)
		}
	}
}

// TestPowerMetrics_JSON tests PowerMetrics model serialization
func TestPowerMetrics_JSON(t *testing.T) {
	metrics := PowerMetrics{
		PowerSupplies: []PowerSupplyMetrics{
			{
				ID:            "psu1",
				Name:          "PSU1",
				Status:        "ok",
				PowerWatts:    450.5,
				MaxPowerWatts: 650.0,
				Voltage:       12.0,
				Current:       37.5,
				Efficiency:    92.5,
				Temperature:   45.0,
			},
		},
		TotalPowerW: 450.5,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("failed to marshal PowerMetrics: %v", err)
	}

	var decoded PowerMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal PowerMetrics: %v", err)
	}

	if decoded.TotalPowerW != metrics.TotalPowerW {
		t.Errorf("TotalPowerW = %f, want %f", decoded.TotalPowerW, metrics.TotalPowerW)
	}
	if len(decoded.PowerSupplies) != len(metrics.PowerSupplies) {
		t.Errorf("PowerSupplies length = %d, want %d", len(decoded.PowerSupplies), len(metrics.PowerSupplies))
	}
}

// TestServiceMetrics_JSON tests ServiceMetrics model serialization
func TestServiceMetrics_JSON(t *testing.T) {
	metrics := ServiceMetrics{
		ServiceName:  "nginx",
		Timestamp:    time.Now().UTC(),
		CPUPercent:   25.5,
		MemoryRSS:    524288000,
		ProcessCount: 3,
		ThreadCount:  8,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("failed to marshal ServiceMetrics: %v", err)
	}

	var decoded ServiceMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ServiceMetrics: %v", err)
	}

	if decoded.ServiceName != metrics.ServiceName {
		t.Errorf("ServiceName = %s, want %s", decoded.ServiceName, metrics.ServiceName)
	}
	if decoded.CPUPercent != metrics.CPUPercent {
		t.Errorf("CPUPercent = %f, want %f", decoded.CPUPercent, metrics.CPUPercent)
	}
}

// TestIncident_JSON tests Incident model serialization
func TestIncident_JSON(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	incident := Incident{
		GormModel: GormModel{
			ID:        1,
			CreatedAt: &now,
		},
		Title:          "Service Outage",
		Description:    "Database connection failed",
		Severity:       "critical",
		Status:         "investigating",
		Source:         "monitoring",
		OrganizationID: 1,
		StartedAt:      &now,
	}

	data, err := json.Marshal(incident)
	if err != nil {
		t.Fatalf("failed to marshal Incident: %v", err)
	}

	var decoded Incident
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Incident: %v", err)
	}

	if decoded.Title != incident.Title {
		t.Errorf("Title = %s, want %s", decoded.Title, incident.Title)
	}
	if decoded.Severity != incident.Severity {
		t.Errorf("Severity = %s, want %s", decoded.Severity, incident.Severity)
	}
}

// TestHardwareDetails_JSON tests HardwareDetails model serialization with nested structures
func TestHardwareDetails_JSON(t *testing.T) {
	hardware := HardwareDetails{
		CPU: []ServerCPUInfo{
			{
				ModelName:     "Intel Xeon E5",
				PhysicalCores: 8,
				LogicalCores:  16,
				BaseSpeed:     2400.0,
			},
		},
		Memory: &ServerMemoryInfo{
			TotalSize:  32768,
			MemoryType: "DDR4",
			Speed:      2400,
		},
		Disks: []ServerDiskInfo{
			{
				Device: "/dev/sda",
				Size:   512000,
				Type:   "SSD",
			},
		},
		Network: []ServerNetworkInterfaceInfo{
			{
				Name:      "eth0",
				SpeedMbps: 1000,
			},
		},
	}

	data, err := json.Marshal(hardware)
	if err != nil {
		t.Fatalf("failed to marshal HardwareDetails: %v", err)
	}

	var decoded HardwareDetails
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal HardwareDetails: %v", err)
	}

	if len(decoded.CPU) != len(hardware.CPU) {
		t.Errorf("CPU count = %d, want %d", len(decoded.CPU), len(hardware.CPU))
	}
	if decoded.Memory == nil {
		t.Fatal("Memory should not be nil")
	}
	if decoded.Memory.TotalSize != hardware.Memory.TotalSize {
		t.Errorf("Memory.TotalSize = %d, want %d", decoded.Memory.TotalSize, hardware.Memory.TotalSize)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
