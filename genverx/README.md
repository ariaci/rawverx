# genverx

A small Go library for generating Windows version resources from semantic application versions.

`genverx` builds on [`rawverx`](https://github.com/ariaci/rawverx) and generates architecture-specific Windows `.syso` files containing version and application metadata.

The generated resources are picked up automatically by the Go toolchain during normal Windows builds. No external resource compiler, linker flags, or additional build tooling is required.

## ✨ Features

- Generates native Windows version resources for Go applications
- Uses semantic version information provided by `rawverx`
- Supports version information supplied directly or from a file
- Embeds product name, author, copyright, version, and Git build information
- Generates architecture-specific `.syso` files
- Supports `386`, `amd64`, `arm`, and `arm64`
- Designed for use with `go generate`
- No external tools required
- Normal application builds remain a plain `go build`

## 📦 Installation

Add `genverx` to the Go module used for resource generation:

```bash
go get github.com/ariaci/rawverx/genverx
```

Then import it:

```go
import "github.com/ariaci/rawverx/genverx"
```

`genverx` is maintained as a separate Go module from `rawverx` so applications that only need runtime version information do not need to pull in the Windows resource generation dependencies.

The current module requires Go 1.25 or newer.

## 🚀 Usage

### Create version information

Version information can be created directly from a semantic version:

```go
package main

import (
	"log"

	"github.com/ariaci/rawverx/genverx"
)

func main() {
	info, err := genverx.NewInfo("1.2.3", genverx.AdditionalInfo{
		ProductName: "example",
		Author:      "Example Author",
		Copyright:   "© Example Author. Licensed under MIT.",
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := genverx.Generate(info, genverx.ArchAMD64, "winres"); err != nil {
		log.Fatal(err)
	}
}
```

This generates:

```text
winres_windows_amd64.syso
```

The file can reside directly in the Go package containing the application entry point or other package that should provide the Windows version resource.

When building for Windows/AMD64, the Go toolchain automatically includes the matching resource file.

## 📄 Version from file

Applications often keep their version in a separate file such as:

```text
1.2.3
```

For this case, `genverx` provides `NewInfoWithVersionFromFile`:

```go
info, err := genverx.NewInfoWithVersionFromFile("VERSION", genverx.AdditionalInfo{
	ProductName: "example",
	Author:      "Example Author",
	Copyright:   "© Example Author. Licensed under MIT.",
})
if err != nil {
	log.Fatal(err)
}
```

Leading and trailing whitespace in the version file is ignored.

The version itself follows the semantic version format accepted by `rawverx`:

```text
<major>.<minor>.<patch>[-<pre-release>]
```

For example:

```text
1.0.0
1.2.3-alpha
1.2.3-alpha.1
1.2.3-rc.1
```

VCS information such as the Git commit and dirty state is obtained through `rawverx` from the Go build information.

## 🏗️ Generate resources for multiple architectures

A single generator can create resources for all required target architectures:

```go
package main

import (
	"log"

	"github.com/ariaci/rawverx/genverx"
)

func main() {
	info, err := genverx.NewInfoWithVersionFromFile("../../version/VERSION", genverx.AdditionalInfo{
		ProductName: "example",
		Author:      "Example Author",
		Copyright:   "© Example Author. Licensed under MIT.",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, arch := range []genverx.Arch{
		genverx.ArchI386,
		genverx.ArchAMD64,
		genverx.ArchARM64,
	} {
		if err := genverx.Generate(info, arch, "../../version/winres"); err != nil {
			log.Fatal(err)
		}
	}
}
```

This produces:

```text
winres_windows_386.syso
winres_windows_amd64.syso
winres_windows_arm64.syso
```

The supported architectures are:

| genverx architecture | Generated suffix |
| --- | --- |
| `ArchI386` | `_windows_386.syso` |
| `ArchAMD64` | `_windows_amd64.syso` |
| `ArchARM` | `_windows_arm.syso` |
| `ArchARM64` | `_windows_arm64.syso` |

The target passed to `Generate` is a filename prefix. The platform-specific suffix is appended automatically.

## ⚙️ Using with `go generate`

A typical project can keep the generator separate from the application package:

```text
.
├── tools/
│   └── genversioninfo/
│       └── main.go
└── version/
    ├── VERSION
    └── versioninfo.go
```

`version/versioninfo.go` can contain only the generation directive:

```go
//go:generate go run -C ../tools/genversioninfo -buildvcs .

package version
```

The generator reads the application version and creates the required resources:

```go
package main

import (
	"log"

	"github.com/ariaci/rawverx/genverx"
)

func main() {
	info, err := genverx.NewInfoWithVersionFromFile("../../version/VERSION", genverx.AdditionalInfo{
		ProductName: "example",
		Author:      "Example Author",
		Copyright:   "© Example Author. Licensed under MIT.",
	})
	if err != nil {
		log.Fatalf("error creating version info: %v", err)
	}

	for _, arch := range []genverx.Arch{
		genverx.ArchI386,
		genverx.ArchAMD64,
		genverx.ArchARM64,
	} {
		if err := genverx.Generate(info, arch, "../../version/winres"); err != nil {
			log.Fatalf("error generating version info for architecture %v: %v", arch, err)
		}
	}
}
```

Resources can then be regenerated with:

```bash
go generate ./...
```

After generation, building the application requires nothing special:

```bash
go build
```

No standalone resource compiler or additional build tool needs to be installed.

## 🪟 Windows version information

The generated Windows resource contains information derived from both the application metadata and the `rawverx` version.

The following fields are populated:

| Windows field | Source |
| --- | --- |
| `CompanyName` | `Author` |
| `FileDescription` | Product name and version |
| `FileVersion` | Application version |
| `InternalName` | `ProductName` |
| `LegalCopyright` | `Copyright` |
| `OriginalFilename` | `<ProductName>.exe` |
| `PrivateBuild` | Git build information |
| `ProductName` | `ProductName` |
| `ProductVersion` | Application version |

The numeric Windows file and product versions use:

```text
major.minor.patch.0
```

while the textual version fields can additionally contain pre-release and VCS information supplied by `rawverx`.

## 🔗 Relationship to rawverx

`genverx` deliberately focuses only on resource generation.

[`rawverx`](https://github.com/ariaci/rawverx) provides the application version model and runtime VCS information, while `genverx` converts that information into native Windows resources.

Keeping both modules separate allows applications that only need version information to depend solely on the lightweight `rawverx` module without pulling in resource-generation dependencies.

## 📄 License

`genverx` is licensed under the [MIT License](../LICENSE).

<p align="center">
  <a href="https://hub.xerj.org/contributors/ariaci">
    <img src="https://hub.xerj.org/contributors/ariaci/badge.svg"
         alt="XERJ Corpus Hub contributor">
  </a>
</p>
