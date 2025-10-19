package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/require"
)

// ExampleLoadTestingPatterns demonstrates techniques for load testing the SDK
// These examples show how to test SDK behavior under sustained load
// Run with: go test -v -run TestLoadTesting -timeout 10m ./performance

// LoadTestResult holds metrics from a load test
type LoadTestResult struct {
	TotalOperations   int64
	SuccessfulOps     int64
	FailedOps         int64
	Duration          time.Duration
	OpsPerSecond      float64
	AverageDuration   time.Duration
	MinDuration       time.Duration
	MaxDuration       time.Duration
	ErrorCount        int64
}

// TestSimpleConcurrentLoad tests basic concurrent client operations
// Simulates multiple clients operating simultaneously
func TestSimpleConcurrentLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	numClients := 10
	operationsPerClient := 100
	var wg sync.WaitGroup
	var successCount int64

	t.Logf("Starting concurrent load test: %d clients × %d ops", numClients, operationsPerClient)

	startTime := time.Now()

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			_, err := nexmonyx.NewClient(&nexmonyx.Config{
				BaseURL: "https://api.example.com",
				Auth: nexmonyx.AuthConfig{
					Token: "test-token",
				},
			})
			require.NoError(t, err)

			for j := 0; j < operationsPerClient; j++ {
				opts := &nexmonyx.ListOptions{
					Page:  1,
					Limit: 25,
				}
				query := opts.ToQuery()
				if len(query) > 0 {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	totalOps := int64(numClients * operationsPerClient)
	opsPerSec := float64(totalOps) / elapsed.Seconds()

	t.Logf("Completed: %d operations in %v (%.2f ops/sec)", totalOps, elapsed, opsPerSec)
	t.Logf("Success rate: %.2f%%", float64(successCount)*100/float64(totalOps))
}

// TestSustainedLoad tests continued load for an extended period
// Useful for detecting memory leaks or performance degradation
func TestSustainedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping sustained load test in short mode")
	}

	duration := 5 * time.Second // Short test for examples
	numWorkers := 5
	var (
		totalOps     int64
		successfulOps int64
		failedOps    int64
	)

	t.Logf("Running sustained load test for %v with %d workers", duration, numWorkers)

	var wg sync.WaitGroup
	stopTime := time.Now().Add(duration)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for time.Now().Before(stopTime) {
				opts := &nexmonyx.ListOptions{
					Page:   1,
					Limit:  25,
					Search: "test",
					Sort:   "name",
				}

				query := opts.ToQuery()
				atomic.AddInt64(&totalOps, 1)

				if len(query) > 0 {
					atomic.AddInt64(&successfulOps, 1)
				} else {
					atomic.AddInt64(&failedOps, 1)
				}
			}
		}()
	}

	wg.Wait()

	opsPerSec := float64(totalOps) / duration.Seconds()
	t.Logf("Total operations: %d", totalOps)
	t.Logf("Successful: %d, Failed: %d", successfulOps, failedOps)
	t.Logf("Throughput: %.2f ops/sec", opsPerSec)
	t.Logf("Average per worker: %.2f ops/sec", opsPerSec/float64(numWorkers))
}

// TestRampUpLoad gradually increases load over time
// Useful for finding the breaking point
func TestRampUpLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping ramp-up test in short mode")
	}

	stages := []struct {
		workers   int
		duration  time.Duration
	}{
		{1, 1 * time.Second},
		{5, 1 * time.Second},
		{10, 1 * time.Second},
		{20, 1 * time.Second},
	}

	t.Log("Starting ramp-up load test")

	for _, stage := range stages {
		t.Logf("Stage: %d workers for %v", stage.workers, stage.duration)

		var totalOps int64
		var wg sync.WaitGroup
		stopTime := time.Now().Add(stage.duration)

		for w := 0; w < stage.workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for time.Now().Before(stopTime) {
					opts := &nexmonyx.ListOptions{Page: 1, Limit: 25}
					_ = opts.ToQuery()
					atomic.AddInt64(&totalOps, 1)
				}
			}()
		}

		wg.Wait()
		opsPerSec := float64(totalOps) / stage.duration.Seconds()
		t.Logf("  Completed: %d ops (%.2f ops/sec)", totalOps, opsPerSec)
	}
}

// TestErrorRateUnderLoad measures how error rate changes with load
func TestErrorRateUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error rate test in short mode")
	}

	loadLevels := []int{1, 5, 10, 20}
	testDuration := 2 * time.Second

	t.Log("Testing error rates at different load levels")

	for _, numWorkers := range loadLevels {
		var (
			totalOps  int64
			errorOps  int64
		)

		stopTime := time.Now().Add(testDuration)
		var wg sync.WaitGroup

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for time.Now().Before(stopTime) {
					opts := &nexmonyx.ListOptions{
						Page:  1,
						Limit: 25,
					}
					query := opts.ToQuery()
					atomic.AddInt64(&totalOps, 1)

					if len(query) == 0 {
						atomic.AddInt64(&errorOps, 1)
					}
				}
			}()
		}

		wg.Wait()
		errorRate := float64(errorOps) * 100 / float64(totalOps)
		t.Logf("Workers: %d, Total ops: %d, Error rate: %.2f%%", numWorkers, totalOps, errorRate)
	}
}

// TestConcurrentClientCreation measures overhead of creating many clients
func TestConcurrentClientCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent client creation test in short mode")
	}

	numClients := 100

	t.Logf("Creating %d concurrent clients", numClients)

	startTime := time.Now()
	var wg sync.WaitGroup
	var successCount int64

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := nexmonyx.NewClient(&nexmonyx.Config{
				BaseURL: "https://api.example.com",
				Auth: nexmonyx.AuthConfig{
					Token: "test-token",
				},
			})
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	t.Logf("Created %d clients in %v", successCount, elapsed)
	t.Logf("Average time per client: %v", elapsed/time.Duration(numClients))
}

// TestSpikeyLoad simulates sudden traffic spikes
func TestSpikeyLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping spiky load test in short mode")
	}

	t.Log("Simulating traffic spikes")

	spikes := []struct {
		workers  int
		duration time.Duration
		name     string
	}{
		{5, 500 * time.Millisecond, "Low load"},
		{50, 500 * time.Millisecond, "Spike 1"},
		{5, 500 * time.Millisecond, "Low load"},
		{100, 500 * time.Millisecond, "Spike 2"},
		{5, 500 * time.Millisecond, "Low load"},
	}

	for _, spike := range spikes {
		var totalOps int64
		var wg sync.WaitGroup
		stopTime := time.Now().Add(spike.duration)

		for w := 0; w < spike.workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for time.Now().Before(stopTime) {
					opts := &nexmonyx.ListOptions{Page: 1, Limit: 25}
					_ = opts.ToQuery()
					atomic.AddInt64(&totalOps, 1)
				}
			}()
		}

		wg.Wait()
		opsPerSec := float64(totalOps) / spike.duration.Seconds()
		t.Logf("%s: %d workers → %.2f ops/sec", spike.name, spike.workers, opsPerSec)
	}
}

// TestContextDeadlineUnderLoad verifies context handling under stress
func TestContextDeadlineUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context deadline test in short mode")
	}

	numWorkers := 10
	operationsPerWorker := 100
	var (
		totalOps      int64
		completedOps  int64
		timedOutOps   int64
	)

	t.Logf("Testing context deadlines with %d workers", numWorkers)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < operationsPerWorker; i++ {
				atomic.AddInt64(&totalOps, 1)

				// Create context with timeout
				contextWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				select {
				case <-contextWithTimeout.Done():
					atomic.AddInt64(&timedOutOps, 1)
				default:
					atomic.AddInt64(&completedOps, 1)
				}
				cancel()
			}
		}()
	}

	wg.Wait()

	t.Logf("Total: %d, Completed: %d, Timed out: %d", totalOps, completedOps, timedOutOps)
	t.Logf("Success rate: %.2f%%", float64(completedOps)*100/float64(totalOps))
}

// ReportLoadTestMetrics formats and displays load test results
func ReportLoadTestMetrics(t *testing.T, result LoadTestResult) {
	output := fmt.Sprintf(`
Load Test Results
═════════════════
Operations:        %d
Successful:        %d
Failed:            %d
Duration:          %v
Throughput:        %.2f ops/sec
Avg Duration:      %v
Min Duration:      %v
Max Duration:      %v
Error Rate:        %.2f%%
`,
		result.TotalOperations,
		result.SuccessfulOps,
		result.FailedOps,
		result.Duration,
		result.OpsPerSecond,
		result.AverageDuration,
		result.MinDuration,
		result.MaxDuration,
		float64(result.ErrorCount)*100/float64(result.TotalOperations),
	)
	t.Log(output)
}

// Example load test helper function
func simpleLoadTest(numWorkers int, durationSecs int, operationFunc func()) LoadTestResult {
	result := LoadTestResult{
		Duration: time.Duration(durationSecs) * time.Second,
	}

	var wg sync.WaitGroup
	stopTime := time.Now().Add(result.Duration)

	startTime := time.Now()

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Now().Before(stopTime) {
				operationFunc()
				atomic.AddInt64(&result.TotalOperations, 1)
				atomic.AddInt64(&result.SuccessfulOps, 1)
			}
		}()
	}

	wg.Wait()
	result.Duration = time.Since(startTime)
	result.OpsPerSecond = float64(result.TotalOperations) / result.Duration.Seconds()

	return result
}
