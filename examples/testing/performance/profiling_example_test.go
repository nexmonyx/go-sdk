package main

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/nexmonyx/go-sdk/v2"
)

// ExampleProfilingPatterns demonstrates memory and CPU profiling techniques
// These examples show how to profile SDK operations
// Run with: go test -v ./performance

// TestMemoryProfiling demonstrates how to profile memory usage
// Run with: go test -v -run TestMemoryProfiling -memprofile=mem.prof ./performance
// Analyze: go tool pprof mem.prof
func TestMemoryProfiling(t *testing.T) {
	// Get initial memory stats
	baseline := runtime.MemStats{}
	runtime.ReadMemStats(&baseline)

	t.Logf("Initial allocations: %d", baseline.Alloc)
	t.Logf("Initial heap objects: %d", baseline.HeapObjects)

	// Simulate typical SDK operations
	for i := 0; i < 1000; i++ {
		client, _ := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: "https://api.example.com",
			Auth: nexmonyx.AuthConfig{
				Token: "test-token",
			},
		})
		_ = client

		// Simulate query generation
		opts := &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		}
		_ = opts.ToQuery()
	}

	// Force garbage collection to get accurate measurements
	runtime.GC()

	// Get final memory stats
	current := runtime.MemStats{}
	runtime.ReadMemStats(&current)

	allocated := current.Alloc - baseline.Alloc
	allocatedMB := float64(allocated) / 1024 / 1024

	t.Logf("Memory allocated: %d bytes (%.2f MB)", allocated, allocatedMB)
	t.Logf("Current heap objects: %d", current.HeapObjects)
	t.Logf("Total allocations: %d", current.Mallocs)
	t.Logf("Total deallocations: %d", current.Frees)

	// Report if memory usage is unexpectedly high
	if allocatedMB > 10 {
		t.Logf("WARNING: High memory allocation detected (%.2f MB)", allocatedMB)
	}
}

// TestHeapGrowthDetection demonstrates detecting memory leaks
// If heap keeps growing after operations complete, there may be a leak
func TestHeapGrowthDetection(t *testing.T) {
	measurements := make([]uint64, 5)

	for round := 0; round < 5; round++ {
		// Run operations
		for i := 0; i < 100; i++ {
			client, _ := nexmonyx.NewClient(&nexmonyx.Config{
				BaseURL: "https://api.example.com",
				Auth: nexmonyx.AuthConfig{
					Token: "test-token",
				},
			})
			_ = client
		}

		// Force GC
		runtime.GC()
		time.Sleep(100 * time.Millisecond)

		// Record heap size
		m := runtime.MemStats{}
		runtime.ReadMemStats(&m)
		measurements[round] = m.HeapAlloc

		t.Logf("Round %d: Heap size: %d MB", round+1, m.HeapAlloc/1024/1024)
	}

	// Check for leaks: heap size should stabilize
	lastGrowth := measurements[4] - measurements[3]
	if lastGrowth > 1000000 { // > 1 MB growth between rounds
		t.Logf("WARNING: Possible memory leak detected (%.2f MB growth)", float64(lastGrowth)/1024/1024)
	}
}

// TestGarbageCollectionPauses measures GC impact
// Shows pause times and frequency of garbage collection
func TestGarbageCollectionPauses(t *testing.T) {
	type gcMetrics struct {
		startCount   uint32
		startPauses  [256]uint64
		endCount     uint32
		pauseDurations []time.Duration
	}

	metrics := gcMetrics{
		pauseDurations: make([]time.Duration, 0),
	}

	// Record initial GC state
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	metrics.startCount = m.NumGC
	copy(metrics.startPauses[:], m.PauseNs[:])

	// Run operations
	startTime := time.Now()
	for i := 0; i < 10000; i++ {
		opts := &nexmonyx.ListOptions{
			Page:   1,
			Limit:  25,
			Search: "test",
			Sort:   "name",
			Order:  "asc",
		}
		_ = opts.ToQuery()
	}
	elapsed := time.Since(startTime)

	// Record final GC state
	m2 := runtime.MemStats{}
	runtime.ReadMemStats(&m2)
	metrics.endCount = m2.NumGC
	gcRuns := metrics.endCount - metrics.startCount

	// Calculate pause times
	totalPauseDuration := time.Duration(0)
	for i := uint32(0); i < gcRuns; i++ {
		idx := (metrics.startCount + i) % 256
		pauseNs := m2.PauseNs[idx]
		totalPauseDuration += time.Duration(pauseNs)
	}

	avgPause := time.Duration(0)
	if gcRuns > 0 {
		avgPause = totalPauseDuration / time.Duration(gcRuns)
	}

	t.Logf("Total time: %v", elapsed)
	t.Logf("GC runs: %d", gcRuns)
	t.Logf("Total GC pause time: %v", totalPauseDuration)
	t.Logf("Average pause time: %v", avgPause)
	t.Logf("GC frequency: %.2f runs/sec", float64(gcRuns)/elapsed.Seconds())
}

// TestAllocationRateMeasurement measures how fast allocations occur
func TestAllocationRateMeasurement(t *testing.T) {
	baseline := runtime.MemStats{}
	runtime.ReadMemStats(&baseline)

	startTime := time.Now()
	duration := time.Second

	allocCount := 0
	for time.Since(startTime) < duration {
		opts := &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		}
		_ = opts.ToQuery()
		allocCount++
	}

	elapsed := time.Since(startTime)

	current := runtime.MemStats{}
	runtime.ReadMemStats(&current)

	allocated := current.Alloc - baseline.Alloc
	rate := float64(allocated) / elapsed.Seconds() / 1024 / 1024 // MB/sec

	t.Logf("Operations: %d in %v", allocCount, elapsed)
	t.Logf("Throughput: %.2f ops/sec", float64(allocCount)/elapsed.Seconds())
	t.Logf("Memory allocation rate: %.2f MB/sec", rate)
}

// TestConcurrentMemoryUsage measures memory under concurrent load
// Shows heap behavior with multiple goroutines
func TestConcurrentMemoryUsage(t *testing.T) {
	numGoroutines := 100
	opsPerGoroutine := 1000

	baseline := runtime.MemStats{}
	runtime.ReadMemStats(&baseline)

	// Run concurrent operations
	done := make(chan bool, numGoroutines)
	for g := 0; g < numGoroutines; g++ {
		go func() {
			for i := 0; i < opsPerGoroutine; i++ {
				opts := &nexmonyx.ListOptions{
					Page:  1,
					Limit: 25,
				}
				_ = opts.ToQuery()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for g := 0; g < numGoroutines; g++ {
		<-done
	}

	runtime.GC()

	current := runtime.MemStats{}
	runtime.ReadMemStats(&current)

	allocated := current.Alloc - baseline.Alloc
	allocatedMB := float64(allocated) / 1024 / 1024

	t.Logf("Total operations: %d", numGoroutines*opsPerGoroutine)
	t.Logf("Memory allocated: %.2f MB", allocatedMB)
	t.Logf("Average per operation: %d bytes", allocated/(uint64(numGoroutines*opsPerGoroutine)))
}

// TestContextCancellationMemory verifies context cancellation doesn't leak memory
func TestContextCancellationMemory(t *testing.T) {
	baseline := runtime.MemStats{}
	runtime.ReadMemStats(&baseline)

	// Create and cancel many contexts
	for i := 0; i < 10000; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_ = ctx
	}

	runtime.GC()

	current := runtime.MemStats{}
	runtime.ReadMemStats(&current)

	allocated := current.Alloc - baseline.Alloc
	allocatedKB := float64(allocated) / 1024

	t.Logf("Memory for 10k contexts: %.2f KB", allocatedKB)

	if allocated > 5000000 { // > 5 MB
		t.Logf("WARNING: High context memory usage detected")
	}
}

// ProfileMemoryUsage provides a helper for profiling specific code sections
// Usage: defer ProfileMemoryUsage("operation name")()
func ProfileMemoryUsage(name string) func() {
	baseline := runtime.MemStats{}
	runtime.ReadMemStats(&baseline)
	startTime := time.Now()

	return func() {
		current := runtime.MemStats{}
		runtime.ReadMemStats(&current)

		elapsed := time.Since(startTime)
		allocated := current.Alloc - baseline.Alloc
		allocatedMB := float64(allocated) / 1024 / 1024

		fmt.Printf("[PROFILE] %s: %.2f ms, %.2f MB\n", name, elapsed.Seconds()*1000, allocatedMB)
	}
}

// Example of using the profiling helper
