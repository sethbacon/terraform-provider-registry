package provider

import (
	"fmt"
	"strconv"
	"strings"
)

type semver struct{ major, minor, patch int }

// parseSemver parses a "vX.Y.Z" or "X.Y.Z" string.
func parseSemver(s string) (semver, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return semver{}, fmt.Errorf("invalid semver %q", s)
	}
	var v semver
	var err error
	if v.major, err = strconv.Atoi(parts[0]); err != nil {
		return semver{}, fmt.Errorf("invalid semver %q: %w", s, err)
	}
	if v.minor, err = strconv.Atoi(parts[1]); err != nil {
		return semver{}, fmt.Errorf("invalid semver %q: %w", s, err)
	}
	if v.patch, err = strconv.Atoi(parts[2]); err != nil {
		return semver{}, fmt.Errorf("invalid semver %q: %w", s, err)
	}
	return v, nil
}

// less reports whether v is strictly less than other.
func (v semver) less(other semver) bool {
	if v.major != other.major {
		return v.major < other.major
	}
	if v.minor != other.minor {
		return v.minor < other.minor
	}
	return v.patch < other.patch
}
