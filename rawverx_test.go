package rawverx

import (
	"fmt"
	"strings"
	"testing"
)

type corePrefixCase struct {
	prefix string
	valid  bool
}

type coreCase struct {
	major uint64
	minor uint64
	patch uint64
}

type preReleaseCase struct {
	preRelease string
	valid      bool
}

type buildCase struct {
	build    Build
	expected string
}

type parseFunc func(string) (SemVer, error)

func expectValidation(t *testing.T, s SemVer, validCore, validPreRelease bool) {
	t.Helper()

	if err := s.Validate(); (err == nil) != (validCore && validPreRelease) {
		t.Errorf("SemVer.Validate() = %v", err)
	}
	if err := s.Core.Validate(); (err == nil) != validCore {
		t.Errorf("Core.Validate() = %v", err)
	}
	if err := s.PreRelease.Validate(); (err == nil) != validPreRelease {
		t.Errorf("PreRelease.Validate() = %v", err)
	}
}

func expectValidSemVer(t *testing.T, parseFn parseFunc, input string, expected SemVer) {
	t.Helper()

	core, err := parseFn(input)
	if err != nil {
		t.Fatalf("expected valid SemVer, got error: %v", err)
	}
	if core != expected {
		t.Fatalf("expected SemVer %q, got %q", expected, core)
	}

	if core.Build.HasCommit() {
		t.Fatalf("expected empty commit hash, got %q", core.Build)
	}
}

func expectInvalidSemVer(t *testing.T, parseFn parseFunc, input string) {
	t.Helper()

	core, err := parseFn(input)
	if err == nil {
		t.Fatalf("expected invalid SemVer, got valid: %q", core)
	}
}

func testPreReleaseCases(t *testing.T, parseFn parseFunc, c string, expected Core, validCore bool) {
	t.Helper()

	preReleaseCases := []preReleaseCase{
		{"", true},
		{"-", false},
		{"--", true},
		{"-0", true},
		{"-1", true},
		{"-01", false},
		{"-001", false},
		{"-01a", true},
		{"-a01", true},
		{"-0a1", true},
		{"-alpha", true},
		{"-.alpha", false},
		{"-alpha.", false},
		{"-alpha.0", true},
		{"-alpha.01", false},
		{"-01.alpha", false},
		{"-alpha.1", true},
		{"-alpha..1", false},
		{"-alpha_1", false},
		{"-ALPHA", true},
		{"-ALPHA.1", true},
		{"-ALPHA_1", false},
		{"-alpha.develop", true},
		{"-alpha.develop.1", true},
		{"-alpha_develop_1", false},
		{"-ALPHA.DEVELOP", true},
		{"-ALPHA.DEVELOP.1", true},
		{"-ALPHA_DEVELOP_1", false},
	}

	for _, prCase := range preReleaseCases {
		p := strings.TrimPrefix(prCase.preRelease, "-")
		i := fmt.Sprintf("%s%s", c, prCase.preRelease)
		t.Run(i, func(t *testing.T) {
			s := SemVer{Core: expected, PreRelease: PreRelease(p)}
			if validCore && prCase.valid {
				expectValidSemVer(t, parseFn, i, s)
			} else {
				expectInvalidSemVer(t, parseFn, i)
			}

			expectValidation(t, s, true, prCase.valid || p == "")
		})
	}
}

func testCorePrefixCases(t *testing.T, parseFn parseFunc, c coreCase, p corePrefixCase) {
	t.Helper()

	expected := Core{c.major, c.minor, c.patch}
	if p.prefix == "" {
		i := fmt.Sprintf("%d.%d.%d", c.major, c.minor, c.patch)
		t.Run(i, func(t *testing.T) {
			testPreReleaseCases(t, parseFn, i, expected, p.valid)
		})
		return
	}

	for mask := 1; mask <= 7; mask++ {
		major := fmt.Sprintf("%d", c.major)
		minor := fmt.Sprintf("%d", c.minor)
		patch := fmt.Sprintf("%d", c.patch)

		e := expected
		if mask&1 != 0 {
			major = p.prefix + major
			e.Major = mustUint64(major)
		}
		if mask&2 != 0 {
			minor = p.prefix + minor
			e.Minor = mustUint64(minor)
		}
		if mask&4 != 0 {
			patch = p.prefix + patch
			e.Patch = mustUint64(patch)
		}

		i := fmt.Sprintf("%s.%s.%s", major, minor, patch)
		t.Run(i, func(t *testing.T) {
			testPreReleaseCases(t, parseFn, i, e, p.valid)
		})
	}
}

func testCoreCases(t *testing.T, parseFn parseFunc, c coreCase) {
	t.Helper()

	cases := []corePrefixCase{
		{"", true},
		{"0", false},
		{"1", true},
		{" ", false},
	}

	for _, p := range cases {
		i := fmt.Sprintf("%s%d.%s%d.%s%d", p.prefix, c.major, p.prefix, c.minor, p.prefix, c.patch)
		t.Run(i, func(t *testing.T) {
			testCorePrefixCases(t, parseFn, c, p)
		})
	}
}

func testParseSemVer(t *testing.T, parseFn parseFunc) {
	t.Helper()

	for v := uint64(0); v <= 7; v++ {
		c := coreCase{major: v & 1, minor: v & 2, patch: v & 4}
		i := fmt.Sprintf("%d.%d.%d", c.major, c.minor, c.patch)
		t.Run(i, func(t *testing.T) {
			testCoreCases(t, parseFn, c)
		})
	}
}

func testParseNegativeSemVer(t *testing.T, parseFn parseFunc) {
	t.Helper()

	cases := []string{
		"-1.0.0",
		"1.-1.0",
		"1.0.-1",
		"-1.-1.-1",
	}

	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			expectInvalidSemVer(t, parseFn, c)
		})
	}
}

func testParseLargeSemVer(t *testing.T, parseFn parseFunc) {
	t.Helper()

	cases := []struct {
		c Core
		v bool
	}{
		{Core{maxCoreValue, 0, 0}, true},
		{Core{0, maxCoreValue, 0}, true},
		{Core{0, 0, maxCoreValue}, true},
		{Core{maxCoreValue + 1, 0, 0}, false},
		{Core{0, maxCoreValue + 1, 0}, false},
		{Core{0, 0, maxCoreValue + 1}, false},
	}

	for _, c := range cases {
		d := fmt.Sprintf("%d.%d.%d", c.c.Major, c.c.Minor, c.c.Patch)
		t.Run(d, func(t *testing.T) {
			if c.v {
				expectValidSemVer(t, parseFn, d, SemVer{Core: c.c})
			} else {
				expectInvalidSemVer(t, parseFn, d)
			}
			expectValidation(t, SemVer{Core: c.c}, c.v, true)
		})
	}
}

func TestParseSemVer(t *testing.T) {
	t.Parallel()

	parseCases := []struct {
		name  string
		parse parseFunc
	}{
		{"ParseSemVerCore", ParseSemVerCore},
		{"NewInfo", func(s string) (SemVer, error) {
			info, err := NewInfo(s)
			info.Version.Build = Build{}

			return info.Version, err
		}},
	}

	for _, c := range parseCases {
		t.Run(c.name, func(t *testing.T) {
			testParseSemVer(t, c.parse)
			testParseNegativeSemVer(t, c.parse)
			testParseLargeSemVer(t, c.parse)
		})
	}
}

func testBuildString(t *testing.T, c buildCase) {
	t.Helper()

	if got := c.build.String(); got != c.expected {
		t.Errorf("Build.String() = %q, want %q", got, c.expected)
	}
}

func testSemVerString(t *testing.T, c buildCase) {
	t.Helper()

	v := SemVer{Build: c.build}
	s := v.String()

	expected := "0.0.0"
	if c.expected != "" {
		expected += "+git." + c.expected
	}

	if s != expected {
		t.Errorf("SemVer.String() = %q, want %q", s, expected)
	}
}

func testInfoString(t *testing.T, c buildCase) {
	t.Helper()

	i := Info{Version: SemVer{Build: c.build}}
	s := i.String()

	expected := "0.0.0 build: "
	if c.expected == "" {
		expected += "unknown"
	} else {
		expected += c.expected
	}

	if s != expected {
		t.Errorf("Info.String() = %q, want %q", s, expected)
	}
}

func TestBuildString(t *testing.T) {
	t.Parallel()

	cases := []buildCase{
		{Build{Commit: "", Dirty: false}, ""},
		{Build{Commit: "", Dirty: true}, ""},
		{Build{Commit: "1234567", Dirty: false}, "1234567"},
		{Build{Commit: "1234567", Dirty: true}, "1234567.dirty"},
		{Build{Commit: "1234567890", Dirty: false}, "1234567890"},
		{Build{Commit: "1234567890", Dirty: true}, "1234567890.dirty"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q/%t", c.build.Commit, c.build.Dirty), func(t *testing.T) {
			testBuildString(t, c)
			testSemVerString(t, c)
			testInfoString(t, c)
		})
	}
}

func TestBuildShortCommitString(t *testing.T) {
	t.Parallel()

	cases := []buildCase{
		{Build{Commit: "1234", Dirty: false}, "1234"},
		{Build{Commit: "1234", Dirty: true}, "1234.dirty"},
		{Build{Commit: "1234567", Dirty: false}, "1234567"},
		{Build{Commit: "1234567", Dirty: true}, "1234567.dirty"},
		{Build{Commit: "1234567890", Dirty: false}, "1234567"},
		{Build{Commit: "1234567890", Dirty: true}, "1234567.dirty"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q/%t", c.build.Commit, c.build.Dirty), func(t *testing.T) {
			testBuildString(t, buildCase{c.build.ShortCommit(), c.expected})
		})
	}
}
