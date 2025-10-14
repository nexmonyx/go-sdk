package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nexmonyx/go-sdk/v2"
)

// AdvancedMonitoringAgent demonstrates a more realistic monitoring agent
// with concurrent probe execution, proper error handling, and graceful shutdown
type AdvancedMonitoringAgent struct {
	client           *nexmonyx.Client
	region           string
	agentID          string
	probes           []*nexmonyx.ProbeAssignment
	probeMutex       sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	heartbeatTicker  *time.Ticker
	probeRefreshTicker *time.Ticker
	
	// Statistics
	stats struct {
		sync.RWMutex
		probesExecuted   int64
		probesSuccessful int64
		probesFailed     int64
		totalResponseTime int64
	}
}

func main() {
	// Get configuration from environment
	monitoringKey := os.Getenv("NEXMONYX_MONITORING_KEY")
	if monitoringKey == "" {
		log.Fatal("NEXMONYX_MONITORING_KEY environment variable is required")
	}

	apiEndpoint := os.Getenv("NEXMONYX_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.nexmonyx.com"
	}

	region := os.Getenv("NEXMONYX_REGION")
	if region == "" {
		region = "us-east-1"
	}

	agentID := os.Getenv("NEXMONYX_AGENT_ID")
	if agentID == "" {
		agentID = fmt.Sprintf("advanced-agent-%d", time.Now().Unix())
	}

	// Create monitoring agent client
	client, err := nexmonyx.NewMonitoringAgentClient(&nexmonyx.Config{
		BaseURL: apiEndpoint,
		Auth: nexmonyx.AuthConfig{
			MonitoringKey: monitoringKey,
		},
		Debug:         os.Getenv("DEBUG") == "true",
		RetryCount:    3,
		RetryWaitTime: 2 * time.Second,
		RetryMaxWait:  30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create monitoring agent client: %v", err)
	}

	// Create agent instance
	ctx, cancel := context.WithCancel(context.Background())
	agent := &AdvancedMonitoringAgent{
		client:             client,
		region:             region,
		agentID:            agentID,
		ctx:                ctx,
		cancel:             cancel,
		heartbeatTicker:    time.NewTicker(30 * time.Second),
		probeRefreshTicker: time.NewTicker(5 * time.Minute),
	}

	fmt.Printf("Starting advanced monitoring agent [ID: %s, Region: %s]\n", agentID, region)

	// Test authentication
	if err := client.HealthCheck(ctx); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Println("✓ Authentication successful")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start agent
	agent.start()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nReceived shutdown signal, stopping agent...")
	agent.stop()
	fmt.Println("Agent stopped gracefully")
}

// start begins the agent's operation
func (a *AdvancedMonitoringAgent) start() {
	// Initial probe fetch
	if err := a.refreshProbes(); err != nil {
		log.Fatalf("Failed to fetch initial probes: %v", err)
	}

	// Send initial heartbeat
	a.sendHeartbeat()

	// Start background routines
	a.wg.Add(3)
	go a.heartbeatLoop()
	go a.probeRefreshLoop()
	go a.probeExecutionLoop()

	fmt.Printf("✓ Agent started with %d probes\n", len(a.probes))
}

// stop gracefully shuts down the agent
func (a *AdvancedMonitoringAgent) stop() {
	a.cancel()
	a.heartbeatTicker.Stop()
	a.probeRefreshTicker.Stop()
	
	// Send final heartbeat with unhealthy status
	nodeInfo := a.createNodeInfo()
	nodeInfo.Status = "stopping"
	if err := a.client.Monitoring.Heartbeat(context.Background(), nodeInfo); err != nil {
		// Log error but continue shutdown
		fmt.Printf("Warning: Failed to send final heartbeat: %v\n", err)
	}
	
	a.wg.Wait()
}

// heartbeatLoop sends periodic heartbeats
func (a *AdvancedMonitoringAgent) heartbeatLoop() {
	defer a.wg.Done()
	
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-a.heartbeatTicker.C:
			a.sendHeartbeat()
		}
	}
}

// probeRefreshLoop periodically refreshes the probe assignments
func (a *AdvancedMonitoringAgent) probeRefreshLoop() {
	defer a.wg.Done()
	
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-a.probeRefreshTicker.C:
			if err := a.refreshProbes(); err != nil {
				log.Printf("Failed to refresh probes: %v", err)
			}
		}
	}
}

// probeExecutionLoop executes probes based on their intervals
func (a *AdvancedMonitoringAgent) probeExecutionLoop() {
	defer a.wg.Done()
	
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()
	
	lastExecution := make(map[uint]time.Time)
	
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.probeMutex.RLock()
			probes := make([]*nexmonyx.ProbeAssignment, len(a.probes))
			copy(probes, a.probes)
			a.probeMutex.RUnlock()
			
			now := time.Now()
			var results []nexmonyx.ProbeExecutionResult
			
			for _, probe := range probes {
				if !probe.Enabled {
					continue
				}
				
				lastExec, exists := lastExecution[probe.ProbeID]
				if !exists || now.Sub(lastExec) >= time.Duration(probe.Interval)*time.Second {
					result := a.executeProbe(probe)
					results = append(results, result)
					lastExecution[probe.ProbeID] = now
				}
			}
			
			if len(results) > 0 {
				a.submitResults(results)
			}
		}
	}
}

// refreshProbes fetches the latest probe assignments
func (a *AdvancedMonitoringAgent) refreshProbes() error {
	probes, err := a.client.Monitoring.GetAssignedProbes(a.ctx, a.region)
	if err != nil {
		return err
	}
	
	a.probeMutex.Lock()
	a.probes = probes
	a.probeMutex.Unlock()
	
	fmt.Printf("✓ Refreshed probes: %d assigned\n", len(probes))
	return nil
}

// sendHeartbeat sends a heartbeat with current node information
func (a *AdvancedMonitoringAgent) sendHeartbeat() {
	nodeInfo := a.createNodeInfo()
	
	if err := a.client.Monitoring.Heartbeat(a.ctx, nodeInfo); err != nil {
		log.Printf("Failed to send heartbeat: %v", err)
	} else {
		fmt.Printf("✓ Heartbeat sent [Probes: %d, Success Rate: %.1f%%]\n", 
			nodeInfo.ProbesAssigned, nodeInfo.SuccessRate)
	}
}

// createNodeInfo creates current node information
func (a *AdvancedMonitoringAgent) createNodeInfo() nexmonyx.NodeInfo {
	a.stats.RLock()
	executed := a.stats.probesExecuted
	successful := a.stats.probesSuccessful
	failed := a.stats.probesFailed
	totalResponseTime := a.stats.totalResponseTime
	a.stats.RUnlock()
	
	a.probeMutex.RLock()
	probesAssigned := len(a.probes)
	a.probeMutex.RUnlock()
	
	var successRate float64
	if executed > 0 {
		successRate = float64(successful) / float64(executed) * 100
	}
	
	var avgResponseTime float64
	if successful > 0 {
		avgResponseTime = float64(totalResponseTime) / float64(successful)
	}
	
	hostname, _ := os.Hostname()
	
	return nexmonyx.NodeInfo{
		AgentID:            a.agentID,
		AgentVersion:       "1.0.0",
		Region:             a.region,
		Hostname:           hostname,
		IPAddress:          "10.0.1.100", // Mock IP
		Status:             "healthy",
		Uptime:             time.Since(time.Now().Add(-time.Hour)), // Mock uptime
		LastSeen:           time.Now(),
		ProbesAssigned:     probesAssigned,
		ProbesExecuted:     executed,
		ProbesSuccessful:   successful,
		ProbesFailed:       failed,
		SuccessRate:        successRate,
		AvgResponseTime:    avgResponseTime,
		MaxConcurrency:     10,
		SupportedTypes:     []string{"http", "https", "tcp", "icmp"},
		Capabilities:       []string{"tls_validation", "content_matching", "redirects"},
		Environment:        "production",
		Metadata: map[string]interface{}{
			"go_version": "1.24",
			"sdk_version": "1.2.0",
		},
	}
}

// executeProbe executes a single probe and returns the result
func (a *AdvancedMonitoringAgent) executeProbe(probe *nexmonyx.ProbeAssignment) nexmonyx.ProbeExecutionResult {
	start := time.Now()
	
	// Simulate probe execution based on type
	var result nexmonyx.ProbeExecutionResult
	
	switch probe.Type {
	case "http", "https":
		result = a.executeHTTPProbe(probe)
	case "tcp":
		result = a.executeTCPProbe(probe)
	case "icmp":
		result = a.executeICMPProbe(probe)
	default:
		result = nexmonyx.ProbeExecutionResult{
			ProbeID:    probe.ProbeID,
			ProbeUUID:  probe.ProbeUUID,
			ExecutedAt: start,
			Region:     probe.Region,
			Status:     "error",
			Error:      fmt.Sprintf("unsupported probe type: %s", probe.Type),
		}
	}
	
	// Update statistics
	a.stats.Lock()
	a.stats.probesExecuted++
	if result.Status == "success" {
		a.stats.probesSuccessful++
		a.stats.totalResponseTime += int64(result.ResponseTime)
	} else {
		a.stats.probesFailed++
	}
	a.stats.Unlock()
	
	return result
}

// executeHTTPProbe simulates HTTP/HTTPS probe execution
func (a *AdvancedMonitoringAgent) executeHTTPProbe(probe *nexmonyx.ProbeAssignment) nexmonyx.ProbeExecutionResult {
	// Simulate HTTP request timing
	// Safe conversion: uint → int for modulo operation (ProbeID is always < max int)
	responseTime := 50 + int(probe.ProbeID%200) // Mock response time
	
	result := nexmonyx.ProbeExecutionResult{
		ProbeID:       probe.ProbeID,
		ProbeUUID:     probe.ProbeUUID,
		ExecutedAt:    time.Now(),
		Region:        probe.Region,
		Status:        "success",
		ResponseTime:  responseTime,
		StatusCode:    200,
		DNSTime:       10,
		ConnectTime:   20,
		TLSTime:       15,
		FirstByteTime: responseTime - 30,
		TotalTime:     responseTime,
		ResponseSize:  2048,
		ContentMatch:  &[]bool{true}[0],
	}
	
	// Simulate occasional failures
	if probe.ProbeID%13 == 0 {
		result.Status = "failed"
		result.StatusCode = 500
		result.Error = "Internal server error"
		result.ContentMatch = &[]bool{false}[0]
	}
	
	return result
}

// executeTCPProbe simulates TCP probe execution
func (a *AdvancedMonitoringAgent) executeTCPProbe(probe *nexmonyx.ProbeAssignment) nexmonyx.ProbeExecutionResult {
	// Safe conversion: perform modulo before conversion to avoid overflow
	responseTime := 20 + int(probe.ProbeID%50)
	
	return nexmonyx.ProbeExecutionResult{
		ProbeID:     probe.ProbeID,
		ProbeUUID:   probe.ProbeUUID,
		ExecutedAt:  time.Now(),
		Region:      probe.Region,
		Status:      "success",
		ResponseTime: responseTime,
		ConnectTime: responseTime,
		TotalTime:   responseTime,
	}
}

// executeICMPProbe simulates ICMP probe execution
func (a *AdvancedMonitoringAgent) executeICMPProbe(probe *nexmonyx.ProbeAssignment) nexmonyx.ProbeExecutionResult {
	// Safe conversion: perform modulo before conversion to avoid overflow
	responseTime := 5 + int(probe.ProbeID%20)
	
	return nexmonyx.ProbeExecutionResult{
		ProbeID:      probe.ProbeID,
		ProbeUUID:    probe.ProbeUUID,
		ExecutedAt:   time.Now(),
		Region:       probe.Region,
		Status:       "success",
		ResponseTime: responseTime,
		TotalTime:    responseTime,
	}
}

// submitResults submits probe execution results
func (a *AdvancedMonitoringAgent) submitResults(results []nexmonyx.ProbeExecutionResult) {
	if err := a.client.Monitoring.SubmitResults(a.ctx, results); err != nil {
		log.Printf("Failed to submit probe results: %v", err)
	} else {
		fmt.Printf("✓ Submitted %d probe results\n", len(results))
	}
}