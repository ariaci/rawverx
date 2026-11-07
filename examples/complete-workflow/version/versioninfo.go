package version

import (
	_ "embed"
	"strings"

	"github.com/ariaci/rawverx"
)

//go:generate go -C ../tools/genversioninfo run -buildvcs .

//go:embed VERSION
var applicationVersion string

// NewInfo returns the embedded application version together with
// VCS information available from the Go build.
func NewInfo() (rawverx.Info, error) {
	return rawverx.NewInfo(strings.TrimSpace(applicationVersion))
}
