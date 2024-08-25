// Package ldflags provide access to ld flags
package ldflags

import "fmt"

type BuildInfo struct {
	BuildVersion string
	BuildDate    string
	BuildCommit  string
}

// Print prints linter flag on server build
// @example go build -buildvcs=false -ldflags "-X main.buildVersion=v1.2 -X 'main.buildDate=$(date +'%Y-%m-%d %H:%M:%S')' -X main.buildCommit=c6c208b" -o server
func Print(bi BuildInfo) {
	res := ""
	for k, v := range map[string]string{
		"Build version": bi.BuildVersion,
		"Build date   ": bi.BuildDate,
		"Build commit ": bi.BuildCommit} {
		if v == "" {
			v = `N\A`
		}
		res += k + ": " + v + "\n"
	}

	fmt.Println(res)
}
