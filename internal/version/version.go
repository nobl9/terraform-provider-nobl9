package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

const programName = "terraform"

// Set during build time.
var (
	BuildGitRevision string
	BuildGitBranch   string
	BuildVersion     string
)

// GetBuildVersion returns the semantic version without the `v` prefix and other metadata.
func GetBuildVersion() string {
	version := BuildVersion
	if version == "" {
		version = getRuntimeVersion()
	}
	return strings.TrimSpace(version)
}

// GetUserAgent returns a fully qualified User-Agent header value.
func GetUserAgent() string {
	return fmt.Sprintf("%s/%s-%s-%s (%s %s %s)",
		programName, GetBuildVersion(), BuildGitBranch, BuildGitRevision,
		runtime.GOOS, runtime.GOARCH, runtime.Version(),
	)
}

func getRuntimeVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "(devel)" {
		return "0.0.0"
	}
	return strings.TrimPrefix(info.Main.Version, "v")
}
