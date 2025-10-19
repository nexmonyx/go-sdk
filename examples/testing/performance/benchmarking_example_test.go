package main

import (
	"context"
	"testing"

	"github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/require"
)

// ExampleBenchmarkingPatterns demonstrates how to benchmark SDK operations
// These examples show common patterns for measuring performance
// Run with: go test -bench=. -benchmem ./performance

// BenchmarkClientCreation measures client creation performance
// Output shows time per operation and memory allocations
func BenchmarkClientCreation(b *testing.B) {
	b.ReportAllocs() // Enable allocation reporting
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: "https://api.example.com",
			Auth: nexmonyx.AuthConfig{
				Token: "test-token",
			},
		})
	}
}

// BenchmarkClientWithAPIKey measures client creation with API key auth
func BenchmarkClientWithAPIKey(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: "https://api.example.com",
			Auth: nexmonyx.AuthConfig{
				APIKey:    "test-key",
				APISecret: "test-secret",
			},
		})
	}
}

// BenchmarkClientWithServerCredentials measures agent authentication
func BenchmarkClientWithServerCredentials(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: "https://api.example.com",
			Auth: nexmonyx.AuthConfig{
				ServerUUID:   "server-uuid",
				ServerSecret: "server-secret",
			},
		})
	}
}

// BenchmarkQueryGeneration measures query parameter conversion
// This is a hot path called on every List() operation
func BenchmarkQueryGeneration(b *testing.B) {
	b.ReportAllocs()

	opts := &nexmonyx.ListOptions{
		Page:   1,
		Limit:  25,
		Search: "test-search",
		Sort:   "name",
		Order:  "asc",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opts.ToQuery()
	}
}

// BenchmarkJSONMarshaling measures JSON serialization performance
func BenchmarkJSONMarshaling(b *testing.B) {
	b.ReportAllocs()

	// Create a simple value to marshal (SDK models handle JSON encoding)
	// In real code, you'd test your actual API response structures
	data := map[string]interface{}{
		"server_uuid": "server-123",
		"hostname":    "web-01.example.com",
		"os":          "Linux",
		"os_version":  "5.10.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate map access
		_ = data["hostname"]
	}
}

// BenchmarkAuthMethodSwitching measures cost of changing authentication
func BenchmarkAuthMethodSwitching(b *testing.B) {
	b.ReportAllocs()

	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.example.com",
		Auth: nexmonyx.AuthConfig{
			Token: "initial-token",
		},
	})
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.WithAPIKey("key", "secret")
	}
}

// BenchmarkConcurrentClients measures performance with multiple concurrent clients
// Useful for understanding overhead of maintaining many client instances
func BenchmarkConcurrentClients(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = nexmonyx.NewClient(&nexmonyx.Config{
				BaseURL: "https://api.example.com",
				Auth: nexmonyx.AuthConfig{
					Token: "test-token",
				},
			})
		}
	})
}

// BenchmarkContextCreation measures context creation overhead
func BenchmarkContextCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, cancel := context.WithCancel(context.Background())
		cancel()
	}
}

// BenchmarkParallelListOptions measures query generation under concurrent load
// Simulates multiple goroutines generating queries simultaneously
func BenchmarkParallelListOptions(b *testing.B) {
	opts := &nexmonyx.ListOptions{
		Page:   1,
		Limit:  25,
		Search: "test",
		Sort:   "name",
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = opts.ToQuery()
		}
	})
}

// BenchmarkMemoryAllocation demonstrates tracking memory allocations
// This benchmark shows how to profile specific operations
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate typical API operation allocations
		data := make([]byte, 1024)
		data = append(data, []byte("test-data")...)
		_ = data
	}
}

// Example of running benchmarks with comparison to baseline
// Save baseline: go test -bench=BenchmarkClientCreation -benchmem ./performance > old.txt
// Run optimized: go test -bench=BenchmarkClientCreation -benchmem ./performance > new.txt
// Compare: benchstat old.txt new.txt
//
// Expected output shows:
// - name: Benchmark name
// - Old time: Previous performance
// - New time: Current performance
// - % difference: Speed improvement/regression
// - Memory (B/op): Bytes per operation
// - Allocations (allocs/op): Number of memory allocations
