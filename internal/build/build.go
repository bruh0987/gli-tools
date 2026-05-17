package build

import "strings"

const (
	RepoURL = "https://github.com/bruh0987/gli-tools"
	RepoGit = "https://github.com/bruh0987/gli-tools.git"
)

var Version = "dev"
var Commit = "unknown"

func DisplayVersion() string {
	version := Version
	if version == "" {
		version = "dev"
	}
	if version == "dev" {
		if Commit == "" || Commit == "unknown" {
			return version
		}
		return version + " (" + Commit + ")"
	}
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if Commit == "" || Commit == "unknown" {
		return version
	}
	return version + " (" + Commit + ")"
}
