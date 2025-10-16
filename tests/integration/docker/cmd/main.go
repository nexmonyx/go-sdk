// Package main implements a standalone mock API server for integration testing.
// This server mimics the Nexmonyx API and can be run as a Docker container
// or directly as a Go application.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// MockAPIServer provides a standalone HTTP server that mimics the Nexmonyx API
type MockAPIServer struct {
	Server  http.Server
	mu      sync.RWMutex
	servers map[string]interface{}
	orgs    map[string]interface{}
	alerts  map[string]interface{}
	users   map[string]interface{}
	probes      map[string]interface{}
	nextID      int
	authEnabled bool
	authToken   string
}

// testing.T interface for compatibility with test code
type testing struct{}

// Fake testing interface for non-test usage
func (t *testing) Logf(format string, args ...interface{})  { log.Printf(format, args...) }
func (t *testing) Errorf(format string, args ...interface{}) { log.Printf("ERROR: "+format, args...) }
func (t *testing) Fatalf(format string, args ...interface{}) { log.Fatalf(format, args...) }
func (t *testing) FailNow()                                  { os.Exit(1) }

// NewMockAPIServer creates and starts a new mock API server
func NewMockAPIServer(addr string, authToken string, logf func(string, ...interface{})) *MockAPIServer {
	mock := &MockAPIServer{
		servers:     make(map[string]interface{}),
		orgs:        make(map[string]interface{}),
		alerts:      make(map[string]interface{}),
		users:       make(map[string]interface{}),
		probes:      make(map[string]interface{}),
		nextID:      1000,
		authEnabled: authToken != "",
		authToken:   authToken,
		t:           &testing{},
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mock.registerRoutes(mux)

	mock.Server = http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return mock
}

// registerRoutes registers all API routes
func (m *MockAPIServer) registerRoutes(mux *http.ServeMux) {
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// Ready check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	// Servers endpoints
	mux.HandleFunc("/api/v1/servers", m.handleServers)
	mux.HandleFunc("/api/v1/servers/", m.handleServerDetail)

	// Organizations endpoints
	mux.HandleFunc("/api/v1/organizations", m.handleOrganizations)
	mux.HandleFunc("/api/v1/organizations/", m.handleOrgDetail)

	// Alerts endpoints
	mux.HandleFunc("/api/v1/alerts", m.handleAlerts)
	mux.HandleFunc("/api/v1/alerts/", m.handleAlertDetail)

	// Users endpoints
	mux.HandleFunc("/api/v1/users", m.handleUsers)
	mux.HandleFunc("/api/v1/users/", m.handleUserDetail)

	// Probes endpoints
	mux.HandleFunc("/api/v1/probes", m.handleProbes)
	mux.HandleFunc("/api/v1/probes/", m.handleProbeDetail)

	// Metrics endpoints
	mux.HandleFunc("/api/v1/metrics", m.handleMetrics)
	mux.HandleFunc("/api/v1/metrics/submit", m.handleMetricsSubmit)
}

// Middleware for authentication check
func (m *MockAPIServer) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.authEnabled {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token != m.authToken {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}

// handleServers handles GET /api/v1/servers
func (m *MockAPIServer) handleServers(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		servers := make([]interface{}, 0, len(m.servers))
		for _, s := range m.servers {
			servers = append(servers, s)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": servers})
	}))(w, r)
}

// handleServerDetail handles server detail endpoints
func (m *MockAPIServer) handleServerDetail(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/servers/")

		m.mu.RLock()
		server, exists := m.servers[id]
		m.mu.RUnlock()

		if !exists {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(server)
	}))(w, r)
}

// handleOrganizations handles GET /api/v1/organizations
func (m *MockAPIServer) handleOrganizations(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		orgs := make([]interface{}, 0, len(m.orgs))
		for _, o := range m.orgs {
			orgs = append(orgs, o)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": orgs})
	}))(w, r)
}

// handleOrgDetail handles organization detail endpoints
func (m *MockAPIServer) handleOrgDetail(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/")

		m.mu.RLock()
		org, exists := m.orgs[id]
		m.mu.RUnlock()

		if !exists {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(org)
	}))(w, r)
}

// handleAlerts handles GET /api/v1/alerts
func (m *MockAPIServer) handleAlerts(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		alerts := make([]interface{}, 0, len(m.alerts))
		for _, a := range m.alerts {
			alerts = append(alerts, a)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": alerts})
	}))(w, r)
}

// handleAlertDetail handles alert detail endpoints
func (m *MockAPIServer) handleAlertDetail(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/alerts/")

		m.mu.RLock()
		alert, exists := m.alerts[id]
		m.mu.RUnlock()

		if !exists {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alert)
	}))(w, r)
}

// handleUsers handles GET /api/v1/users
func (m *MockAPIServer) handleUsers(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		users := make([]interface{}, 0, len(m.users))
		for _, u := range m.users {
			users = append(users, u)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": users})
	}))(w, r)
}

// handleUserDetail handles user detail endpoints
func (m *MockAPIServer) handleUserDetail(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")

		m.mu.RLock()
		user, exists := m.users[id]
		m.mu.RUnlock()

		if !exists {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}))(w, r)
}

// handleProbes handles GET /api/v1/probes
func (m *MockAPIServer) handleProbes(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		probes := make([]interface{}, 0, len(m.probes))
		for _, p := range m.probes {
			probes = append(probes, p)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": probes})
	}))(w, r)
}

// handleProbeDetail handles probe detail endpoints
func (m *MockAPIServer) handleProbeDetail(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/probes/")

		m.mu.RLock()
		probe, exists := m.probes[id]
		m.mu.RUnlock()

		if !exists {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(probe)
	}))(w, r)
}

// handleMetrics handles GET /api/v1/metrics
func (m *MockAPIServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{}})
	}))(w, r)
}

// handleMetricsSubmit handles POST /api/v1/metrics/submit
func (m *MockAPIServer) handleMetricsSubmit(w http.ResponseWriter, r *http.Request) {
	m.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "accepted"})
	}))(w, r)
}

func main() {
	// Get configuration from environment variables
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("API_HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	authToken := os.Getenv("AUTH_TOKEN")

	addr := fmt.Sprintf("%s:%s", host, port)

	// Create mock API server
	mock := NewMockAPIServer(addr, authToken, log.Printf)

	// Start server in goroutine
	go func() {
		log.Printf("Starting mock API server on %s\n", addr)
		if err := mock.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	log.Println("Shutting down mock API server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mock.Server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Mock API server stopped")
}
