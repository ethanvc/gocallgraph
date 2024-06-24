package callgraph

import (
	"go/types"
	"golang.org/x/tools/go/ssa"
	"strings"
)

type InterfaceSet struct {
	s map[InterfaceKey]InterfaceInfo
}

func NewInterfaceSet() *InterfaceSet {
	return &InterfaceSet{
		s: make(map[InterfaceKey]InterfaceInfo),
	}
}

func (s *InterfaceSet) Add(pkg *ssa.Package, name string, v *types.Interface) {
	s.s[InterfaceKey{pkg, name}] = InterfaceInfo{pkg, name, v}
}

func (s *InterfaceSet) FindByName(name string) []InterfaceInfo {
	var result []InterfaceInfo
	for k, v := range s.s {
		if strings.Contains(k.Name, name) {
			result = append(result, v)
		}
	}
	return result
}

type InterfaceKey struct {
	Pkg  *ssa.Package
	Name string
}

type InterfaceInfo struct {
	Pkg   *ssa.Package
	Name  string
	Iface *types.Interface
}
