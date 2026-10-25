package rawverx

import (
	"fmt"
	"testing"
)

func TestBuildValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		build Build
		valid bool
	}{
		{Build{}, true},
		{Build{Dirty: true}, true},
		{Build{Commit: "1234567"}, true},
		{Build{Commit: "abcdef0", Dirty: true}, true},
		{Build{Commit: "0123456789abcdef"}, true},
		{Build{Commit: "ABCDEF0"}, false},
		{Build{Commit: "123xyz0"}, false},
		{Build{Commit: "123 567"}, false},
		{Build{Commit: "1234567.dirty", Dirty: true}, false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q/%t", c.build.Commit, c.build.Dirty), func(t *testing.T) {
			if err := c.build.Validate(); (err == nil) != c.valid {
				t.Errorf("Build.Validate() = %v, expected valid = %t", err, c.valid)
			}
		})
	}
}

func TestInfoValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		info  Info
		valid bool
	}{
		{"empty", Info{}, true},
		{"valid version", Info{Version: SemVer{Core: Core{1, 2, 3}}}, true},
		{"valid timestamp", Info{Time: "2024-01-01T12:00:00Z"}, true},
		{"valid timezone", Info{Time: "2024-01-01T13:00:00+01:00"}, true},
		{"invalid timestamp", Info{Time: "2024-01-01"}, false},
		{"invalid timestamp format", Info{Time: "2024-01-01 12:00:00"}, false},
		{"invalid timestamp value", Info{Time: "2024-13-01T12:00:00Z"}, false},
		{"invalid core", Info{Version: SemVer{Core: Core{Major: maxCoreValue + 1}}}, false},
		{"invalid prerelease", Info{Version: SemVer{PreRelease: "alpha.01"}}, false},
		{"invalid build", Info{Version: SemVer{Build: Build{Commit: "XYZ"}}}, false},
		{"valid fractional seconds", Info{Time: "2024-01-01T12:00:00.123456789Z"}, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.info.HasTime(); got != (c.info.Time != "") {
				t.Errorf("Info.HasTime() = %t, expected %t", got, c.info.Time != "")
			}
			if err := c.info.Validate(); (err == nil) != c.valid {
				t.Errorf("Info.Validate() = %v, expected valid = %t", err, c.valid)
			}
		})
	}
}
