package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// SubmitSystemdServices submits systemd service data
func (s *SystemdService) Submit(ctx context.Context, request *SystemdServiceRequest) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/api/v1/systemd",
		Body:   request,
		Result: &resp,
	})
	return err
}

// GetSystemdServices retrieves systemd services for a server
func (s *SystemdService) Get(ctx context.Context, serverUUID string) ([]*SystemdServiceInfo, error) {
	var resp StandardResponse
	var services []*SystemdServiceInfo
	resp.Data = &services

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/systemd/%s", serverUUID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return services, nil
}

// ListSystemdServices lists systemd services with filters
func (s *SystemdService) List(ctx context.Context, opts *ListOptions) ([]*SystemdServiceInfo, *PaginationMeta, error) {
	var resp PaginatedResponse
	var services []*SystemdServiceInfo
	resp.Data = &services

	req := &Request{
		Method: "GET",
		Path:   "/api/v1/systemd",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return services, resp.Meta, nil
}

// GetServiceByName retrieves a specific systemd service by name
func (s *SystemdService) GetServiceByName(ctx context.Context, serverUUID, serviceName string) (*SystemdServiceInfo, error) {
	var resp StandardResponse
	resp.Data = &SystemdServiceInfo{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/systemd/%s/service/%s", serverUUID, serviceName),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if service, ok := resp.Data.(*SystemdServiceInfo); ok {
		return service, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetSystemStats retrieves systemd system statistics
func (s *SystemdService) GetSystemStats(ctx context.Context, serverUUID string) (*SystemdSystemStats, error) {
	var resp StandardResponse
	resp.Data = &SystemdSystemStats{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/systemd/%s/stats", serverUUID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if stats, ok := resp.Data.(*SystemdSystemStats); ok {
		return stats, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// SystemdServiceRequest represents a request to submit systemd service data
type SystemdServiceRequest struct {
	ServerUUID  string               `json:"server_uuid"`
	CollectedAt string               `json:"collected_at"`
	Services    []SystemdServiceInfo `json:"services"`
	SystemStats *SystemdSystemStats  `json:"system_stats,omitempty"`
}

// SystemdServiceInfo represents information about a systemd service
type SystemdServiceInfo struct {
	Name                 string                 `json:"name"`
	UnitType             string                 `json:"unit_type"`
	Description          string                 `json:"description,omitempty"`
	LoadState            string                 `json:"load_state"`
	ActiveState          string                 `json:"active_state"`
	SubState             string                 `json:"sub_state"`
	UnitState            string                 `json:"unit_state"`
	MainPID              int                    `json:"main_pid,omitempty"`
	Type                 string                 `json:"type,omitempty"`
	User                 string                 `json:"user,omitempty"`
	Group                string                 `json:"group,omitempty"`
	WorkingDir           string                 `json:"working_dir,omitempty"`
	ExecStart            []string               `json:"exec_start,omitempty"`
	ExecReload           []string               `json:"exec_reload,omitempty"`
	ExecStop             []string               `json:"exec_stop,omitempty"`
	Environment          []string               `json:"environment,omitempty"`
	Wants                []string               `json:"wants,omitempty"`
	After                []string               `json:"after,omitempty"`
	Before               []string               `json:"before,omitempty"`
	MemoryCurrent        int64                  `json:"memory_current,omitempty"`
	MemoryPeak           int64                  `json:"memory_peak,omitempty"`
	MemoryLimit          int64                  `json:"memory_limit,omitempty"`
	CPUUsageNSec         int64                  `json:"cpu_usage_nsec,omitempty"`
	CPUUsagePercent      float64                `json:"cpu_usage_percent,omitempty"`
	TasksCurrent         int                    `json:"tasks_current,omitempty"`
	TasksLimit           int                    `json:"tasks_limit,omitempty"`
	ExitCode             int                    `json:"exit_code,omitempty"`
	ExitStatus           string                 `json:"exit_status,omitempty"`
	Result               string                 `json:"result,omitempty"`
	StatusText           string                 `json:"status_text,omitempty"`
	RestartCount         int                    `json:"restart_count,omitempty"`
	StartupDuration      float64                `json:"startup_duration,omitempty"`
	HealthScore          int                    `json:"health_score,omitempty"`
	PrivateTmp           bool                   `json:"private_tmp,omitempty"`
	PrivateNetwork       bool                   `json:"private_network,omitempty"`
	ProtectSystem        string                 `json:"protect_system,omitempty"`
	ProtectHome          string                 `json:"protect_home,omitempty"`
	NoNewPrivileges      bool                   `json:"no_new_privileges,omitempty"`
	DetectionMethod      string                 `json:"detection_method,omitempty"`
	ActiveEnterTimestamp *time.Time             `json:"active_enter_timestamp,omitempty"`
	ActiveExitTimestamp  *time.Time             `json:"active_exit_timestamp,omitempty"`
	NextElapseTime       *time.Time             `json:"next_elapse_time,omitempty"` // For timer units
	AdditionalInfo       map[string]interface{} `json:"additional_info,omitempty"`
}

// SystemdSystemStats represents system-wide systemd statistics
type SystemdSystemStats struct {
	TotalUnits         int       `json:"total_units"`
	ServiceUnits       int       `json:"service_units"`
	SocketUnits        int       `json:"socket_units"`
	TargetUnits        int       `json:"target_units"`
	TimerUnits         int       `json:"timer_units"`
	MountUnits         int       `json:"mount_units"`
	DeviceUnits        int       `json:"device_units"`
	ScopeUnits         int       `json:"scope_units"`
	SliceUnits         int       `json:"slice_units"`
	ActiveUnits        int       `json:"active_units"`
	InactiveUnits      int       `json:"inactive_units"`
	FailedUnits        int       `json:"failed_units"`
	EnabledUnits       int       `json:"enabled_units"`
	DisabledUnits      int       `json:"disabled_units"`
	MaskedUnits        int       `json:"masked_units"`
	SystemStartupTime  float64   `json:"system_startup_time"`
	LastBootTime       time.Time `json:"last_boot_time"`
	SystemManagerPID   int       `json:"system_manager_pid"`
	SystemState        string    `json:"system_state"`
	TotalMemoryUsage   int64     `json:"total_memory_usage"`
	TotalCPUUsage      float64   `json:"total_cpu_usage"`
	TotalTaskCount     int       `json:"total_task_count"`
	OverallHealthScore int       `json:"overall_health_score"`
	CriticalServices   []string  `json:"critical_services,omitempty"`
	SystemdIssues      []string  `json:"systemd_issues,omitempty"`
	RecentFailures     int       `json:"recent_failures"`
}

// Helper methods for SystemdServiceInfo

// IsHealthy returns true if the service is considered healthy
func (s *SystemdServiceInfo) IsHealthy() bool {
	return s.ActiveState == "active" && s.SubState == "running" && s.HealthScore >= 80
}

// IsFailed returns true if the service is in a failed state
func (s *SystemdServiceInfo) IsFailed() bool {
	return s.ActiveState == "failed" || s.SubState == "failed"
}

// IsEnabled returns true if the service is enabled
func (s *SystemdServiceInfo) IsEnabled() bool {
	return s.UnitState == "enabled"
}

// GetAdditionalInfo retrieves a value from additional info
func (s *SystemdServiceInfo) GetAdditionalInfo(key string) (interface{}, bool) {
	if s.AdditionalInfo == nil {
		return nil, false
	}
	val, ok := s.AdditionalInfo[key]
	return val, ok
}

// Helper methods for SystemdSystemStats

// IsHealthy returns true if the system is healthy
func (s *SystemdSystemStats) IsHealthy() bool {
	return s.SystemState == "running" && s.FailedUnits == 0 && s.OverallHealthScore >= 90
}

// IsDegraded returns true if the system is in degraded state
func (s *SystemdSystemStats) IsDegraded() bool {
	return s.SystemState == "degraded" || s.FailedUnits > 0
}
