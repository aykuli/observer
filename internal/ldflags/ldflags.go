// Package ldflags provide access to ld flags
package ldflags

type Info struct {
	BuildVersion string
	BuildDate    string
	BuildCommit  string
}

// BuildInfo prints linter flag on server build
// @example go build -buildvcs=false -ldflags "-X main.buildVersion=v1.2 -X 'main.buildDate=$(date +'%Y-%m-%d %H:%M:%S')' -X main.buildCommit=c6c208b" -o server
func BuildInfo(bi Info) string {
	res := ""
	if bi.BuildVersion == "" {
		bi.BuildVersion = `N\A`
	}
	res += "Build version: " + bi.BuildVersion + "\n"

	if bi.BuildDate == "" {
		bi.BuildDate = `N\A`
	}
	res += "Build date   : " + bi.BuildDate + "\n"

	if bi.BuildCommit == "" {
		bi.BuildCommit = `N\A`
	}
	res += "Build commit : " + bi.BuildCommit + "\n"

	return res
}
