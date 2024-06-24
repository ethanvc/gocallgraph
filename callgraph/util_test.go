package callgraph

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	"testing"
)

func parseWithSSA(t *testing.T) *ssa.Program {
	cfg := &packages.Config{
		Tests: true,
		Mode: packages.NeedName | packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedExportFile | packages.NeedTypes | packages.NeedSyntax |
			packages.NeedTypesInfo | packages.NeedTypesSizes | packages.NeedModule | packages.NeedEmbedFiles |
			packages.NeedEmbedPatterns,
		Dir: "./testcodedir",
	}
	pkgs, err := packages.Load(cfg, ".")
	require.NoError(t, err)
	prog, _ := ssautil.AllPackages(pkgs, 0)
	prog.Build()
	return prog
}
