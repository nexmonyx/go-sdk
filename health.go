package nexmonyx

import (
	"context"
	"fmt"
)

// GetHealth retrieves the health status of the API
func (s *HealthService) GetHealth(ctx context.Context) (*HealthStatus, error) {
	var resp StandardResponse
	resp.Data = &HealthStatus{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/api/v1/health",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if health, ok := resp.Data.(*HealthStatus); ok {
		return health, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetHealthDetailed retrieves detailed health status
func (s *HealthService) GetHealthDetailed(ctx context.Context) (*DetailedHealthStatus, error) {
	var resp StandardResponse
	resp.Data = &DetailedHealthStatus{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/api/v1/health/detailed",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if health, ok := resp.Data.(*DetailedHealthStatus); ok {
		return health, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// HealthStatus represents the basic health status
type HealthStatus struct {
	Status    string      `json:"status"`
	Healthy   bool        `json:"healthy"`
	Version   string      `json:"version"`
	Timestamp *CustomTime `json:"timestamp"`
}

// DetailedHealthStatus represents detailed health information
type DetailedHealthStatus struct {
	Status      string                 `json:"status"`
	Healthy     bool                   `json:"healthy"`
	Version     string                 `json:"version"`
	Timestamp   *CustomTime            `json:"timestamp"`
	Uptime      int64                  `json:"uptime"`
	Services    map[string]ServiceHealth `json:"services"`
	Database    *DatabaseHealth        `json:"database"`
	Redis       *RedisHealth           `json:"redis"`
	Metrics     map[string]interface{} `json:"metrics"`
}

// ServiceHealth represents the health of a service
type ServiceHealth struct {
	Healthy      bool   `json:"healthy"`
	Status       string `json:"status"`
	Message      string `json:"message,omitempty"`
	ResponseTime int    `json:"response_time,omitempty"` // milliseconds
}

// DatabaseHealth represents database health
type DatabaseHealth struct {
	Healthy         bool   `json:"healthy"`
	ConnectionCount int    `json:"connection_count"`
	MaxConnections  int    `json:"max_connections"`
	ResponseTime    int    `json:"response_time"` // milliseconds
	Version         string `json:"version"`
}

// RedisHealth represents Redis health
type RedisHealth struct {
	Healthy         bool   `json:"healthy"`
	Connected       bool   `json:"connected"`
	ResponseTime    int    `json:"response_time"` // milliseconds
	MemoryUsage     int64  `json:"memory_usage"`
	ConnectedClients int   `json:"connected_clients"`
	Version         string `json:"version"`
}