package rawverx

import (
	_ "embed"
	"fmt"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type Core struct {
	Major uint64
	Minor uint64
	Patch uint64
}
type PreRelease string
type Build struct {
	Commit string
	Dirty  bool
}

type SemVer struct {
	Core       Core
	PreRelease PreRelease
	Build      Build
}

type Info struct {
	Version SemVer
	Time    string
}

var semverCoreRe = regexp.MustCompile(
	`^(0|[1-9][0-9]{0,18})\.(0|[1-9][0-9]{0,18})\.(0|[1-9][0-9]{0,18})(?:-((?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?$`,
)
var preReleasePartNumRe = regexp.MustCompile(`^(?:0|[1-9][0-9]*)$`)
var commitHashRe = regexp.MustCompile(`^[0-9a-f]*$`)

func NewInfo(c string) (Info, error) {
	v, err := ParseSemVerCore(c)
	if err != nil {
		return Info{}, err
	}

	i := Info{
		Version: v,
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return i, nil
	}

	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			i.Version.Build.Commit = s.Value
		case "vcs.modified":
			i.Version.Build.Dirty = s.Value == "true"
		case "vcs.time":
			i.Time = s.Value
		}
	}

	return i, nil
}

func (c Core) Compare(o Core) int {
	cmp := func(a uint64, b uint64) int {
		if a < b {
			return -1
		}

		return 1
	}

	if c.Major != o.Major {
		return cmp(c.Major, o.Major)
	}
	if c.Minor != o.Minor {
		return cmp(c.Minor, o.Minor)
	}
	if c.Patch != o.Patch {
		return cmp(c.Patch, o.Patch)
	}

	return 0
}

func (p PreRelease) Compare(o PreRelease) int {
	switch {
	case p == o:
		return 0
	// 1.0.0 > 1.0.0-alpha
	case p == "":
		return 1
	// 1.0.0-alpha < 1.0.0
	case o == "":
		return -1
	}

	pp := strings.Split(p.String(), ".")
	po := strings.Split(o.String(), ".")

	spp := len(pp)
	spo := len(po)

	s := spp
	if spo < spp {
		s = spo
	}

	for i := 0; i < s; i++ {
		ppi := pp[i]
		poi := po[i]

		nppi := preReleasePartNumRe.MatchString(ppi)
		npoi := preReleasePartNumRe.MatchString(poi)

		switch {
		case nppi && npoi:
			lppi := len(ppi)
			lpoi := len(poi)
			if lppi < lpoi {
				return -1
			}
			if lppi > lpoi {
				return 1
			}
		case nppi && !npoi:
			return -1
		case !nppi && npoi:
			return 1
		}

		if r := strings.Compare(ppi, poi); r != 0 {
			return r
		}
	}

	switch {
	case spp < spo:
		return -1
	case spo < spp:
		return 1
	default:
		return 0
	}
}

func (v SemVer) Compare(o SemVer) int {
	if r := v.Core.Compare(o.Core); r != 0 {
		return r
	}

	return v.PreRelease.Compare(o.PreRelease)
}

func (i Info) Compare(o Info) int {
	if r := i.Version.Compare(o.Version); r != 0 {
		return r
	}

	if i.Time == "" || o.Time == "" {
		return strings.Compare(i.Time, o.Time)
	}

	ti, _ := time.Parse(time.RFC3339, i.Time)
	to, _ := time.Parse(time.RFC3339, o.Time)

	switch {
	case ti.Before(to):
		return -1
	case ti.After(to):
		return 1
	default:
		return 0
	}
}

func (p PreRelease) Validate() error {
	if p == "" {
		return nil
	}

	if semverCoreRe.MatchString("0.0.0-" + p.String()) {
		return nil
	}

	return fmt.Errorf("invalid prerelease %q", p)
}

func (b Build) Validate() error {
	if commitHashRe.MatchString(b.Commit) {
		return nil
	}

	return fmt.Errorf("invalid commit hash %q", b.Commit)
}

func (v SemVer) Validate() error {
	if err := v.PreRelease.Validate(); err != nil {
		return err
	}

	if err := v.Build.Validate(); err != nil {
		return err
	}

	return nil
}

func (i Info) Validate() error {
	if err := i.Version.Validate(); err != nil {
		return err
	}

	if i.Time == "" {
		return nil
	}

	if _, err := time.Parse(time.RFC3339, i.Time); err != nil {
		return fmt.Errorf("invalid timestamp %q: %w", i.Time, err)
	}

	return nil
}

func (c Core) String() string {
	return fmt.Sprintf("%d.%d.%d", c.Major, c.Minor, c.Patch)
}

func (p PreRelease) HasValue() bool {
	return len(p) > 0
}

func (p PreRelease) String() string {
	return string(p)
}

func (b Build) ShortCommit() Build {
	c := b.Commit
	if len(c) > 7 {
		c = c[:7]
	}

	return Build{
		Commit: c,
		Dirty:  b.Dirty,
	}
}

func (b Build) HasCommit() bool {
	return b.Commit != ""
}

func (b Build) String() string {
	if b.Commit == "" {
		return ""
	}

	if b.Dirty {
		return b.Commit + ".dirty"
	}
	return b.Commit
}

func (v SemVer) String() string {
	s := fmt.Sprint(v.Core)
	if v.PreRelease.HasValue() {
		s += fmt.Sprint("-", v.PreRelease)
	}
	if v.Build.HasCommit() {
		s += fmt.Sprint("+git.", v.Build)
	}

	return s
}

func (i Info) String() string {
	v := i.Version
	v.Build = Build{}

	s := fmt.Sprint(v, " build: ")

	c := "unknown"
	if i.Version.Build.HasCommit() {
		c = fmt.Sprint(i.Version.Build)
	}
	s += c

	if i.Time != "" {
		s += fmt.Sprint(" (", i.Time, ")")
	}

	return s
}

func (b Info) HasTime() bool {
	return b.Time != ""
}

func mustUint64(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint64(v)
}

func ParseSemVerCore(c string) (SemVer, error) {
	m := semverCoreRe.FindStringSubmatch(c)
	if m == nil {
		return SemVer{}, fmt.Errorf("invalid SemVer: %s", c)
	}

	return SemVer{
		Core: Core{
			Major: mustUint64(m[1]),
			Minor: mustUint64(m[2]),
			Patch: mustUint64(m[3]),
		},
		PreRelease: PreRelease(m[4]),
	}, nil
}
