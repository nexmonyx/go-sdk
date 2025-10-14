package nexmonyx

import (
	"encoding/json"
	"testing"
	"time"
)

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

// TestOrganization_JSON tests Organization model serialization
func TestOrganization_JSON(t *testing.T) {
	now := CustomTime{Time: time.Now().UTC()}
	org := Organization{
		GormModel: GormModel{
			ID:        1,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
		UUID:        "org-uuid-123",
		Name:        "Test Organization",
		Description: "Test description",
		Industry:    "Technology",
		Website:     "https://example.com",
		Size:        "50-200",
		Country:     "US",
		TimeZone:    "America/New_York",
		MaxServers:  100,
		MaxUsers:    50,
		MaxProbes:   200,
		AlertsEnabled:     true,
		MonitoringEnabled: true,
		Tags:        []string{"test", "demo"},
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
		Auth0ID:          "auth0|123",
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
			"theme":        "dark",
			"notifications": true,
			"max_retries":  3,
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
