package provider

import "testing"

func TestParseSemver(t *testing.T) {
	cases := []struct {
		input   string
		wantErr bool
		major   int
		minor   int
		patch   int
	}{
		{"1.0.0", false, 1, 0, 0},
		{"v1.2.3", false, 1, 2, 3},
		{"0.9.99", false, 0, 9, 99},
		{"bad", true, 0, 0, 0},
		{"1.2", true, 0, 0, 0},
	}
	for _, tc := range cases {
		got, err := parseSemver(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseSemver(%q): expected error, got none", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseSemver(%q): unexpected error: %v", tc.input, err)
			continue
		}
		if got.major != tc.major || got.minor != tc.minor || got.patch != tc.patch {
			t.Errorf("parseSemver(%q) = %+v, want {%d %d %d}", tc.input, got, tc.major, tc.minor, tc.patch)
		}
	}
}

func TestSemverLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"0.9.0", "1.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.1", "1.0.0", false},
		{"1.0.0", "1.0.1", true},
		{"2.0.0", "1.9.9", false},
	}
	for _, tc := range cases {
		a, _ := parseSemver(tc.a)
		b, _ := parseSemver(tc.b)
		if got := a.less(b); got != tc.want {
			t.Errorf("%s < %s: got %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
