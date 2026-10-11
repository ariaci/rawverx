package genverx

import (
	"fmt"
	"os"
	"strings"

	rawverx "github.com/ariaci/rawverx"

	winres "github.com/tc-hib/winres"
	winres_version "github.com/tc-hib/winres/version"
)

type Arch uint8

const (
	ArchI386 Arch = iota
	ArchAMD64
	ArchARM
	ArchARM64
)

type AdditionalInfo struct {
	ProductName string
	Author      string
	Copyright   string
}

type Info struct {
	AdditionalInfo
	Version rawverx.SemVer
}

type archDefinition struct {
	Arch    winres.Arch
	Postfix string
}

var archDefinitions = map[Arch]archDefinition{
	ArchI386:  {Arch: winres.ArchI386, Postfix: "_windows_386"},
	ArchAMD64: {Arch: winres.ArchAMD64, Postfix: "_windows_amd64"},
	ArchARM:   {Arch: winres.ArchARM, Postfix: "_windows_arm"},
	ArchARM64: {Arch: winres.ArchARM64, Postfix: "_windows_arm64"},
}

func NewInfo(core string, additional AdditionalInfo) (Info, error) {
	ri, err := rawverx.NewInfo(core)
	if err != nil {
		return Info{}, err
	}

	return Info{AdditionalInfo: additional, Version: ri.Version}, nil
}

func NewInfoWithVersionFromFile(filename string, additional AdditionalInfo) (Info, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Info{}, fmt.Errorf("read version file %q: %w", filename, err)
	}

	return NewInfo(strings.TrimSpace(string(data)), additional)
}

func createWinResVersionInfo(i Info) winres_version.Info {
	b := i.Version.Build
	i.Version.Build = i.Version.Build.ShortCommit()
	v := i.Version.String()

	vi := winres_version.Info{}

	vi.SetFileVersion(fmt.Sprintf("%d.%d.%d.0", i.Version.Core.Major, i.Version.Core.Minor, i.Version.Core.Patch))
	vi.SetProductVersion(fmt.Sprintf("%d.%d.%d.0", i.Version.Core.Major, i.Version.Core.Minor, i.Version.Core.Patch))

	vi.Set(winres.LCIDDefault, "CompanyName", i.Author)
	vi.Set(winres.LCIDDefault, "FileDescription", fmt.Sprintf("%s %s", i.ProductName, v))
	vi.Set(winres.LCIDDefault, "FileVersion", v)
	vi.Set(winres.LCIDDefault, "InternalName", i.ProductName)
	vi.Set(winres.LCIDDefault, "LegalCopyright", i.Copyright)
	vi.Set(winres.LCIDDefault, "OriginalFilename", fmt.Sprintf("%s.exe", i.ProductName))
	vi.Set(winres.LCIDDefault, "PrivateBuild", b.String())
	vi.Set(winres.LCIDDefault, "ProductName", i.ProductName)
	vi.Set(winres.LCIDDefault, "ProductVersion", v)

	return vi
}

func writeWinResResources(rs winres.ResourceSet, a Arch, t string) (err error) {
	d, ok := archDefinitions[a]
	if !ok {
		return fmt.Errorf("unsupported architecture: %v", a)
	}

	out, err := os.Create(fmt.Sprintf("%s%s.syso", t, d.Postfix))
	if err != nil {
		return fmt.Errorf("unable to create output file: %w", err)
	}

	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("unable to close output file: %w", closeErr)
		}
	}()

	if err = rs.WriteObject(out, d.Arch); err != nil {
		return fmt.Errorf("unable to write output file: %w", err)
	}

	return nil
}

func Generate(info Info, arch Arch, targetPrefix string) error {
	rs := winres.ResourceSet{}
	rs.SetVersionInfo(createWinResVersionInfo(info))

	return writeWinResResources(rs, arch, targetPrefix)
}
