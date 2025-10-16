# Benchmark Quick Reference

Quick reference for common benchmarking commands and tasks.

## Running Benchmarks

### Quick Benchmarks
```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkClientCreation -benchmem ./...

# Run and save results
go test -bench=. -benchmem ./... > results.txt
```

### Stable Results
```bash
# Run 5 times for stability
go test -bench=. -benchmem -count=5 ./...

# Run for longer time
go test -bench=. -benchmem -benchtime=10s ./...

# Both for maximum stability
go test -bench=. -benchmem -count=5 -benchtime=10s ./...
```

### Focused Benchmarks
```bash
# Client benchmarks only
go test -bench=Client -benchmem ./...

# Model benchmarks only
go test -bench=Model -benchmem ./...

# Error benchmarks only
go test -bench=Error -benchmem ./...

# Concurrency benchmarks only
go test -bench=Concurrent -benchmem ./...
```

## Comparing Results

### Before/After Comparison
```bash
# Establish baseline
go test -bench=. -benchmem ./... > baseline.txt

# Make changes
# ... edit code ...

# Get new results
go test -bench=. -benchmem ./... > current.txt

# Compare
benchstat baseline.txt current.txt
```

### Multiple Runs
```bash
# Run 5 times on both versions for stability
go test -bench=. -benchmem -count=5 ./... > baseline.txt
# Make changes
go test -bench=. -benchmem -count=5 ./... > current.txt

# Compare with statistics
benchstat baseline.txt current.txt
```

## Profiling

### CPU Profile
```bash
# Generate CPU profile
go test -bench=. -cpuprofile=cpu.prof ./...

# View in pprof
go tool pprof cpu.prof

# Web view (requires graphviz)
go tool pprof -http=:8080 cpu.prof
```

### Memory Profile
```bash
# Generate memory profile
go test -bench=. -memprofile=mem.prof ./...

# View in pprof
go tool pprof mem.prof
```

### Lock Contention
```bash
# Generate mutex profile
go test -bench=Concurrent -mutexprofile=mutex.prof ./...

# View
go tool pprof mutex.prof
```

## Common pprof Commands

Once in `pprof` interactive shell:

```
top             - Show top 10 functions
top 20          - Show top 20 functions
list main       - Show source for main
web             - Generate graph visualization
png             - Export as PNG
pdf             - Export as PDF
quit            - Exit pprof
```

## Understanding Output

### Benchmark Output
```
BenchmarkClientCreation/WithJWTToken-8  300000  3500 ns/op  400 B/op  10 allocs/op
```

| Part | Meaning |
|------|---------|
| `BenchmarkClientCreation/WithJWTToken` | Test name |
| `-8` | CPU cores |
| `300000` | Iterations |
| `3500 ns/op` | Time per iteration |
| `400 B/op` | Memory per iteration |
| `10 allocs/op` | Allocations per iteration |

### Performance Levels
- **ns** (nanoseconds) = 10^-9 seconds
- **µs** (microseconds) = 10^-6 seconds = 1,000 ns
- **ms** (milliseconds) = 10^-3 seconds = 1,000,000 ns

## Performance Baselines

### Client Operations
| Operation | Time | Allocations |
|-----------|------|-------------|
| Create client | 3-4 µs | 10 |
| Get config | 0.1-0.2 µs | 0 |
| Set timeout | 0.05-0.1 µs | 0 |

### Data Models
| Operation | Time | Allocations |
|-----------|------|-------------|
| Organization Marshal | 1-2 µs | 5 |
| Organization Unmarshal | 2-3 µs | 8 |
| Metrics Marshal (large) | 20-30 µs | 20 |
| Timestamp Parse (RFC) | 0.1 µs | 0 |

### Error Handling
| Operation | Time | Allocations |
|-----------|------|-------------|
| Create error | 0.05 µs | 1 |
| Type switch | 0.01 µs | 0 |
| Unmarshal error | 1-2 µs | 4 |

## Batch Benchmarking

### Setup Baseline
```bash
#!/bin/bash
# benchmark_baseline.sh
echo "Setting up baseline benchmarks..."
mkdir -p benchmark_results

# Run benchmarks multiple times
for i in {1..5}; do
  echo "Run $i..."
  go test -bench=. -benchmem ./... >> benchmark_results/baseline_$i.txt
done

# Combine results
cat benchmark_results/baseline_*.txt > benchmark_results/baseline.txt
echo "Baseline complete"
```

### Compare Against Baseline
```bash
#!/bin/bash
# benchmark_compare.sh
echo "Running comparison benchmarks..."

# Run new benchmarks
go test -bench=. -benchmem ./... > benchmark_results/current.txt

# Install benchstat if needed
go install golang.org/x/perf/cmd/benchstat@latest

# Show comparison
benchstat benchmark_results/baseline.txt benchmark_results/current.txt
```

## Performance Optimization Checklist

Before optimizing:
- [ ] Benchmark confirms problem
- [ ] Profile shows hot spot
- [ ] Measurement is stable
- [ ] Impact is significant (>5%)

When optimizing:
- [ ] Create separate test case
- [ ] Verify correctness
- [ ] Benchmark improvement
- [ ] Document change
- [ ] Get code review

After optimization:
- [ ] Update baseline
- [ ] Add to CI/CD if significant
- [ ] Monitor for regressions
- [ ] Share results

## Continuous Benchmarking

### Local Pre-commit Hook
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Run benchmarks on modified files
if git diff --cached --name-only | grep -q "\.go$"; then
  go test -bench=. -benchmem -short ./...
  if [ $? -ne 0 ]; then
    echo "Benchmarks failed"
    exit 1
  fi
fi
```

### GitHub Actions
```yaml
name: Benchmarks
on: [push, pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run benchmarks
        run: go test -bench=. -benchmem ./... | tee results.txt

      - name: Upload results
        uses: actions/upload-artifact@v4
        with:
          name: benchmark-results
          path: results.txt
```

## Troubleshooting

### Results Vary Widely
```bash
# Use longer benchtime
go test -bench=. -benchtime=10s ./...

# Run multiple times
go test -bench=. -count=10 ./...
```

### Out of Memory During Benchmarking
```bash
# Run fewer iterations or shorter time
go test -bench=. -benchtime=1s -timeout=30s ./...

# Run specific benchmark
go test -bench=BenchmarkSmall -benchmem ./...
```

### Need Detailed Output
```bash
# Verbose benchmarking
go test -bench=. -v -benchmem ./...

# With timing for each run
go test -bench=. -benchtime=1x -benchmem ./...
```

## Best Practices

1. **Run in quiet environment** - Close other apps
2. **Multiple runs** - Use `-count=5` for stability
3. **Compare on same machine** - Different hardware = different results
4. **Document changes** - Note why you modified implementation
5. **Profile first** - Don't optimize without data
6. **Test correctness** - Performance doesn't matter if broken
7. **Small improvements** - Each 5-10% gain matters
8. **Monitor over time** - Track performance across versions

## Common Mistakes

❌ **Don't**
- Benchmark without profiling first
- Compare results from different machines
- Optimize for edge cases
- Sacrifice correctness for speed
- Run once and trust the result

✅ **Do**
- Profile with real data
- Run multiple times for stability
- Benchmark real-world scenarios
- Verify correctness always
- Document and track results

## Resources

- [Go Testing Guide](https://golang.org/pkg/testing/)
- [benchstat](https://golang.org/x/perf/cmd/benchstat)
- [pprof](https://github.com/google/pprof)
- [Full Benchmarking Guide](./docs/BENCHMARKING.md)
