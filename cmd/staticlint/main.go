package staticlint

import (
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sortslice"

	"github.com/aykuli/observer/cmd/staticlint/noosexit"
)

func main() {
	multichecker.Main(
		shadow.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		sortslice.Analyzer,
		inspect.Analyzer,
		atomicalign.Analyzer,
		noosexit.Analyzer,
	)
}
