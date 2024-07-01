package callgraph

import (
	"github.com/stretchr/testify/require"
	"go/types"
	"testing"
)

func TestNewDatabaseBuilder(t *testing.T) {
	b := NewDatabaseBuilder()
	b.Dir = "testcodedir"
	b.IncludeTestFiles = true
	err := b.Build()
	require.NoError(t, err)
	structs := b.findStructs("testcodedir")
	require.Len(t, structs, 1)
	n, s := b.getStruct(structs[0])
	_ = s
	require.Equal(t, "github.com/ethanvc/gocallgraph/callgraph/testcodedir.Abc", n.String())
	funcs := b.findFuncs("testcodedir")
	require.Len(t, funcs, 7)

	ifs := b.findInterfaces("AbcInterface")
	require.Len(t, ifs, 1)
	_, ifObj := b.getInterface(ifs[0])
	require.NotNil(t, ifObj)
	m, wrongType := types.MissingMethod(n.Underlying(), ifObj, true)
	_ = m
	_ = wrongType
}
