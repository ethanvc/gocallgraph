package callgraph

import (
	"github.com/stretchr/testify/require"
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
}
