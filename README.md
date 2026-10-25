# rawverx

A lightweight, idiomatic Go library for application version information.

`rawverx` combines an application-defined semantic version with VCS information embedded by the Go toolchain. It provides structured access to version components, pre-release identifiers, Git commit information, repository state, and the VCS timestamp without requiring build-time generated source files or linker flags.
The design emphasizes simplicity, reproducible builds, and a clean separation between the application version and build metadata.

## ✨ Features

- Semantic version parsing using `major.minor.patch`
- Optional SemVer pre-release identifiers
- Structured version information through dedicated Go types
- Semantic version comparison following SemVer precedence rules
- Validation of version components, build metadata, and timestamps
- Automatic Git commit detection through Go build information
- Automatic detection of modified (`dirty`) source trees
- Access to the VCS timestamp embedded by the Go toolchain
- Full and shortened Git commit representations
- Human-readable version formatting
- No linker flags required
- No generated source files required
- No external dependencies
- Compatible with reproducible Go builds

## 📦 Installation

Add `rawverx` to your Go module:

```bash
go get github.com/ariaci/rawverx
```

Then import it:

```go
import "github.com/ariaci/rawverx"
```

## 🚀 Usage

### Create version information

The application version is supplied explicitly while VCS metadata is detected automatically from the Go build information:

```go
package main

import (
	"fmt"
	"log"

	"github.com/ariaci/rawverx"
)

func main() {
	info, err := rawverx.NewInfo("1.2.3")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(info)
}
```

When built from a Git repository, the resulting information may look similar to:

```text
1.2.3 build: 2704d5d777b4... (2026-09-24T16:11:13Z)
```
If no VCS information is available:

```text
1.2.3 build: unknown
```

### Pre-release versions

Pre-release identifiers can be supplied as part of the application version:

```go
info, err := rawverx.NewInfo("1.2.3-rc.1")
```

The following are examples of valid versions:

```text
0.1.0
1.0.0
1.2.3-alpha
1.2.3-alpha.1
1.2.3-rc.1
1.2.3-01a
```

Numeric pre-release identifiers follow SemVer rules and therefore cannot contain leading zeroes:

```text
1.2.3-0       // valid
1.2.3-1       // valid
1.2.3-01      // invalid
```
An identifier containing non-numeric characters is treated as an alphanumeric identifier:

```text
1.2.3-01a     // valid
```

## 🧩 Version structure

Version information is represented through a small set of dedicated types:

```go
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
```

This keeps the application-defined version separate from metadata originating from the build environment.

## 🔢 Semantic version parsing

Versions can also be parsed directly:

```go
version, err := rawverx.ParseSemVerCore("1.2.3-rc.1")
if err != nil {
	log.Fatal(err)
}

fmt.Println(version.Core.Major)
fmt.Println(version.Core.Minor)
fmt.Println(version.Core.Patch)
fmt.Println(version.PreRelease)
```

`ParseSemVerCore` accepts:

```text
<major>.<minor>.<patch>[-<pre-release>]
```
Build metadata is intentionally not accepted as part of the supplied version string. Build information is represented separately by `rawverx` and populated from the Go toolchain where available.

The numeric `major`, `minor`, and `patch` components are represented as `uint64`.

## ⚖️ Version comparison

`rawverx` provides `Compare` methods for `Core`, `PreRelease`, `SemVer`, and `Info`.

Each comparison returns:

- `-1` if the receiver is less than the other value
- `0` if both values have equal precedence
- `+1` if the receiver is greater than the other value

### Semantic version precedence

`SemVer.Compare` follows the Semantic Versioning precedence rules: core version components are compared numerically, followed by pre-release identifiers.

```go
v1, err := rawverx.ParseSemVerCore("1.2.3-rc.1")
if err != nil {
	log.Fatal(err)
}

v2, err := rawverx.ParseSemVerCore("1.2.3")
if err != nil {
	log.Fatal(err)
}

fmt.Println(v1.Compare(v2)) // -1
fmt.Println(v2.Compare(v1)) // +1
fmt.Println(v1.Compare(v1)) //  0
```

Pre-release identifiers are compared from left to right:

```text
1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-beta < 1.0.0-rc.1 < 1.0.0
```

Numeric identifiers are compared numerically and have lower precedence than non-numeric identifiers. When all shared identifiers are equal, the shorter pre-release has lower precedence.

Git commit information and the dirty state are deliberately ignored by `SemVer.Compare`, as build metadata does not affect SemVer precedence.

### Application information comparison

`Info.Compare` first compares the semantic versions. When they have equal precedence, the VCS timestamps are compared chronologically.

```go
older := rawverx.Info{
	Version: v2,
	Time:    "2024-01-01T00:00:00Z",
}
newer := rawverx.Info{
	Version: v2,
	Time:    "2024-02-01T00:00:00Z",
}

fmt.Println(older.Compare(newer)) // -1
```

An empty timestamp has lower precedence than a non-empty timestamp. Two timestamps representing the same instant compare as equal, even when their timezone offsets differ.

`Compare` expects valid values. Use `Validate` first when comparing values constructed or modified directly.

## ✅ Version validation

All version information types (`Core`, `PreRelease`, `Build`, `SemVer`, and `Info`) provide a `Validate` method.

Validation returns `nil` for a valid value or an error describing the invalid component.

```go
version := rawverx.SemVer{
	Core: rawverx.Core{Major: 1, Minor: 2, Patch: 3},
	PreRelease: "rc.1",
}

if err := version.Validate(); err != nil {
	log.Fatal(err)
}
```

Validation checks the following constraints:

| Type | Validation |
| --- | --- |
| `Core` | Major, minor, and patch must each be within the supported 19-digit range (`0` to `9999999999999999999`) |
| `PreRelease` | Must be empty or follow SemVer pre-release identifier rules |
| `Build` | Commit must be empty or contain only lowercase hexadecimal characters (`0-9`, `a-f`) |
| `SemVer` | Validates its core version, pre-release, and build information |
| `Info` | Validates its semantic version and requires an RFC3339 timestamp when non-empty |

For example, a directly constructed version with an invalid pre-release identifier is rejected:

```go
version.PreRelease = "rc.01"

if err := version.Validate(); err != nil {
	fmt.Println(err) // invalid prerelease "rc.01"
}
```

An empty commit or timestamp is valid because VCS information may be unavailable. The `Dirty` flag does not require an accompanying commit.

`ParseSemVerCore` and `NewInfo` already validate the supplied semantic version string during parsing. `Validate` is especially useful for values assembled or changed directly through the exported fields.

## 🔨 VCS build information

`NewInfo` uses Go's build information to retrieve VCS metadata automatically.

When available, the following settings are used:

| Go build setting | rawverx field |
| --- | --- |
| `vcs.revision` | `Version.Build.Commit` |
| `vcs.modified` | `Version.Build.Dirty` |
| `vcs.time` | `Time` |

No `-ldflags` configuration is required.

### Git commit

The complete Git commit is available through:

```go
info.Version.Build.Commit
```

Check whether a commit is available:

```go
if info.Version.Build.HasCommit() {
	fmt.Println(info.Version.Build.Commit)
}
```

### Short commit

A shortened seven-character representation can be generated with:

```go
build := info.Version.Build.ShortCommit()

fmt.Println(build.Commit)
```

For example:

```text
2704d5d
```

### Dirty builds

When Go detects modifications in the working tree, the build is marked as dirty:

```go
if info.Version.Build.Dirty {
	fmt.Println("modified source tree")
}
```

The string representation appends `.dirty` to the commit:

```text
2704d5d.dirty
```

## 🏷️ Version formatting

### Core version

```go
fmt.Println(info.Version.Core)
```

Example:

```text
1.2.3
```

### Semantic version

`SemVer.String()` combines the application version with available VCS information.

For a clean repository:

```text
1.2.3+git.2704d5d777b4...
```

For a dirty repository:

```text
1.2.3+git.2704d5d777b4....dirty
```

With a pre-release version:

```text
1.2.3-rc.1+git.2704d5d777b4...
```

If no Git commit is available, only the application version is returned:

```text
1.2.3
```

### Application information

`Info.String()` provides a human-readable representation intended for application output:

```go
fmt.Println(info)
```

Example:

```text
1.2.3 build: 2704d5d777b4... (2026-09-24T16:11:13Z)
```

The VCS timestamp is included only when it is available.

## ♻️ Reproducible builds

`rawverx` deliberately does not generate or embed the current build time.

A build timestamp would make otherwise identical builds produce different binaries and therefore break reproducibility.

Instead, `rawverx` uses the VCS metadata already provided by the Go toolchain:

```text
vcs.revision
vcs.modified
vcs.time
```

This keeps version information tied to the source state rather than to the time at which a particular build happened.

## 🪟 Windows version resources

The repository also contains the optional [`genverx`](genverx) module.

`genverx` can be used together with `rawverx` to generate Windows `.syso` resources containing version and application metadata.

It is maintained as a separate Go module so applications that only require `rawverx` do not need to pull in the additional Windows resource generation dependencies.

## 📄 License

`rawverx` is licensed under the [MIT License](LICENSE).
