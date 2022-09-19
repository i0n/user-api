package version

// Build information. Populated at build-time.
var (
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
	GoVersion string
)

// Map provides the iterable version information.
var Map = map[string]string{
	"version":   Version,
	"revision":  Revision,
	"branch":    Branch,
	"buildUser": BuildUser,
	"buildDate": BuildDate,
	"goVersion": GoVersion,
}

// GetVersion gets the current version string
func GetVersion() string {
	v := Map["version"]
	return v
}

// GetRevision gets the current revision string
func GetRevision() string {
	v := Map["revision"]
	return v
}

// GetBranch gets the current branch string
func GetBranch() string {
	v := Map["branch"]
	return v
}

// GetBuildUser gets the current build user string
func GetBuildUser() string {
	v := Map["buildUser"]
	return v
}

// GetBuildDate gets the current build date string
func GetBuildDate() string {
	v := Map["buildDate"]
	return v
}

// GetGoVersion gets the current go version string
func GetGoVersion() string {
	v := Map["goVersion"]
	return v
}
