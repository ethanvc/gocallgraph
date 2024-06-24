package callgraph

import (
	"github.com/stretchr/testify/require"
	"go/types"
	"strings"
	"testing"
)

func TestInterfaceSet_Basic(t *testing.T) {
	set := NewInterfaceSet()
	prog := parseWithSSA(t)
	for _, pkg := range prog.AllPackages() {
		if !strings.Contains(pkg.String(), "testcodedir") {
			continue
		}
		for _, mem := range pkg.Members {
			name := mem.Name()
			ssaIf, _ := mem.Type().Underlying().(*types.Interface)
			if ssaIf != nil {
				set.Add(pkg, name, ssaIf)
			}
		}
	}
	ifaces := set.FindByName("AbcInterface")
	require.Equal(t, 1, len(ifaces))
}
