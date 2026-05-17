package build

import "strings"

const (
	RepoURL = "https://github.com/bruh0987/gli-tools"
	RepoGit = "https://github.com/bruh0987/gli-tools.git"
)

var Version = "0.1.0"
var Commit = "unknown"

func DisplayVersion() string {
	version := Version
	if version == "" {
		version = "0.1.0"
	}
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if Commit == "" || Commit == "unknown" {
		return version
	}
	return version + " (" + Commit + ")"
}
