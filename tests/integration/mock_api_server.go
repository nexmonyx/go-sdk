package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// MockAPIServer provides a test HTTP server that mimics the Nexmonyx API
type MockAPIServer struct {
	Server      *httptest.Server
	t           *testing.T
	mu          sync.RWMutex
	servers     map[string]interface{}
	orgs        map[string]interface{}
	alerts      map[string]interface{}
	users       map[string]interface{}
	probes      map[string]interface{}
	nextID      int
	authEnabled bool
	authToken   string
}

// NewMockAPIServer creates and starts a new mock API server
func NewMockAPIServer(t *testing.T) *MockAPIServer {
	mock := &MockAPIServer{
		t:           t,
		servers:     make(map[string]interface{}),
		orgs:        make(map[string]interface{}),
		alerts:      make(map[string]interface{}),
		users:       make(map[string]interface{}),
		probes:      make(map[string]interface{}),
		nextID:      1000,
		authEnabled: true,
		authToken:   "test-token",
	}

	// Load fixtures into state
	mock.loadFixtures()

	// Create HTTP server with routes
	mux := http.NewServeMux()
	mock.registerRoutes(mux)
	mock.Server = httptest.NewServer(mux)

	t.Logf("Mock API server started at %s", mock.Server.URL)
	return mock
}

// Close stops the mock server
func (m *MockAPIServer) Close() {
	m.Server.Close()
	m.t.Log("Mock API server stopped")
}

// SetAuthToken sets the expected authentication token
func (m *MockAPIServer) SetAuthToken(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.authToken = token
}

// DisableAuth disables authentication checking
func (m *MockAPIServer) DisableAuth() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.authEnabled = false
}

// loadFixtures loads test data from fixtures
func (m *MockAPIServer) loadFixtures() {
	// Load servers
	servers := loadFixture(m.t, "servers.json")
	if serverList, ok := servers.([]interface{}); ok {
		for _, s := range serverList {
			if server, ok := s.(map[string]interface{}); ok {
				if uuid, ok := server["server_uuid"].(string); ok {
					m.servers[uuid] = server
				}
			}
		}
	}

	// Load organizations
	orgs := loadFixture(m.t, "organizations.json")
	if orgList, ok := orgs.([]interface{}); ok {
		for _, o := range orgList {
			if org, ok := o.(map[string]interface{}); ok {
				if uuid, ok := org["uuid"].(string); ok {
					m.orgs[uuid] = org
				}
			}
		}
	}

	// Load alerts
	alerts := loadFixture(m.t, "alerts.json")
	if alertList, ok := alerts.([]interface{}); ok {
		for _, a := range alertList {
			if alert, ok := a.(map[string]interface{}); ok {
				if uuid, ok := alert["uuid"].(string); ok {
					m.alerts[uuid] = alert
				}
			}
		}
	}

	// Load users
	users := loadFixture(m.t, "users.json")
	if userList, ok := users.([]interface{}); ok {
		for _, u := range userList {
			if user, ok := u.(map[string]interface{}); ok {
				if uuid, ok := user["uuid"].(string); ok {
					m.users[uuid] = user
				}
			}
		}
	}

	// Load probes
	probes := loadFixture(m.t, "probes.json")
	if probeList, ok := probes.([]interface{}); ok {
		for _, p := range probeList {
			if probe, ok := p.(map[string]interface{}); ok {
				if uuid, ok := probe["probe_uuid"].(string); ok {
					m.probes[uuid] = probe
				}
			}
		}
	}

	m.t.Logf("Loaded fixtures: %d servers, %d orgs, %d alerts, %d users, %d probes",
		len(m.servers), len(m.orgs), len(m.alerts), len(m.users), len(m.probes))
}

// registerRoutes sets up all API routes
func (m *MockAPIServer) registerRoutes(mux *http.ServeMux) {
	// Server routes (support both v1 and v2)
	mux.HandleFunc("/v1/register", m.handleServersCreate)      // Server registration endpoint
	mux.HandleFunc("/v1/server/", m.handleServerByID)          // Server details endpoint (singular)
	mux.HandleFunc("/v1/admin/server/", m.handleServerByID)    // Admin delete endpoint
	mux.HandleFunc("/v1/servers", m.handleServers)
	mux.HandleFunc("/v1/servers/", m.handleServerByID)
	mux.HandleFunc("/v2/servers", m.handleServers)
	mux.HandleFunc("/v2/servers/", m.handleServerByID)

	// Organization routes (support both v1 and v2)
	mux.HandleFunc("/v1/organizations", m.handleOrganizations)
	mux.HandleFunc("/v1/organizations/", m.handleOrganizationByID)
	mux.HandleFunc("/v2/organizations", m.handleOrganizations)
	mux.HandleFunc("/v2/organizations/", m.handleOrganizationByID)

	// Metrics routes
	mux.HandleFunc("/v1/metrics/submit", m.handleMetricsSubmit)
	mux.HandleFunc("/v1/metrics", m.handleMetricsQuery)
	mux.HandleFunc("/v2/metrics/submit", m.handleMetricsSubmit)
	mux.HandleFunc("/v2/metrics", m.handleMetricsQuery)

	// Alert routes (support both v1 and v2)
	mux.HandleFunc("/v1/alerts/rules", m.handleAlerts)
	mux.HandleFunc("/v1/alerts/rules/", m.handleAlertByID)
	mux.HandleFunc("/v1/alerts/", m.handleAlertActions)
	mux.HandleFunc("/v2/alerts", m.handleAlerts)
	mux.HandleFunc("/v2/alerts/", m.handleAlertByID)

	// Monitoring routes
	mux.HandleFunc("/v1/monitoring/probes", m.handleProbes)
	mux.HandleFunc("/v1/monitoring/probes/", m.handleProbeByID)
	mux.HandleFunc("/v2/monitoring/probes", m.handleProbes)
	mux.HandleFunc("/v2/monitoring/probes/", m.handleProbeByID)

	// System routes
	mux.HandleFunc("/v1/system/health", m.handleHealth)
	mux.HandleFunc("/v1/system/version", m.handleVersion)
	mux.HandleFunc("/v2/system/health", m.handleHealth)
	mux.HandleFunc("/v2/system/version", m.handleVersion)

	// Catch-all for unimplemented routes
	mux.HandleFunc("/", m.handleNotFound)
}

// Middleware to check authentication
func (m *MockAPIServer) checkAuth(r *http.Request) error {
	if !m.authEnabled {
		return nil
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return fmt.Errorf("missing authorization header")
	}

	expectedAuth := "Bearer " + m.authToken
	if auth != expectedAuth {
		return fmt.Errorf("invalid authorization token")
	}

	return nil
}

// handleServers handles /v2/servers (GET, POST)
func (m *MockAPIServer) handleServers(w http.ResponseWriter, r *http.Request) {
	if err := m.checkAuth(r); err != nil {
		m.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		m.handleServersList(w, r)
	case http.MethodPost:
		m.handleServersCreate(w, r)
	default:
		m.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleServersList returns list of servers
func (m *MockAPIServer) handleServersList(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	serverList := make([]interface{}, 0, len(m.servers))
	for _, server := range m.servers {
		serverList = append(serverList, server)
	}

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": serverList,
		"meta": map[string]interface{}{
			"total_items":  len(serverList),
			"page":         1,
			"limit":        25,
			"total_pages":  1,
			"has_more":     false,
			"first_page":   1,
			"last_page":    1,
			"from":         1,
			"to":           len(serverList),
			"per_page":     25,
			"current_page": 1,
		},
	})
}

// handleServersCreate creates a new server
func (m *MockAPIServer) handleServersCreate(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	hostname, ok := req["hostname"].(string)
	if !ok || hostname == "" {
		m.writeError(w, http.StatusBadRequest, "hostname is required")
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate new server
	uuid := fmt.Sprintf("server-%d", m.nextID)
	m.nextID++

	now := time.Now().Format(time.RFC3339)
	server := map[string]interface{}{
		"id":              m.nextID,
		"server_uuid":     uuid,
		"hostname":        hostname,
		"organization_id": req["organization_id"],
		"main_ip":         req["main_ip"],
		"location":        req["location"],
		"environment":     req["environment"],
		"classification":  req["classification"],
		"status":          "active",
		"agent_version":   "1.5.0",
		"created_at":      now,
		"updated_at":      now,
	}

	m.servers[uuid] = server

	m.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data": server,
	})
}

// handleServerByID handles /v1/servers/{uuid} and /v2/servers/{uuid} (GET, PUT, DELETE)
func (m *MockAPIServer) handleServerByID(w http.ResponseWriter, r *http.Request) {
	if err := m.checkAuth(r); err != nil {
		m.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Extract UUID from path (support both v1 and v2, singular and plural, admin endpoints)
	uuid := r.URL.Path
	uuid = strings.TrimPrefix(uuid, "/v1/admin/server/")  // Admin delete endpoint
	uuid = strings.TrimPrefix(uuid, "/v1/server/")        // Singular form (used by GetByUUID)
	uuid = strings.TrimPrefix(uuid, "/v1/servers/")       // Plural form
	uuid = strings.TrimPrefix(uuid, "/v2/servers/")       // V2 plural form
	uuid = strings.TrimSuffix(uuid, "/details")           // Strip /details suffix if present
	uuid = strings.TrimSuffix(uuid, "/")                  // Strip trailing slash
	if uuid == "" {
		m.writeError(w, http.StatusBadRequest, "Server UUID required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		m.handleServerGet(w, r, uuid)
	case http.MethodPut, http.MethodPatch:
		m.handleServerUpdate(w, r, uuid)
	case http.MethodDelete:
		m.handleServerDelete(w, r, uuid)
	default:
		m.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleServerGet returns a single server
func (m *MockAPIServer) handleServerGet(w http.ResponseWriter, r *http.Request, uuid string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, ok := m.servers[uuid]
	if !ok {
		m.writeError(w, http.StatusNotFound, "Server not found")
		return
	}

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": server,
	})
}

// handleServerUpdate updates a server
func (m *MockAPIServer) handleServerUpdate(w http.ResponseWriter, r *http.Request, uuid string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, ok := m.servers[uuid]
	if !ok {
		m.writeError(w, http.StatusNotFound, "Server not found")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		m.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update server fields
	serverMap := server.(map[string]interface{})
	for key, value := range updates {
		if key != "uuid" && key != "created_at" {
			serverMap[key] = value
		}
	}
	serverMap["updated_at"] = time.Now().Format(time.RFC3339)

	m.servers[uuid] = serverMap

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": serverMap,
	})
}

// handleServerDelete deletes a server
func (m *MockAPIServer) handleServerDelete(w http.ResponseWriter, r *http.Request, uuid string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.servers[uuid]; !ok {
		m.writeError(w, http.StatusNotFound, "Server not found")
		return
	}

	delete(m.servers, uuid)
	w.WriteHeader(http.StatusNoContent)
}

// handleOrganizations handles /v2/organizations (GET, POST)
func (m *MockAPIServer) handleOrganizations(w http.ResponseWriter, r *http.Request) {
	if err := m.checkAuth(r); err != nil {
		m.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		m.handleOrganizationsList(w, r)
	case http.MethodPost:
		m.handleOrganizationsCreate(w, r)
	default:
		m.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleOrganizationsList returns list of organizations with pagination
func (m *MockAPIServer) handleOrganizationsList(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	orgList := make([]interface{}, 0, len(m.orgs))
	for _, org := range m.orgs {
		orgList = append(orgList, org)
	}

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": orgList,
		"meta": map[string]interface{}{
			"total_items":  len(orgList),
			"page":         1,
			"limit":        25,
			"total_pages":  1,
			"has_more":     false,
			"first_page":   1,
			"last_page":    1,
			"from":         1,
			"to":           len(orgList),
			"per_page":     25,
			"current_page": 1,
		},
	})
}

// handleOrganizationsCreate creates a new organization
func (m *MockAPIServer) handleOrganizationsCreate(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	name, ok := req["name"].(string)
	if !ok || name == "" {
		m.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate UUID
	uuid := fmt.Sprintf("org-%03d", len(m.orgs)+1)

	// Create organization
	org := map[string]interface{}{
		"uuid":        uuid,
		"name":        name,
		"description": req["description"],
		"industry":    req["industry"],
		"website":     req["website"],
		"country":     req["country"],
		"timezone":    req["timezone"],
		"status":      "active",
		"created_at":  "2025-01-01T00:00:00Z",
		"updated_at":  "2025-01-01T00:00:00Z",
	}

	m.orgs[uuid] = org

	m.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data": org,
	})
}

// handleOrganizationByID handles /v1/organizations/{uuid} and /v2/organizations/{uuid} and sub-resources
func (m *MockAPIServer) handleOrganizationByID(w http.ResponseWriter, r *http.Request) {
	if err := m.checkAuth(r); err != nil {
		m.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Remove version prefix
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/v1/organizations/")
	path = strings.TrimPrefix(path, "/v2/organizations/")
	parts := strings.Split(path, "/")

	// Handle /uuid/{uuid} path format
	uuid := parts[0]
	if uuid == "uuid" && len(parts) > 1 {
		uuid = parts[1]
		parts = parts[1:] // Skip "uuid" part
	}

	// Handle sub-resources
	if len(parts) > 1 {
		switch parts[1] {
		case "servers":
			m.handleOrganizationServers(w, r, uuid)
			return
		case "users":
			m.handleOrganizationUsers(w, r, uuid)
			return
		case "alerts":
			m.handleOrganizationAlerts(w, r, uuid)
			return
		default:
			m.writeError(w, http.StatusNotFound, "Resource not found")
			return
		}
	}

	// Handle main organization resource
	switch r.Method {
	case http.MethodGet:
		m.handleOrganizationGet(w, r, uuid)
	case http.MethodPut:
		m.handleOrganizationUpdate(w, r, uuid)
	case http.MethodDelete:
		m.handleOrganizationDelete(w, r, uuid)
	default:
		m.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleOrganizationGet retrieves an organization by UUID
func (m *MockAPIServer) handleOrganizationGet(w http.ResponseWriter, r *http.Request, uuid string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	org, ok := m.orgs[uuid]
	if !ok {
		m.writeError(w, http.StatusNotFound, "Organization not found")
		return
	}

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": org,
	})
}

// handleOrganizationUpdate updates an organization
func (m *MockAPIServer) handleOrganizationUpdate(w http.ResponseWriter, r *http.Request, uuid string) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	org, ok := m.orgs[uuid].(map[string]interface{})
	if !ok {
		m.writeError(w, http.StatusNotFound, "Organization not found")
		return
	}

	// Update fields
	if name, ok := req["name"].(string); ok && name != "" {
		org["name"] = name
	}
	if desc, ok := req["description"].(string); ok {
		org["description"] = desc
	}
	if industry, ok := req["industry"].(string); ok {
		org["industry"] = industry
	}
	if website, ok := req["website"].(string); ok {
		org["website"] = website
	}
	org["updated_at"] = "2025-01-01T01:00:00Z"

	m.orgs[uuid] = org

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": org,
	})
}

// handleOrganizationDelete deletes an organization
func (m *MockAPIServer) handleOrganizationDelete(w http.ResponseWriter, r *http.Request, uuid string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.orgs[uuid]
	if !ok {
		m.writeError(w, http.StatusNotFound, "Organization not found")
		return
	}

	delete(m.orgs, uuid)
	w.WriteHeader(http.StatusNoContent)
}

// handleOrganizationServers returns servers for an organization
func (m *MockAPIServer) handleOrganizationServers(w http.ResponseWriter, r *http.Request, orgUUID string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if organization exists
	_, ok := m.orgs[orgUUID]
	if !ok {
		m.writeError(w, http.StatusNotFound, "Organization not found")
		return
	}

	// Filter servers by organization
	serverList := make([]interface{}, 0)
	for _, server := range m.servers {
		if serverMap, ok := server.(map[string]interface{}); ok {
			if orgID, ok := serverMap["organization_id"].(float64); ok {
				// Map org UUIDs to IDs (org-001 -> 1, org-002 -> 2, etc.)
				expectedOrgID := 0
				if orgUUID == "org-001" {
					expectedOrgID = 1
				} else if orgUUID == "org-002" {
					expectedOrgID = 2
				} else if orgUUID == "org-003" {
					expectedOrgID = 3
				}
				if int(orgID) == expectedOrgID {
					serverList = append(serverList, server)
				}
			}
		}
	}

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": serverList,
		"meta": map[string]interface{}{
			"total_items":  len(serverList),
			"page":         1,
			"limit":        25,
			"total_pages":  1,
			"has_more":     false,
			"first_page":   1,
			"last_page":    1,
			"from":         1,
			"to":           len(serverList),
			"per_page":     25,
			"current_page": 1,
		},
	})
}

// handleOrganizationUsers returns users for an organization
func (m *MockAPIServer) handleOrganizationUsers(w http.ResponseWriter, r *http.Request, orgUUID string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if organization exists
	_, ok := m.orgs[orgUUID]
	if !ok {
		m.writeError(w, http.StatusNotFound, "Organization not found")
		return
	}

	// Return empty user list (not implemented in fixtures)
	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": []interface{}{},
		"meta": map[string]interface{}{
			"total_items":  0,
			"page":         1,
			"limit":        25,
			"total_pages":  0,
			"has_more":     false,
			"first_page":   1,
			"last_page":    0,
			"from":         0,
			"to":           0,
			"per_page":     25,
			"current_page": 1,
		},
	})
}

// handleOrganizationAlerts returns alerts for an organization
func (m *MockAPIServer) handleOrganizationAlerts(w http.ResponseWriter, r *http.Request, orgUUID string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if organization exists
	_, ok := m.orgs[orgUUID]
	if !ok {
		m.writeError(w, http.StatusNotFound, "Organization not found")
		return
	}

	// Return alerts for this organization (simplified - return all)
	alertList := make([]interface{}, 0)
	for _, alert := range m.alerts {
		alertList = append(alertList, alert)
	}

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": alertList,
		"meta": map[string]interface{}{
			"total_items":  len(alertList),
			"page":         1,
			"limit":        25,
			"total_pages":  1,
			"has_more":     false,
			"first_page":   1,
			"last_page":    1,
			"from":         1,
			"to":           len(alertList),
			"per_page":     25,
			"current_page": 1,
		},
	})
}

// handleMetricsSubmit handles POST /v2/metrics/submit
func (m *MockAPIServer) handleMetricsSubmit(w http.ResponseWriter, r *http.Request) {
	if err := m.checkAuth(r); err != nil {
		m.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if r.Method != http.MethodPost {
		m.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Just accept the metrics without storing
	var metrics map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		m.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	m.writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"message": "Metrics accepted",
		"status":  "success",
	})
}

// handleMetricsQuery handles GET /v2/metrics
func (m *MockAPIServer) handleMetricsQuery(w http.ResponseWriter, r *http.Request) {
	if err := m.checkAuth(r); err != nil {
		m.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Return sample metrics data
	metrics := loadFixture(m.t, "metrics.json")
	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": metrics,
	})
}

// handleAlerts handles /v1/alerts/rules (GET, POST)
func (m *MockAPIServer) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if err := m.checkAuth(r); err != nil {
		m.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		m.handleAlertsList(w, r)
	case http.MethodPost:
		m.handleAlertsCreate(w, r)
	default:
		m.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleAlertsList returns list of alerts with pagination
func (m *MockAPIServer) handleAlertsList(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alertList := make([]interface{}, 0, len(m.alerts))
	for _, alert := range m.alerts {
		alertList = append(alertList, alert)
	}

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": alertList,
		"meta": map[string]interface{}{
			"total_items":  len(alertList),
			"page":         1,
			"limit":        25,
			"total_pages":  1,
			"has_more":     false,
			"first_page":   1,
			"last_page":    1,
			"from":         1,
			"to":           len(alertList),
			"per_page":     25,
			"current_page": 1,
		},
	})
}

// handleAlertsCreate creates a new alert
func (m *MockAPIServer) handleAlertsCreate(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	name, ok := req["name"].(string)
	if !ok || name == "" {
		m.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID (find max ID and add 1)
	maxID := 0
	for _, alert := range m.alerts {
		if alertMap, ok := alert.(map[string]interface{}); ok {
			// Check float64 (from JSON fixtures)
			if id, ok := alertMap["id"].(float64); ok && int(id) > maxID {
				maxID = int(id)
			}
			// Check int (from Go code)
			if id, ok := alertMap["id"].(int); ok && id > maxID {
				maxID = id
			}
		}
	}
	newID := maxID + 1
	uuid := fmt.Sprintf("alert-%03d", newID)

	// Create alert
	alert := map[string]interface{}{
		"uuid":        uuid,
		"id":          newID,
		"name":        name,
		"description": req["description"],
		"type":        req["type"],
		"metric_name": req["metric_name"],
		"condition":   req["condition"],
		"threshold":   req["threshold"],
		"duration":    req["duration"],
		"frequency":   req["frequency"],
		"enabled":     req["enabled"],
		"status":      req["status"],
		"severity":    req["severity"],
		"channels":    req["channels"],
		"created_at":  "2025-01-01T00:00:00Z",
		"updated_at":  "2025-01-01T00:00:00Z",
	}

	// Add organization_id if provided
	if orgID, ok := req["organization_id"].(float64); ok {
		alert["organization_id"] = int(orgID)
	}

	// Add server_id if provided
	if serverID, ok := req["server_id"].(float64); ok {
		alert["server_id"] = int(serverID)
	}

	m.alerts[uuid] = alert

	m.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data": alert,
	})
}

// handleAlertByID handles /v1/alerts/rules/{id} (GET, PUT, DELETE)
func (m *MockAPIServer) handleAlertByID(w http.ResponseWriter, r *http.Request) {
	if err := m.checkAuth(r); err != nil {
		m.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Remove version and base path
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/v1/alerts/rules/")
	path = strings.TrimPrefix(path, "/v2/alerts/")

	// Get the ID (could be numeric ID or UUID)
	id := path

	switch r.Method {
	case http.MethodGet:
		m.handleAlertGet(w, r, id)
	case http.MethodPut:
		m.handleAlertUpdate(w, r, id)
	case http.MethodDelete:
		m.handleAlertDelete(w, r, id)
	default:
		m.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleAlertGet retrieves an alert by ID or UUID
func (m *MockAPIServer) handleAlertGet(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try to find by UUID first
	if alert, ok := m.alerts[id]; ok {
		m.writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": alert,
		})
		return
	}

	// Try to find by numeric ID
	for _, alert := range m.alerts {
		if alertMap, ok := alert.(map[string]interface{}); ok {
			// Check for float64 (from JSON)
			if alertID, ok := alertMap["id"].(float64); ok && fmt.Sprintf("%.0f", alertID) == id {
				m.writeJSON(w, http.StatusOK, map[string]interface{}{
					"data": alert,
				})
				return
			}
			// Check for int (from Go code)
			if alertID, ok := alertMap["id"].(int); ok && fmt.Sprintf("%d", alertID) == id {
				m.writeJSON(w, http.StatusOK, map[string]interface{}{
					"data": alert,
				})
				return
			}
		}
	}

	m.writeError(w, http.StatusNotFound, "Alert not found")
}

// handleAlertUpdate updates an alert
func (m *MockAPIServer) handleAlertUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Find alert by ID or UUID
	var alertUUID string
	var alertMap map[string]interface{}

	// Try UUID first
	if alert, ok := m.alerts[id]; ok {
		if aMap, ok := alert.(map[string]interface{}); ok {
			alertUUID = id
			alertMap = aMap
		}
	}

	// Try numeric ID
	if alertMap == nil {
		for uuid, alert := range m.alerts {
			if aMap, ok := alert.(map[string]interface{}); ok {
				// Check float64 (from JSON)
				if alertID, ok := aMap["id"].(float64); ok && fmt.Sprintf("%.0f", alertID) == id {
					alertUUID = uuid
					alertMap = aMap
					break
				}
				// Check int (from Go code)
				if alertID, ok := aMap["id"].(int); ok && fmt.Sprintf("%d", alertID) == id {
					alertUUID = uuid
					alertMap = aMap
					break
				}
			}
		}
	}

	if alertMap == nil {
		m.writeError(w, http.StatusNotFound, "Alert not found")
		return
	}

	// Update fields
	if name, ok := req["name"].(string); ok && name != "" {
		alertMap["name"] = name
	}
	if desc, ok := req["description"].(string); ok {
		alertMap["description"] = desc
	}
	if threshold, ok := req["threshold"].(float64); ok {
		alertMap["threshold"] = threshold
	}
	if severity, ok := req["severity"].(string); ok {
		alertMap["severity"] = severity
	}
	if enabled, ok := req["enabled"].(bool); ok {
		alertMap["enabled"] = enabled
	}
	alertMap["updated_at"] = "2025-01-01T01:00:00Z"

	m.alerts[alertUUID] = alertMap

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": alertMap,
	})
}

// handleAlertDelete deletes an alert
func (m *MockAPIServer) handleAlertDelete(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try UUID first
	if _, ok := m.alerts[id]; ok {
		delete(m.alerts, id)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Find and delete alert by numeric ID
	for uuid, alert := range m.alerts {
		if alertMap, ok := alert.(map[string]interface{}); ok {
			// Check float64 (from JSON)
			if alertID, ok := alertMap["id"].(float64); ok && fmt.Sprintf("%.0f", alertID) == id {
				delete(m.alerts, uuid)
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// Check int (from Go code)
			if alertID, ok := alertMap["id"].(int); ok && fmt.Sprintf("%d", alertID) == id {
				delete(m.alerts, uuid)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
	}

	m.writeError(w, http.StatusNotFound, "Alert not found")
}

// handleAlertActions handles alert action endpoints like enable/disable
func (m *MockAPIServer) handleAlertActions(w http.ResponseWriter, r *http.Request) {
	if err := m.checkAuth(r); err != nil {
		m.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Parse path: /v1/alerts/{id}/{action}
	path := strings.TrimPrefix(r.URL.Path, "/v1/alerts/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		m.writeError(w, http.StatusNotFound, "Invalid path")
		return
	}

	id := parts[0]
	action := parts[1]

	switch action {
	case "enable":
		m.handleAlertEnable(w, r, id)
	case "disable":
		m.handleAlertDisable(w, r, id)
	default:
		m.writeError(w, http.StatusNotFound, "Unknown action")
	}
}

// handleAlertEnable enables an alert
func (m *MockAPIServer) handleAlertEnable(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try UUID first
	if alert, ok := m.alerts[id]; ok {
		if alertMap, ok := alert.(map[string]interface{}); ok {
			alertMap["enabled"] = true
			alertMap["status"] = "active"
			alertMap["updated_at"] = "2025-01-01T01:00:00Z"
			m.alerts[id] = alertMap

			m.writeJSON(w, http.StatusOK, map[string]interface{}{
				"data": alertMap,
			})
			return
		}
	}

	// Find alert by numeric ID
	for uuid, alert := range m.alerts {
		if alertMap, ok := alert.(map[string]interface{}); ok {
			// Check float64 (from JSON)
			if alertID, ok := alertMap["id"].(float64); ok && fmt.Sprintf("%.0f", alertID) == id {
				alertMap["enabled"] = true
				alertMap["status"] = "active"
				alertMap["updated_at"] = "2025-01-01T01:00:00Z"
				m.alerts[uuid] = alertMap

				m.writeJSON(w, http.StatusOK, map[string]interface{}{
					"data": alertMap,
				})
				return
			}
			// Check int (from Go code)
			if alertID, ok := alertMap["id"].(int); ok && fmt.Sprintf("%d", alertID) == id {
				alertMap["enabled"] = true
				alertMap["status"] = "active"
				alertMap["updated_at"] = "2025-01-01T01:00:00Z"
				m.alerts[uuid] = alertMap

				m.writeJSON(w, http.StatusOK, map[string]interface{}{
					"data": alertMap,
				})
				return
			}
		}
	}

	m.writeError(w, http.StatusNotFound, "Alert not found")
}

// handleAlertDisable disables an alert
func (m *MockAPIServer) handleAlertDisable(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try UUID first
	if alert, ok := m.alerts[id]; ok {
		if alertMap, ok := alert.(map[string]interface{}); ok {
			alertMap["enabled"] = false
			alertMap["status"] = "inactive"
			alertMap["updated_at"] = "2025-01-01T01:00:00Z"
			m.alerts[id] = alertMap

			m.writeJSON(w, http.StatusOK, map[string]interface{}{
				"data": alertMap,
			})
			return
		}
	}

	// Find alert by numeric ID
	for uuid, alert := range m.alerts {
		if alertMap, ok := alert.(map[string]interface{}); ok {
			// Check float64 (from JSON)
			if alertID, ok := alertMap["id"].(float64); ok && fmt.Sprintf("%.0f", alertID) == id {
				alertMap["enabled"] = false
				alertMap["status"] = "inactive"
				alertMap["updated_at"] = "2025-01-01T01:00:00Z"
				m.alerts[uuid] = alertMap

				m.writeJSON(w, http.StatusOK, map[string]interface{}{
					"data": alertMap,
				})
				return
			}
			// Check int (from Go code)
			if alertID, ok := alertMap["id"].(int); ok && fmt.Sprintf("%d", alertID) == id {
				alertMap["enabled"] = false
				alertMap["status"] = "inactive"
				alertMap["updated_at"] = "2025-01-01T01:00:00Z"
				m.alerts[uuid] = alertMap

				m.writeJSON(w, http.StatusOK, map[string]interface{}{
					"data": alertMap,
				})
				return
			}
		}
	}

	m.writeError(w, http.StatusNotFound, "Alert not found")
}

// handleProbes handles /v1/monitoring/probes and /v2/monitoring/probes (GET, POST)
func (m *MockAPIServer) handleProbes(w http.ResponseWriter, r *http.Request) {
	if err := m.checkAuth(r); err != nil {
		m.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	switch r.Method {
	case "GET":
		m.handleProbesList(w, r)
	case "POST":
		m.handleProbesCreate(w, r)
	default:
		m.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleProbesList returns list of probes with pagination
func (m *MockAPIServer) handleProbesList(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	probeList := make([]interface{}, 0, len(m.probes))
	for _, probe := range m.probes {
		probeList = append(probeList, probe)
	}

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": probeList,
		"meta": map[string]interface{}{
			"total_items":  len(probeList),
			"page":         1,
			"limit":        25,
			"total_pages":  1,
			"has_more":     false,
			"first_page":   1,
			"last_page":    1,
			"from":         1,
			"to":           len(probeList),
			"per_page":     25,
			"current_page": 1,
		},
	})
}

// handleProbesCreate creates a new probe
func (m *MockAPIServer) handleProbesCreate(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	name, ok := req["name"].(string)
	if !ok || name == "" {
		m.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	probeType, ok := req["type"].(string)
	if !ok || probeType == "" {
		m.writeError(w, http.StatusBadRequest, "type is required")
		return
	}

	target, ok := req["target"].(string)
	if !ok || target == "" {
		m.writeError(w, http.StatusBadRequest, "target is required")
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID (find max ID and add 1)
	maxID := 0
	for _, probe := range m.probes {
		if probeMap, ok := probe.(map[string]interface{}); ok {
			// Check float64 (from JSON fixtures)
			if id, ok := probeMap["id"].(float64); ok && int(id) > maxID {
				maxID = int(id)
			}
			// Check int (from Go code)
			if id, ok := probeMap["id"].(int); ok && id > maxID {
				maxID = id
			}
		}
	}
	newID := maxID + 1
	uuid := fmt.Sprintf("probe-%03d", newID)

	// Create probe
	probe := map[string]interface{}{
		"id":              newID,
		"probe_uuid":      uuid,
		"name":            name,
		"description":     req["description"],
		"type":            probeType,
		"target":          target,
		"interval":        req["interval"],
		"timeout":         req["timeout"],
		"enabled":         req["enabled"],
		"organization_id": req["organization_id"],
		"server_id":       req["server_id"],
		"regions":         req["regions"],
		"config":          req["config"],
		"alert_config":    req["alert_config"],
		"tags":            req["tags"],
		"created_at":      "2025-01-01T00:00:00Z",
		"updated_at":      "2025-01-01T00:00:00Z",
	}

	m.probes[uuid] = probe

	m.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data": probe,
	})
}

// handleProbeByID handles /v1/monitoring/probes/{id} (GET, PUT, DELETE)
func (m *MockAPIServer) handleProbeByID(w http.ResponseWriter, r *http.Request) {
	if err := m.checkAuth(r); err != nil {
		m.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/v1/monitoring/probes/")
	path = strings.TrimPrefix(path, "/v2/monitoring/probes/")
	id := strings.TrimSuffix(path, "/")

	switch r.Method {
	case "GET":
		m.handleProbeGet(w, r, id)
	case "PUT":
		m.handleProbeUpdate(w, r, id)
	case "DELETE":
		m.handleProbeDelete(w, r, id)
	default:
		m.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleProbeGet retrieves a probe by ID or UUID
func (m *MockAPIServer) handleProbeGet(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try to find by UUID first
	if probe, ok := m.probes[id]; ok {
		m.writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": probe,
		})
		return
	}

	// Try to find by numeric ID
	for _, probe := range m.probes {
		if probeMap, ok := probe.(map[string]interface{}); ok {
			// Check for float64 (from JSON)
			if probeID, ok := probeMap["id"].(float64); ok && fmt.Sprintf("%.0f", probeID) == id {
				m.writeJSON(w, http.StatusOK, map[string]interface{}{
					"data": probe,
				})
				return
			}
			// Check for int (from Go code)
			if probeID, ok := probeMap["id"].(int); ok && fmt.Sprintf("%d", probeID) == id {
				m.writeJSON(w, http.StatusOK, map[string]interface{}{
					"data": probe,
				})
				return
			}
		}
	}

	m.writeError(w, http.StatusNotFound, "Probe not found")
}

// handleProbeUpdate updates a probe
func (m *MockAPIServer) handleProbeUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Find probe by ID or UUID
	var probeUUID string
	var probeMap map[string]interface{}

	// Try UUID first
	if probe, ok := m.probes[id]; ok {
		if pMap, ok := probe.(map[string]interface{}); ok {
			probeUUID = id
			probeMap = pMap
		}
	}

	// Try numeric ID
	if probeMap == nil {
		for uuid, probe := range m.probes {
			if pMap, ok := probe.(map[string]interface{}); ok {
				// Check float64 (from JSON)
				if probeID, ok := pMap["id"].(float64); ok && fmt.Sprintf("%.0f", probeID) == id {
					probeUUID = uuid
					probeMap = pMap
					break
				}
				// Check int (from Go code)
				if probeID, ok := pMap["id"].(int); ok && fmt.Sprintf("%d", probeID) == id {
					probeUUID = uuid
					probeMap = pMap
					break
				}
			}
		}
	}

	if probeMap == nil {
		m.writeError(w, http.StatusNotFound, "Probe not found")
		return
	}

	// Update fields
	if name, ok := req["name"].(string); ok && name != "" {
		probeMap["name"] = name
	}
	if desc, ok := req["description"].(string); ok {
		probeMap["description"] = desc
	}
	if interval, ok := req["interval"].(float64); ok {
		probeMap["interval"] = interval
	}
	if timeout, ok := req["timeout"].(float64); ok {
		probeMap["timeout"] = timeout
	}
	if enabled, ok := req["enabled"].(bool); ok {
		probeMap["enabled"] = enabled
	}
	if target, ok := req["target"].(string); ok {
		probeMap["target"] = target
	}
	if regions, ok := req["regions"].([]interface{}); ok {
		probeMap["regions"] = regions
	}
	if config, ok := req["config"].(map[string]interface{}); ok {
		probeMap["config"] = config
	}
	if alertConfig, ok := req["alert_config"].(map[string]interface{}); ok {
		probeMap["alert_config"] = alertConfig
	}
	probeMap["updated_at"] = "2025-01-01T01:00:00Z"

	m.probes[probeUUID] = probeMap

	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": probeMap,
	})
}

// handleProbeDelete deletes a probe
func (m *MockAPIServer) handleProbeDelete(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try UUID first
	if _, ok := m.probes[id]; ok {
		delete(m.probes, id)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Find and delete probe by numeric ID
	for uuid, probe := range m.probes {
		if probeMap, ok := probe.(map[string]interface{}); ok {
			// Check float64 (from JSON)
			if probeID, ok := probeMap["id"].(float64); ok && fmt.Sprintf("%.0f", probeID) == id {
				delete(m.probes, uuid)
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// Check int (from Go code)
			if probeID, ok := probeMap["id"].(int); ok && fmt.Sprintf("%d", probeID) == id {
				delete(m.probes, uuid)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
	}

	m.writeError(w, http.StatusNotFound, "Probe not found")
}

// handleHealth handles /v2/system/health
func (m *MockAPIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"version": "2.3.1-mock",
		"uptime":  "1h",
	})
}

// handleVersion handles /v2/system/version
func (m *MockAPIServer) handleVersion(w http.ResponseWriter, r *http.Request) {
	m.writeJSON(w, http.StatusOK, map[string]interface{}{
		"version": "2.3.1-mock",
		"commit":  "abc1234",
	})
}

// handleNotFound handles unknown routes
func (m *MockAPIServer) handleNotFound(w http.ResponseWriter, r *http.Request) {
	m.writeError(w, http.StatusNotFound, fmt.Sprintf("Endpoint not found: %s", r.URL.Path))
}

// Helper methods for writing responses

func (m *MockAPIServer) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (m *MockAPIServer) writeError(w http.ResponseWriter, status int, message string) {
	m.writeJSON(w, status, map[string]interface{}{
		"error": message,
	})
}
