// Package ldflags provide access to ld flags
package ldflags

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildInfo(t *testing.T) {
	info := Info{BuildVersion: "v1.2", BuildDate: "2024-09-01 20:30:13", BuildCommit: "789456d"}
	expected := `Build version: v1.2
Build date   : 2024-09-01 20:30:13
Build commit : 789456d
`
	t.Run("BuildInfo", func(t *testing.T) {
		require.Equal(t, fmt.Sprint(BuildInfo(info)), expected)
	})
}
