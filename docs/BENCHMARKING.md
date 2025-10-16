# Nexmonyx Go SDK - Benchmarking Guide

Comprehensive guide to understanding and running performance benchmarks for the Nexmonyx Go SDK.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Running Benchmarks](#running-benchmarks)
- [Benchmark Suites](#benchmark-suites)
- [Understanding Results](#understanding-results)
- [Performance Baselines](#performance-baselines)
- [Profiling](#profiling)
- [Comparing Results](#comparing-results)
- [Optimization Guidelines](#optimization-guidelines)
- [CI/CD Integration](#cicd-integration)

## Overview

The benchmarking suite for the Nexmonyx Go SDK measures performance of:

- **Client Operations** - Initialization, configuration, connection management
- **Data Serialization** - JSON marshaling/unmarshaling, model allocation
- **Error Handling** - Error creation, type checking, propagation
- **Concurrent Operations** - Parallel execution, synchronization, resource contention
- **Memory Efficiency** - Allocation patterns, garbage collection behavior

### Why Benchmarking Matters

Performance regressions can happen silently. The benchmark suite helps:

1. **Detect Regressions** - Catch performance degradation early
2. **Establish Baselines** - Know current performance characteristics
3. **Guide Optimization** - Identify actual bottlenecks, not guesses
4. **Enable Comparison** - Compare different implementation approaches
5. **Document Performance** - Track changes over time

## Quick Start

### Run All Benchmarks

```bash
# Run all benchmarks with default settings
go test -bench=. -benchmem ./...

# Run only benchmarks (skip tests)
go test -bench=. -run=^$ -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkClientCreation -benchmem ./...
```

### View Results

```bash
# Run and save results to file
go test -bench=. -benchmem ./... > bench.txt

# View baseline comparison
go test -bench=. -benchmem -benchstat=old.txt ./...
```

## Running Benchmarks

### Basic Execution

```bash
# Standard benchmark run
go test -bench=. -benchmem ./...

# Flags:
#   -bench       Pattern to match benchmark names
#   -benchmem    Report memory allocations
#   -benchtime   Duration per benchmark (default 1s)
#   -benchstat   Compare with previous results
#   -count       Run benchmarks N times
#   -timeout     Overall timeout
```

### Advanced Options

```bash
# Run each benchmark 5 times for stability
go test -bench=. -benchmem -count=5 ./...

# Run for longer durations (more stable results)
go test -bench=. -benchmem -benchtime=10s ./...

# Run specific benchmark patterns
go test -bench=Concurrent -benchmem ./...

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./...

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./...

# Run with race detector (slower but catches races)
go test -bench=. -race ./...

# Run with short timeout (quick validation)
go test -bench=. -benchtime=100ms ./...
```

### Batch Benchmark Execution

```bash
# Run benchmarks multiple times and save results
for i in {1..5}; do
  go test -bench=. -benchmem ./... >> results.txt
done

# Compare results
benchstat results.txt
```

## Benchmark Suites

### 1. Client Benchmarks (`benchmark_client_test.go`)

Tests client lifecycle and operations.

```bash
# Run client benchmarks
go test -bench=Client -benchmem ./...
```

**Benchmarks:**
- `BenchmarkClientCreation/WithJWTToken` - Client creation with JWT auth
- `BenchmarkClientCreation/WithAPIKeySecret` - Client creation with API key auth
- `BenchmarkClientCreation/WithServerCredentials` - Client creation with server credentials
- `BenchmarkClientCreation/WithDebugMode` - Client creation with debug mode enabled
- `BenchmarkClientCreation/ConcurrentCreation` - Parallel client creation
- `BenchmarkClientConfiguration/*` - Configuration operations
- `BenchmarkConcurrentClientOperations/*` - Parallel client usage

**Expected Baseline:**
```
BenchmarkClientCreation/WithJWTToken-8        300,000 ns/op    ~400 B/op    10 allocs/op
```

### 2. Model Benchmarks (`benchmark_models_test.go`)

Tests data serialization and model handling.

```bash
# Run model benchmarks
go test -bench=Model -benchmem ./...
```

**Benchmarks:**
- `BenchmarkCustomTimeParsing/*` - Timestamp parsing with different formats
- `BenchmarkCustomTimeMarshaling/*` - Timestamp serialization
- `BenchmarkModelSerialization/*` - Marshal/unmarshal core models
- `BenchmarkLargePayloads/*` - Handling large data structures
- `BenchmarkModelAllocation/*` - Memory allocation patterns

**Performance Tips:**

Timestamp parsing is a hot path:
```
RFC3339        - ~100 ns (fastest)
RFC3339Nano    - ~150 ns
ISO8601        - ~180 ns
UnixTimestamp  - ~50 ns (fastest, if applicable)
```

### 3. Error Benchmarks (`benchmark_errors_test.go`)

Tests error handling performance.

```bash
# Run error benchmarks
go test -bench=Error -benchmem ./...
```

**Benchmarks:**
- `BenchmarkErrorProcessing/*` - Error creation
- `BenchmarkErrorTypeAssertion/*` - Type checking
- `BenchmarkErrorUnmarshaling/*` - Deserializing error responses
- `BenchmarkErrorFormatting/*` - Error string conversion
- `BenchmarkErrorPropagation/*` - Error wrapping

**Expected Baseline:**
```
BenchmarkErrorTypeAssertion/TypeSwitch-8    5,000,000 ns/op    0 B/op    0 allocs/op
```

### 4. Concurrency Benchmarks (`benchmark_concurrency_test.go`)

Tests parallel operation behavior.

```bash
# Run concurrency benchmarks
go test -bench=Concurrent -benchmem ./...
```

**Benchmarks:**
- `BenchmarkConcurrentOperations/*` - Parallel execution patterns
- `BenchmarkConcurrentMemory/*` - Memory under concurrent load
- `BenchmarkConcurrentSynchronization/*` - Lock and sync overhead
- `BenchmarkConcurrentDataStructures/*` - Data structure contention
- `BenchmarkConcurrentLoadPatterns/*` - Realistic load scenarios
- `BenchmarkConcurrentResourceCleanup/*` - Cleanup efficiency

**Load Scenarios:**
- 10 concurrent goroutines - Light load
- 100 concurrent goroutines - Medium load
- 1000 concurrent goroutines - Heavy load

## Understanding Results

### Benchmark Output Format

```
BenchmarkClientCreation/WithJWTToken-8    300000    3500 ns/op    400 B/op    10 allocs/op
```

Breaking down each part:

| Part | Meaning |
|------|---------|
| `BenchmarkClientCreation/WithJWTToken` | Test name |
| `-8` | Number of CPU cores used |
| `300000` | Iterations run |
| `3500 ns/op` | Average time per iteration |
| `400 B/op` | Average memory allocated per iteration |
| `10 allocs/op` | Number of allocations per iteration |

### Interpreting Numbers

**Time (ns/op)**:
- 1 ns = 1 nanosecond
- 1 µs (microsecond) = 1,000 ns
- 1 ms (millisecond) = 1,000,000 ns

**Memory (B/op)**:
- B = bytes
- KB = 1024 bytes
- MB = 1048576 bytes

**Allocations (allocs/op)**:
- Lower is better
- Each allocation is costly (memory + GC pressure)
- Zero allocations is ideal

### Reading a Benchmark Report

```
Name                                    Iterations   Time/op    Memory/op   Allocs/op
BenchmarkClientCreation/WithJWTToken-8     300000    3.5 µs    400 B        10
BenchmarkClientCreation/WithDebugMode-8    250000    4.0 µs    450 B        11
BenchmarkClientConfiguration/GetConfig-8 5000000    0.2 µs      0 B         0
```

Analysis:
- Debug mode adds ~500ns and 1 allocation
- Configuration retrieval is cheap (~200ns, zero allocations)

## Performance Baselines

### Client Operations

| Operation | Time | Memory | Allocations |
|-----------|------|--------|-------------|
| Create with JWT | 3-4 µs | 400 B | 10 |
| Create with API Key | 3-4 µs | 400 B | 10 |
| Create with Server Creds | 3-4 µs | 400 B | 10 |
| Create with Debug | 4-5 µs | 450 B | 11 |
| Get Config | 100-200 ns | 0 B | 0 |
| Set Timeout | 50-100 ns | 0 B | 0 |

### Data Models

| Operation | Time | Memory | Allocations |
|-----------|------|--------|-------------|
| Organization Marshal | 1-2 µs | 500 B | 5 |
| Organization Unmarshal | 2-3 µs | 800 B | 8 |
| Metrics Marshal (large) | 20-30 µs | 5 KB | 20 |
| Metrics Unmarshal (large) | 30-50 µs | 8 KB | 30 |
| Timestamp Parse (RFC) | 100 ns | 0 B | 0 |
| Timestamp Parse (Other) | 200-500 ns | 0 B | 0 |

### Error Handling

| Operation | Time | Memory | Allocations |
|-----------|------|--------|-------------|
| Create APIError | 50 ns | 100 B | 1 |
| Type Switch (5 types) | 5-10 ns | 0 B | 0 |
| Unmarshal Error | 1-2 µs | 400 B | 4 |
| Error String | 200-500 ns | 0 B | 0 |

### Concurrent Operations

| Scenario | Time | Memory | Locks |
|----------|------|--------|-------|
| 10 concurrent ops | 100-500 ms | 100 KB | Low |
| 100 concurrent ops | 1-5 seconds | 1 MB | Medium |
| 1000 concurrent ops | 10-50 seconds | 10 MB | High |

## Profiling

### CPU Profiling

```bash
# Run with CPU profile
go test -bench=. -cpuprofile=cpu.prof ./...

# View profile
go tool pprof cpu.prof

# Common commands in pprof:
#   top        - Show top functions
#   list       - Show function source
#   web        - Visualize as graph (requires graphviz)
```

### Memory Profiling

```bash
# Run with memory profile (heap)
go test -bench=. -memprofile=mem.prof ./...

# View profile
go tool pprof mem.prof

# Allocations profile
go test -bench=. -memprofile=mem.prof -memprofilerate=1 ./...
```

### Lock Profiling

```bash
# Run with mutex contention profiling
go test -bench=. -mutexprofile=mutex.prof ./...

# View profile
go tool pprof mutex.prof
```

### Generate Flamegraph

```bash
# Record profile
go test -bench=. -cpuprofile=cpu.prof ./...

# Generate flamegraph
go tool pprof -http=:8080 cpu.prof
```

## Comparing Results

### Using benchstat

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Run benchmarks multiple times
go test -bench=. -benchmem -count=5 ./... > new.txt

# Compare with previous run
benchstat old.txt new.txt
```

### Output Example

```
name                          old time/op  new time/op  delta
ClientCreation/WithJWT-8      3.5µs ± 2%   3.6µs ± 3%   +2.86% (p=0.008 n=10+10)
ClientCreation/WithDebug-8    4.0µs ± 1%   4.2µs ± 2%   +5.00% (p=0.000 n=10+10)
ModelMarshal/Org-8            2.1µs ± 3%   2.2µs ± 2%   +4.76% (p=0.019 n=10+10)

name                          old alloc/op new alloc/op delta
ClientCreation/WithJWT-8       400B ± 0%    400B ± 0%   (all equal)
ModelMarshal/Org-8            520B ± 0%    520B ± 0%   (all equal)
```

### Manual Comparison

```bash
# Extract baseline
go test -bench=BenchmarkClientCreation -benchmem ./... | tee baseline.txt

# Make changes
# ... edit code ...

# Compare
go test -bench=BenchmarkClientCreation -benchmem ./... | tee current.txt

# Visual comparison
diff baseline.txt current.txt
```

## Optimization Guidelines

### When to Optimize

1. **Measurable Impact** - Optimization must show in benchmarks
2. **Bottleneck Confirmed** - Use profiling to identify actual hot spots
3. **Trade-offs Understood** - Consider memory, readability, maintainability
4. **Regression Tested** - Ensure correctness isn't sacrificed

### Optimization Techniques

### 1. Reduce Allocations

```go
// Before: Multiple allocations
func process(items []Item) []Result {
    var results []Result  // Uninitialized, grows
    for _, item := range items {
        results = append(results, processItem(item))
    }
    return results
}

// After: Pre-allocate
func process(items []Item) []Result {
    results := make([]Result, len(items))  // Single allocation
    for i, item := range items {
        results[i] = processItem(item)
    }
    return results
}
```

### 2. Optimize Hot Paths

```go
// Before: Generic parsing
func parseTime(s string) time.Time {
    for _, layout := range []string{layoutA, layoutB, layoutC} {
        if t, err := time.Parse(layout, s); err == nil {
            return t
        }
    }
    return time.Time{}
}

// After: Fast path first
func parseTime(s string) time.Time {
    // 80% of times are RFC3339
    if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
        return t
    }
    // Fallback for other formats
    return parseTimeGeneric(s)
}
```

### 3. Use sync.Pool for Reusable Objects

```go
// Before: Allocate each time
func marshal(v interface{}) ([]byte, error) {
    buf := &bytes.Buffer{}
    encoder := json.NewEncoder(buf)
    err := encoder.Encode(v)
    return buf.Bytes(), err
}

// After: Reuse buffers
var bufferPool = sync.Pool{
    New: func() interface{} {
        return &bytes.Buffer{}
    },
}

func marshal(v interface{}) ([]byte, error) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    encoder := json.NewEncoder(buf)
    err := encoder.Encode(v)
    return buf.Bytes(), err
}
```

### 4. Minimize Lock Contention

```go
// Before: Global lock
var mu sync.Mutex
var cache map[string]string

// After: Sharded locks
type Cache struct {
    shards []*CacheShard
}

type CacheShard struct {
    mu    sync.RWMutex
    items map[string]string
}
```

## CI/CD Integration

### GitHub Actions Workflow

```yaml
name: Performance Benchmarks

on:
  pull_request:
  push:
    branches:
      - main

jobs:
  bench:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run benchmarks
        run: go test -bench=. -benchmem ./... | tee bench_current.txt

      - name: Compare with main
        if: github.event_name == 'pull_request'
        run: |
          git checkout origin/main
          go test -bench=. -benchmem ./... | tee bench_baseline.txt
          git checkout -

          go install golang.org/x/perf/cmd/benchstat@latest
          benchstat bench_baseline.txt bench_current.txt > comparison.md

      - name: Comment with results
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const comparison = fs.readFileSync('comparison.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '## Benchmark Comparison\n\n' + comparison
            });
```

### Local Benchmark Validation

```bash
# Script to validate performance before committing
#!/bin/bash
set -e

echo "Running benchmarks..."
go test -bench=. -benchmem -count=3 ./... | tee new_bench.txt

if [ -f baseline_bench.txt ]; then
    echo "Comparing with baseline..."
    benchstat baseline_bench.txt new_bench.txt
fi

# Save as new baseline
cp new_bench.txt baseline_bench.txt
```

## Best Practices

1. **Run Multiple Times** - Use `-count=5` for stable results
2. **Disable Background Jobs** - Close other apps when benchmarking
3. **Same Environment** - Run comparisons on same machine
4. **Real Scenarios** - Benchmark realistic use cases
5. **Document Changes** - Record why you changed implementation
6. **Monitor Regressions** - Alert on significant degradation
7. **Profile Before Optimizing** - Don't guess bottlenecks

## Troubleshooting

### Benchmark Results Vary Widely

```bash
# Use longer benchmark time
go test -bench=. -benchtime=10s ./...

# Run multiple times for stability
go test -bench=. -count=10 ./...

# Check system load
top
```

### Memory Allocations Unexpectedly High

```bash
# Profile memory allocations
go test -bench=. -memprofile=mem.prof ./...
go tool pprof mem.prof

# Look for unnecessary allocations in hot paths
```

### Benchmarks Run Too Fast

```bash
# Some benchmarks may complete too quickly for accurate measurement
# Make the benchmark more complex or run more iterations
go test -bench=. -benchtime=10s ./...
```

## References

- [Go Benchmarking Guide](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [benchstat Tool](https://golang.org/x/perf/cmd/benchstat)
- [Go Performance](https://golang.org/doc/diagnostics)
- [pprof Documentation](https://github.com/google/pprof)
