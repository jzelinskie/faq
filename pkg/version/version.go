package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Version is faq's version string
var Version string

// UsageVersion introspects the process debug data for Go modules to return a
// version string.
func UsageVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("failed to read BuildInfo because the program was compiled with Go " + runtime.Version())
	}

	if Version == "" {
		// The version wasn't set by ldflags, so fallback to the Go module version.
		// Although, this value is pretty much guaranteed to just be "devel".
		Version = bi.Main.Version
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s version %s\n", bi.Path, Version)
	for _, dep := range bi.Deps {
		fmt.Fprintf(&b, "\t%s %s\n", dep.Path, dep.Version)
	}
	return b.String()
}
