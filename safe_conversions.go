package nexmonyx

import (
	"fmt"
	"math"
	"time"
)

// SafeInt64ToUint64 safely converts int64 to uint64 with overflow checking.
// Returns an error if the value is negative, as negative values cannot be
// represented as unsigned integers without wraparound.
//
// This is commonly used for size/byte conversions where negative values
// indicate invalid data (e.g., negative disk space, memory usage).
//
// Example:
//
//	val, err := SafeInt64ToUint64(diskMetrics.TotalBytes)
//	if err != nil {
//	    log.Warn("Invalid disk size", "error", err)
//	    return err
//	}
func SafeInt64ToUint64(val int64) (uint64, error) {
	if val < 0 {
		return 0, fmt.Errorf("cannot convert negative int64 to uint64: %d", val)
	}
	return uint64(val), nil
}

// SafeInt64ToUint64OrZero safely converts int64 to uint64, returning 0 for negative values.
// This is useful when you want to treat negative values as zero (e.g., in aggregations)
// rather than failing with an error.
//
// Example:
//
//	total := SafeInt64ToUint64OrZero(diskMetrics.TotalBytes)
func SafeInt64ToUint64OrZero(val int64) uint64 {
	if val < 0 {
		return 0
	}
	return uint64(val)
}

// SafeUint64ToInt64 safely converts uint64 to int64 with overflow checking.
// Returns an error if the value exceeds math.MaxInt64, which would cause
// wraparound to negative values.
//
// This is commonly used when converting large unsigned values to signed types
// like time.Duration (which is int64 internally).
//
// Example:
//
//	duration, err := SafeUint64ToInt64(cpuNanoseconds)
//	if err != nil {
//	    log.Warn("CPU time overflow", "error", err)
//	    return maxDuration
//	}
//	return time.Duration(duration) * time.Nanosecond
func SafeUint64ToInt64(val uint64) (int64, error) {
	if val > math.MaxInt64 {
		return 0, fmt.Errorf("uint64 value exceeds int64 max: %d > %d", val, int64(math.MaxInt64))
	}
	return int64(val), nil
}

// SafeUint64ToDuration safely converts uint64 nanoseconds to time.Duration.
// Returns an error if the value would overflow int64 (> ~292 years).
//
// time.Duration is internally int64, so values larger than math.MaxInt64
// nanoseconds cannot be represented accurately.
//
// Example:
//
//	duration, err := SafeUint64ToDuration(cpuUsageNSec)
//	if err != nil {
//	    log.Warn("Duration overflow", "error", err)
//	    return time.Duration(math.MaxInt64)
//	}
func SafeUint64ToDuration(nanoseconds uint64) (time.Duration, error) {
	if nanoseconds > math.MaxInt64 {
		return 0, fmt.Errorf("nanoseconds value exceeds time.Duration capacity: %d > %d (~292 years)",
			nanoseconds, int64(math.MaxInt64))
	}
	return time.Duration(nanoseconds) * time.Nanosecond, nil
}

// SafeUint64ToDurationCapped safely converts uint64 nanoseconds to time.Duration,
// capping at the maximum representable duration instead of returning an error.
//
// This is useful when you prefer graceful degradation over error handling.
//
// Example:
//
//	duration := SafeUint64ToDurationCapped(cpuUsageNSec)  // Never fails
func SafeUint64ToDurationCapped(nanoseconds uint64) time.Duration {
	if nanoseconds > math.MaxInt64 {
		return time.Duration(math.MaxInt64)
	}
	return time.Duration(nanoseconds) * time.Nanosecond
}
