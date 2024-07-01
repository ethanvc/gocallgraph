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
	structs := b.findStructs("AbcValueRecv")
	require.Len(t, structs, 1)
	n, s := b.getStruct(structs[0])
	_ = s
	require.Equal(t, "github.com/ethanvc/gocallgraph/callgraph/testcodedir.AbcValueRecv", n.String())
	funcs := b.findFuncs("testcodedir")
	require.Len(t, funcs, 8)

	ifs := b.findInterfaces("AbcInterface")
	require.Len(t, ifs, 1)
	_, ifObj := b.getInterface(ifs[0])
	require.NotNil(t, ifObj)
	m, wrongType := types.MissingMethod(n, ifObj, false)
	require.Nil(t, m)
	require.False(t, wrongType)
	structs = b.findStructs("AbcPointerRecv")
	require.Len(t, structs, 1)
	n, s = b.getStruct(structs[0])
	m, wrongType = types.MissingMethod(types.NewPointer(n), ifObj, false)
	require.Nil(t, m)
}
