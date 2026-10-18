package genverx

import (
	"os"
	"path/filepath"
	"testing"
)

type generateTestCase struct {
	arch    Arch
	prefix  string
	postfix string
}

func (g generateTestCase) String() string {
	return g.prefix + "_" + g.postfix + ".syso"
}

func newGenerateInfo(t *testing.T) Info {
	t.Helper()

	i, err := NewInfo("1.2.4", AdditionalInfo{
		ProductName: "Test Generate Product",
		Author:      "Test Generate Author",
		Copyright:   "Test Generate Copyright © äüöß",
	})

	if err != nil {
		t.Fatalf("Failed to create info: %v", err)
	}

	return i
}

func newGenerateTestCases(tmpDir string) []generateTestCase {
	return []generateTestCase{
		{ArchI386, filepath.Join(tmpDir, "testver_i386"), "windows_386"},
		{ArchAMD64, filepath.Join(tmpDir, "testver_amd64"), "windows_amd64"},
		{ArchARM, filepath.Join(tmpDir, "testver_arm"), "windows_arm"},
		{ArchARM64, filepath.Join(tmpDir, "testver_arm64"), "windows_arm64"},
	}
}

func TestSucceededGenerate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	i := newGenerateInfo(t)

	for _, c := range newGenerateTestCases(tmpDir) {
		t.Run(c.postfix, func(t *testing.T) {
			err := Generate(i, c.arch, c.prefix)
			if err != nil {
				t.Fatalf("Failed to generate versioned binary: %v", err)
			}

			stat, err := os.Stat(c.String())
			if err != nil {
				t.Fatalf("Failed to stat generated file %q: %v", c.String(), err)
			}

			if !stat.Mode().IsRegular() || stat.Size() == 0 {
				t.Errorf("Generated file %q is not a non-empty regular file", c.String())
			}
		})
	}
}

func TestFailedGenerate(t *testing.T) {
	t.Parallel()

	tmpDir := filepath.Join(t.TempDir(), "abcdedfg")
	i := newGenerateInfo(t)

	for _, c := range newGenerateTestCases(tmpDir) {
		t.Run(c.postfix, func(t *testing.T) {
			err := Generate(i, c.arch, c.prefix)
			if err == nil {
				t.Fatalf("Expected generation to fail, but it succeeded")
			}

			_, err = os.Stat(c.String())
			if !os.IsNotExist(err) {
				t.Fatalf("Expected file %q to not exist, got: %v", c.String(), err)
			}
		})
	}
}

func TestGenerateInvalidArchitecture(t *testing.T) {
	t.Parallel()

	info := newGenerateInfo(t)
	target := filepath.Join(t.TempDir(), "testver")

	err := Generate(info, Arch(255), target)
	if err == nil {
		t.Fatal("Expected unsupported architecture to fail")
	}
}
