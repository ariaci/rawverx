package genverx

import (
	"os"
	"strings"
	"testing"
)

type newInfoTestCase struct {
	file     string
	valid    bool
	expected string
}

func expectEquality(t *testing.T, i Info, v string, a AdditionalInfo) {
	t.Helper()

	if i.Version.String() != v {
		t.Errorf("Version = %q; want %q", i.Version.String(), v)
	}
	if i.ProductName != a.ProductName {
		t.Errorf("ProductName = %q; want %q", i.ProductName, a.ProductName)
	}
	if i.Author != a.Author {
		t.Errorf("Author = %q; want %q", i.Author, a.Author)
	}
	if i.Copyright != a.Copyright {
		t.Errorf("Copyright = %q; want %q", i.Copyright, a.Copyright)
	}
}

func checkTestCase(t *testing.T, i Info, err error, c newInfoTestCase, a AdditionalInfo) {
	t.Helper()

	if err != nil && c.valid {
		t.Fatalf("Failed to create new info: %v", err)
	}
	if err == nil && !c.valid {
		t.Fatalf("Expected error but got none")
	}

	if c.valid {
		expectEquality(t, i, c.expected, a)
	}
}

func getVersionFromFile(t *testing.T, file string) string {
	t.Helper()

	v, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Failed to read version file: %v", err)
	}

	return strings.TrimSpace(string(v))
}

func TestNewInfo(t *testing.T) {
	t.Parallel()

	testCases := []newInfoTestCase{
		{"testdata/valid.version", true, "1.2.4-alpha.1"},
		{"testdata/invalid.version", false, ""},
		{"testdata/empty.version", false, ""},
	}

	a := AdditionalInfo{
		ProductName: "Test Product",
		Author:      "Test Author",
		Copyright:   "Test Copyright © äüöß",
	}

	for _, c := range testCases {
		t.Run(c.file, func(t *testing.T) {
			v := getVersionFromFile(t, c.file)
			i, err := NewInfo(v, a)
			checkTestCase(t, i, err, c, a)
		})
	}
}

func TestNewInfoWithVersionFromFile(t *testing.T) {
	t.Parallel()

	testCases := []newInfoTestCase{
		{"testdata/valid.version", true, "1.2.4-alpha.1"},
		{"testdata/invalid.version", false, ""},
		{"testdata/empty.version", false, ""},
		{"testdata/missing.version", false, ""},
	}

	a := AdditionalInfo{
		ProductName: "Test Product",
		Author:      "Test Author",
		Copyright:   "Test Copyright © äüöß",
	}

	for _, c := range testCases {
		t.Run(c.file, func(t *testing.T) {
			i, err := NewInfoWithVersionFromFile(c.file, a)
			checkTestCase(t, i, err, c, a)
		})
	}
}
