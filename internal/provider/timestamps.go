package provider

import "time"

// normalizeTimestamp parses an ISO 8601 timestamp and re-formats it with
// microsecond precision. This ensures state values are stable across API
// responses that may return varying sub-microsecond precision (e.g. the
// create response returns Go's time.Now() with nanoseconds, while subsequent
// reads from PostgreSQL return microsecond precision after rounding).
func normalizeTimestamp(s string) string {
	if s == "" {
		return s
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return s
	}
	return t.Round(time.Microsecond).UTC().Format(time.RFC3339Nano)
}
