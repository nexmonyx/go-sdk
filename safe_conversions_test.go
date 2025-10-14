package nexmonyx

import (
	"math"
	"testing"
	"time"
)

func TestSafeInt64ToUint64(t *testing.T) {
	tests := []struct {
		name    string
		input   int64
		want    uint64
		wantErr bool
	}{
		{
			name:    "positive value",
			input:   1234567890,
			want:    1234567890,
			wantErr: false,
		},
		{
			name:    "zero value",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "max int64",
			input:   math.MaxInt64,
			want:    uint64(math.MaxInt64),
			wantErr: false,
		},
		{
			name:    "negative value",
			input:   -1,
			want:    0,
			wantErr: true,
		},
		{
			name:    "large negative value",
			input:   math.MinInt64,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeInt64ToUint64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeInt64ToUint64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeInt64ToUint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeInt64ToUint64OrZero(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  uint64
	}{
		{
			name:  "positive value",
			input: 9876543210,
			want:  9876543210,
		},
		{
			name:  "zero value",
			input: 0,
			want:  0,
		},
		{
			name:  "max int64",
			input: math.MaxInt64,
			want:  uint64(math.MaxInt64),
		},
		{
			name:  "negative value returns zero",
			input: -100,
			want:  0,
		},
		{
			name:  "large negative value returns zero",
			input: math.MinInt64,
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeInt64ToUint64OrZero(tt.input)
			if got != tt.want {
				t.Errorf("SafeInt64ToUint64OrZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeUint64ToInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   uint64
		want    int64
		wantErr bool
	}{
		{
			name:    "small positive value",
			input:   1234567890,
			want:    1234567890,
			wantErr: false,
		},
		{
			name:    "zero value",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "max int64",
			input:   uint64(math.MaxInt64),
			want:    math.MaxInt64,
			wantErr: false,
		},
		{
			name:    "max int64 + 1 (overflow)",
			input:   uint64(math.MaxInt64) + 1,
			want:    0,
			wantErr: true,
		},
		{
			name:    "max uint64 (overflow)",
			input:   math.MaxUint64,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeUint64ToInt64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeUint64ToInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeUint64ToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeUint64ToDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   uint64
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "1 second",
			input:   1_000_000_000,
			want:    1 * time.Second,
			wantErr: false,
		},
		{
			name:    "zero nanoseconds",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "1 nanosecond",
			input:   1,
			want:    1 * time.Nanosecond,
			wantErr: false,
		},
		{
			name:    "max duration (~292 years)",
			input:   uint64(math.MaxInt64),
			want:    time.Duration(math.MaxInt64),
			wantErr: false,
		},
		{
			name:    "overflow duration",
			input:   uint64(math.MaxInt64) + 1,
			want:    0,
			wantErr: true,
		},
		{
			name:    "large overflow",
			input:   math.MaxUint64,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeUint64ToDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeUint64ToDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeUint64ToDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeUint64ToDurationCapped(t *testing.T) {
	tests := []struct {
		name  string
		input uint64
		want  time.Duration
	}{
		{
			name:  "1 second",
			input: 1_000_000_000,
			want:  1 * time.Second,
		},
		{
			name:  "zero nanoseconds",
			input: 0,
			want:  0,
		},
		{
			name:  "max duration",
			input: uint64(math.MaxInt64),
			want:  time.Duration(math.MaxInt64),
		},
		{
			name:  "overflow capped to max",
			input: uint64(math.MaxInt64) + 1,
			want:  time.Duration(math.MaxInt64),
		},
		{
			name:  "large overflow capped to max",
			input: math.MaxUint64,
			want:  time.Duration(math.MaxInt64),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeUint64ToDurationCapped(tt.input)
			if got != tt.want {
				t.Errorf("SafeUint64ToDurationCapped() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark tests to ensure minimal performance overhead
func BenchmarkSafeInt64ToUint64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SafeInt64ToUint64(1234567890)
	}
}

func BenchmarkSafeInt64ToUint64OrZero(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SafeInt64ToUint64OrZero(1234567890)
	}
}

func BenchmarkSafeUint64ToDuration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SafeUint64ToDuration(1_000_000_000)
	}
}

func BenchmarkSafeUint64ToDurationCapped(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SafeUint64ToDurationCapped(1_000_000_000)
	}
}

// Test real-world scenarios
func TestRealWorldDiskMetricsConversion(t *testing.T) {
	tests := []struct {
		name           string
		totalBytes     int64
		usedBytes      int64
		freeBytes      int64
		expectSkip     bool
		expectNonZero  bool
	}{
		{
			name:          "valid large disk",
			totalBytes:    1000000000000, // 1TB
			usedBytes:     500000000000,  // 500GB
			freeBytes:     500000000000,  // 500GB
			expectSkip:    false,
			expectNonZero: true,
		},
		{
			name:          "negative total (invalid data)",
			totalBytes:    -1000,
			usedBytes:     500,
			freeBytes:     500,
			expectSkip:    true,
			expectNonZero: false,
		},
		{
			name:          "zero total disk",
			totalBytes:    0,
			usedBytes:     0,
			freeBytes:     0,
			expectSkip:    true,
			expectNonZero: false,
		},
		{
			name:          "max int64 disk size",
			totalBytes:    math.MaxInt64,
			usedBytes:     math.MaxInt64 / 2,
			freeBytes:     math.MaxInt64 / 2,
			expectSkip:    false,
			expectNonZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalConverted := SafeInt64ToUint64OrZero(tt.totalBytes)
			usedConverted := SafeInt64ToUint64OrZero(tt.usedBytes)
			freeConverted := SafeInt64ToUint64OrZero(tt.freeBytes)

			// Should skip if total is zero (invalid/negative became zero)
			shouldSkip := totalConverted == 0
			if shouldSkip != tt.expectSkip {
				t.Errorf("shouldSkip = %v, want %v", shouldSkip, tt.expectSkip)
			}

			// Check non-zero expectation
			if !shouldSkip && (totalConverted == 0) != !tt.expectNonZero {
				t.Errorf("totalConverted = %v, expectNonZero = %v", totalConverted, tt.expectNonZero)
			}

			// Verify no wraparound occurred for valid positive values
			if tt.totalBytes > 0 && totalConverted == 0 {
				t.Errorf("positive value converted to zero unexpectedly")
			}
			if tt.usedBytes > 0 && usedConverted == 0 {
				t.Errorf("positive used value converted to zero unexpectedly")
			}
			if tt.freeBytes > 0 && freeConverted == 0 {
				t.Errorf("positive free value converted to zero unexpectedly")
			}
		})
	}
}

func TestRealWorldCPUTimeConversion(t *testing.T) {
	tests := []struct {
		name               string
		cpuUsageNsec       uint64
		expectCapped       bool
		expectApproxYears  float64
	}{
		{
			name:              "1 hour of CPU time",
			cpuUsageNsec:      3_600_000_000_000, // 1 hour
			expectCapped:      false,
			expectApproxYears: 0.0001,
		},
		{
			name:              "100 years of CPU time",
			cpuUsageNsec:      3_155_760_000_000_000_000, // ~100 years
			expectCapped:      false,
			expectApproxYears: 100,
		},
		{
			name:              "max valid duration",
			cpuUsageNsec:      uint64(math.MaxInt64),
			expectCapped:      false,
			expectApproxYears: 292, // ~292 years
		},
		{
			name:              "overflow value",
			cpuUsageNsec:      math.MaxUint64,
			expectCapped:      true,
			expectApproxYears: 292, // Capped to ~292 years
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := SafeUint64ToDurationCapped(tt.cpuUsageNsec)

			if tt.expectCapped {
				if duration != time.Duration(math.MaxInt64) {
					t.Errorf("expected capped duration, got %v", duration)
				}
			} else {
				if duration == time.Duration(math.MaxInt64) && tt.cpuUsageNsec < uint64(math.MaxInt64) {
					t.Errorf("unexpectedly capped valid duration")
				}
			}

			// Verify duration is always positive
			if duration < 0 {
				t.Errorf("negative duration: %v", duration)
			}

			// Verify duration is reasonable
			years := duration.Hours() / 24 / 365
			if math.Abs(years-tt.expectApproxYears) > 10 && !tt.expectCapped {
				t.Errorf("duration %v = ~%.1f years, want ~%.1f years",
					duration, years, tt.expectApproxYears)
			}
		})
	}
}
