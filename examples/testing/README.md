# Nexmonyx Go SDK - Testing Examples

Comprehensive examples demonstrating how to test the Nexmonyx Go SDK in various scenarios.

## Overview

This directory contains testing examples organized by category:

- **Unit Testing** - Basic client operations, assertions, and error handling
- **Integration Testing** - Testing with mock API servers and real API endpoints
- **Performance Testing** - Benchmarking, profiling, and load testing

## Directory Structure

```
examples/testing/
├── unit/
│   └── basic_test.go                    # Unit testing patterns
├── integration/
│   └── mock_api_test.go                 # Integration testing patterns
├── performance/
│   ├── benchmarking_example_test.go     # Benchmark examples
│   ├── profiling_example_test.go        # Memory/CPU profiling examples
│   └── load_testing_example_test.go     # Load testing patterns
└── README.md                             # This file
```

## Quick Start

### Unit Testing

Basic client creation and testing patterns:

```bash
cd examples/testing/unit
go test -v basic_test.go
```

**What you'll learn:**
- Creating clients with different authentication methods
- Table-driven testing patterns
- Using testify assertions
- Error handling patterns

### Integration Testing

Testing with mock API server:

```bash
# Start mock API server (requires Docker and docker-compose)
docker-compose -f tests/integration/docker/docker-compose.yml up -d

# Run integration tests
cd examples/testing/integration
INTEGRATION_TESTS=true go test -v mock_api_test.go
```

**What you'll learn:**
- Testing with mock API endpoints
- CRUD operation workflows
- Concurrent operation testing
- Error handling in API context

### Performance Testing

Benchmarking and profiling examples:

```bash
# Run benchmarks
cd examples/testing/performance
go test -bench=. -benchmem benchmarking_example.go

# Run profiling examples
go test -v profiling_example.go

# Run load testing
go test -v -run TestSimpleConcurrentLoad load_testing_example.go
```

**What you'll learn:**
- Benchmarking SDK operations
- Memory profiling techniques
- CPU profiling patterns
- Load testing strategies

## Unit Testing Examples

### File: `unit/basic_test.go`

#### Client Creation Tests
```go
// Test JWT authentication
client, err := nexmonyx.NewClient(&nexmonyx.Config{
    BaseURL: "https://api.example.com",
    Auth: nexmonyx.AuthConfig{
        Token: "eyJhbGci...",
    },
})
```

#### Available Tests
- `TestClientCreation` - Basic client creation with JWT
- `TestClientWithAPIKeyAuth` - API key authentication
- `TestClientWithServerCredentials` - Server/agent authentication
- `TestClientAuthMethodSwitching` - Changing auth methods
- `TestResponseTypeParsing` - Handling response types
- `TestErrorHandling` - Error handling patterns
- `TestListOptions` - Pagination options
- `TestTableDrivenTests` - Table-driven test patterns
- `BenchmarkClientCreation` - Performance benchmark

## Integration Testing Examples

### File: `integration/mock_api_test.go`

#### Running Tests

Set environment variable to enable integration tests:

```bash
export INTEGRATION_TESTS=true
go test -v ./integration
```

#### Available Tests
- `TestServerListingWithMockAPI` - List servers from mock API
- `TestOrganizationCRUD` - CRUD workflow (Create, Read, Update, Delete)
- `TestErrorHandlingWithMockAPI` - API error handling
- `TestConcurrentOperations` - Multiple concurrent API calls
- `TestMetricsSubmissionWithMockAPI` - Metrics submission workflow
- `TestDataValidation` - Input validation patterns
- `TestRealAPIIntegration` - Test against real dev API (optional)

#### Running Against Real API

```bash
export INTEGRATION_TESTS=true
export INTEGRATION_TEST_MODE=dev
export INTEGRATION_TEST_API_URL=https://api-dev.nexmonyx.com
export INTEGRATION_TEST_AUTH_TOKEN=your-jwt-token
go test -v ./integration
```

## Performance Testing Examples

### File: `performance/benchmarking_example_test.go`

Demonstrates benchmarking various SDK operations:

```bash
# Run all benchmarks
go test -bench=. -benchmem benchmarking_example.go

# Run specific benchmark
go test -bench=BenchmarkClientCreation -benchmem benchmarking_example.go

# Compare with baseline
go test -bench=. -benchmem benchmarking_example.go > new.txt
benchstat old.txt new.txt
```

#### Available Benchmarks
- `BenchmarkClientCreation` - JWT client creation
- `BenchmarkClientWithAPIKey` - API key authentication
- `BenchmarkClientWithServerCredentials` - Server credentials
- `BenchmarkQueryGeneration` - List options query conversion
- `BenchmarkJSONMarshaling` - JSON serialization
- `BenchmarkAuthMethodSwitching` - Auth method changes
- `BenchmarkConcurrentClients` - Multiple concurrent clients
- `BenchmarkContextCreation` - Context overhead
- `BenchmarkParallelListOptions` - Concurrent query generation
- `BenchmarkMemoryAllocation` - Memory allocation patterns

### File: `performance/profiling_example_test.go`

Memory and CPU profiling examples:

```bash
# Run profiling tests
go test -v profiling_example.go

# Generate memory profile
go test -v -memprofile=mem.prof profiling_example.go
go tool pprof mem.prof

# Analyze allocations
go tool pprof -alloc_space mem.prof
```

#### Available Tests
- `TestMemoryProfiling` - Basic memory usage tracking
- `TestHeapGrowthDetection` - Detecting memory leaks
- `TestGarbageCollectionPauses` - GC impact measurement
- `TestAllocationRateMeasurement` - Allocation speed
- `TestConcurrentMemoryUsage` - Memory under load
- `TestContextCancellationMemory` - Context memory impact

### File: `performance/load_testing_example_test.go`

Load testing patterns:

```bash
# Run load tests (5-10 seconds each)
go test -v -run TestSimpleConcurrentLoad load_testing_example.go
go test -v -run TestSustainedLoad load_testing_example.go
go test -v -run TestRampUpLoad load_testing_example.go

# Run all load tests with timeout
go test -v -timeout 5m ./performance
```

#### Available Load Tests
- `TestSimpleConcurrentLoad` - Basic concurrent operations
- `TestSustainedLoad` - Extended load test
- `TestRampUpLoad` - Gradually increasing load
- `TestErrorRateUnderLoad` - Error behavior under stress
- `TestConcurrentClientCreation` - Client creation overhead
- `TestSpikeyLoad` - Traffic spike simulation
- `TestContextDeadlineUnderLoad` - Deadline handling under load

## Common Patterns

### Table-Driven Tests

```go
tests := []struct {
    name      string
    input     interface{}
    wantValid bool
}{
    {"Valid case", "test", true},
    {"Invalid case", "", false},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test logic
    })
}
```

### Concurrent Testing

```go
var wg sync.WaitGroup
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // parallel operation
    }()
}
wg.Wait()
```

### Memory Profiling

```bash
# Generate profile
go test -memprofile=mem.prof -v ./performance

# Analyze profile
go tool pprof -http=:8080 mem.prof
```

### Benchmarking

```bash
# Run with memory tracking
go test -bench=. -benchmem ./performance

# Compare benchmarks
benchstat baseline.txt current.txt
```

## Environment Setup

### Docker Compose for Integration Tests

Start the mock API server:

```bash
cd tests/integration/docker
docker-compose up -d
```

Stop the server:

```bash
docker-compose down
```

Check server status:

```bash
curl http://localhost:8080/health
```

### Required Environment Variables

For integration tests:
- `INTEGRATION_TESTS=true` - Enable integration tests
- `INTEGRATION_TEST_MODE=dev` - Test against dev API (optional)
- `INTEGRATION_TEST_API_URL` - API endpoint URL (optional)
- `INTEGRATION_TEST_AUTH_TOKEN` - Authentication token (optional)

## Best Practices

### Unit Testing
1. Test one thing per test
2. Use table-driven tests for multiple cases
3. Use testify assertions for clarity
4. Mock external dependencies
5. Test error paths, not just success

### Integration Testing
1. Use mock servers when possible
2. Keep tests independent
3. Clean up test data after tests
4. Use environment variables for configuration
5. Skip slow tests with `-short` flag

### Performance Testing
1. Use `b.ReportAllocs()` for memory tracking
2. Reset timer after setup: `b.ResetTimer()`
3. Run benchmarks multiple times: `benchstat`
4. Profile actual code paths
5. Compare before and after optimizations

## Troubleshooting

### Integration Tests Not Running

```bash
# Verify environment variable is set
echo $INTEGRATION_TESTS

# Set it if not
export INTEGRATION_TESTS=true

# Run with verbose output
go test -v -run TestServerListingWithMockAPI ./integration
```

### Mock API Server Not Responding

```bash
# Check if server is running
docker-compose -f tests/integration/docker/docker-compose.yml ps

# Check server logs
docker-compose -f tests/integration/docker/docker-compose.yml logs

# Restart server
docker-compose -f tests/integration/docker/docker-compose.yml restart
```

### Benchmarks Too Slow

```bash
# Run subset of benchmarks
go test -bench=BenchmarkClientCreation -benchmem ./performance

# Use timeout
go test -bench=. -benchmem -timeout=30s ./performance
```

### Memory Profile Generation Issues

```bash
# Check if pprof is available
go tool pprof -h

# Generate with verbose output
go test -v -memprofile=mem.prof profiling_example.go

# View profile
go tool pprof mem.prof
# Type: top
# Type: list TestMemoryProfiling
```

## References

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [pprof Documentation](https://github.com/google/pprof)
- [Go Benchmarking](https://golang.org/doc/effective_go#benchmarks)
- [Nexmonyx SDK Documentation](../../docs/)

## Next Steps

1. **Copy examples** - Use these as templates for your tests
2. **Run the tests** - Try each example with the commands above
3. **Modify for your use case** - Adapt examples to your specific API calls
4. **Add to CI/CD** - Integrate these tests into your pipeline
5. **Monitor performance** - Use profiling examples to track regressions

## Support

For questions about the SDK:
- Check the [TESTING.md](../../TESTING.md) guide
- Review the [API Documentation](../../docs/)
- See the [PERFORMANCE.md](../../docs/PERFORMANCE.md) guide
- Check [benchmark reference](../../docs/BENCHMARK_REFERENCE.md)
