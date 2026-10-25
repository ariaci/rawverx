package rawverx

import (
	"fmt"
	"testing"
)

type comparable[T any] interface {
	fmt.Stringer
	Compare(T) int
}

func testCompare[T comparable[T], Cases ~[]T](t *testing.T, cases Cases) {
	t.Helper()

	for o := 0; o < len(cases); o++ {
		oc := cases[o]
		for i := o; i < len(cases); i++ {
			ic := cases[i]

			if o == i {
				t.Run(fmt.Sprintf("%v == %v", oc, ic), func(t *testing.T) {
					if got := oc.Compare(ic); got != 0 {
						t.Fatalf("expected %q == %q, got %d", oc, oc, got)
					}
				})
			} else {
				t.Run(fmt.Sprintf("%v < %v && %v > %v", oc, ic, ic, oc), func(t *testing.T) {
					if got := oc.Compare(ic); got >= 0 {
						t.Fatalf("expected %q < %q, got %d", oc, ic, got)
					}
					if got := ic.Compare(oc); got <= 0 {
						t.Fatalf("expected %q > %q, got %d", ic, oc, got)
					}
				})
			}
		}
	}
}

func TestCoreCompare(t *testing.T) {
	t.Parallel()

	cores := []Core{
		{0, 0, 0},
		{0, 0, 1},
		{0, 0, 2},
		{0, 1, 0},
		{0, 1, 1},
		{0, 2, 0},
		{1, 0, 0},
		{1, 0, 1},
		{1, 1, 0},
		{1, 1, 1},
		{1, 1, 2},
		{1, 2, 0},
		{1, 2, 1},
		{1, 2, 2},
		{2, 0, 0},
		{2, 0, 1},
		{2, 1, 0},
		{10, 0, 0},
		{100, 0, 0},
		{1000, 0, 0},
	}

	testCompare(t, cores)
}

func TestPreReleaseCompare(t *testing.T) {
	t.Parallel()

	preReleases := []PreRelease{
		"0",
		"1",
		"2",
		"10",
		"01a",
		"0a1",
		"ALPHA",
		"ALPHA.1",
		"ALPHA.DEVELOP",
		"ALPHA.DEVELOP.1",
		"a01",
		"alpha",
		"alpha.0",
		"alpha.1",
		"alpha.1.0",
		"alpha.2",
		"alpha.10",
		"alpha.01a",
		"alpha.1a",
		"alpha.beta",
		"alpha.develop",
		"alpha.develop.1",
		"",
	}

	testCompare(t, preReleases)
}

func TestSemVerCompare(t *testing.T) {
	t.Parallel()

	newSemVer := func(s string) SemVer {
		t.Helper()

		v, err := ParseSemVerCore(s)
		if err != nil {
			t.Fatalf("expected valid SemVer for %q, got error: %v", s, err)
		}

		return v
	}

	versions := []SemVer{
		newSemVer("0.0.0-alpha"),
		newSemVer("0.0.0"),
		newSemVer("0.0.1"),
		newSemVer("1.0.0-alpha.1"),
		newSemVer("1.0.0-alpha.2"),
		newSemVer("1.0.0-beta"),
		newSemVer("1.0.0"),
		newSemVer("1.0.1-alpha"),
		newSemVer("1.0.1"),
		newSemVer("2.0.0-alpha"),
		newSemVer("2.0.0"),
	}

	testCompare(t, versions)
}

func TestSemVerCompareEqualBuild(t *testing.T) {
	t.Parallel()

	a := SemVer{
		Core: Core{1, 2, 3},
	}
	b := SemVer{
		Core:  Core{1, 2, 3},
		Build: Build{Commit: "abcdef", Dirty: true},
	}

	if got := a.Compare(b); got != 0 {
		t.Fatalf("expected equal versions, got %d", got)
	}
	if got := b.Compare(a); got != 0 {
		t.Fatalf("expected equal versions, got %d", got)
	}
}

func TestInfoCompare(t *testing.T) {
	t.Parallel()

	newInfo := func(s, d string) Info {
		t.Helper()

		v, err := ParseSemVerCore(s)
		if err != nil {
			t.Fatalf("expected valid SemVer for %q, got error: %v", s, err)
		}

		return Info{Version: v, Time: d}
	}

	infos := []Info{
		newInfo("0.0.0-alpha", ""),
		newInfo("0.0.0", ""),
		newInfo("1.0.0-alpha", ""),

		newInfo("1.0.0", ""),
		newInfo("1.0.0", "2024-01-01T00:00:00Z"),
		newInfo("1.0.0", "2024-01-01T00:00:01Z"),
		newInfo("1.0.0", "2024-01-01T10:00:00+02:00"), // 08:00 UTC
		newInfo("1.0.0", "2024-01-01T09:00:00Z"),
		newInfo("1.0.0", "2025-01-01T00:00:00Z"),

		newInfo("1.0.1-alpha", ""),
		newInfo("1.0.1", ""),
		newInfo("2.0.0", ""),
	}

	testCompare(t, infos)
}

func TestInfoCompareEqualTime(t *testing.T) {
	t.Parallel()

	a := Info{Version: SemVer{Core: Core{1, 0, 0}}, Time: "2024-01-01T12:00:00Z"}
	b := Info{Version: SemVer{Core: Core{1, 0, 0}}, Time: "2024-01-01T13:00:00+01:00"}

	if got := a.Compare(b); got != 0 {
		t.Fatalf("expected equal timestamps, got %d", got)
	}
	if got := b.Compare(a); got != 0 {
		t.Fatalf("expected equal timestamps, got %d", got)
	}
}
