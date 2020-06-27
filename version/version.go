package version

// VERSION the build version set in the make file using version.txt content
var VERSION string

// GITCOMMIT the build gitcommit set in the make file
var GITCOMMIT string

// GetGitCommit returns the GITCOMMIT if exists
func GetGitCommit() string {
	if GITCOMMIT == "" {
		return "development"
	}

	return GITCOMMIT
}

// GetVersion returns the VERSION if exists
func GetVersion() string {
	if VERSION == "" {
		return "0.0.0"
	}

	return VERSION
}
