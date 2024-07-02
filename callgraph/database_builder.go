package callgraph

import (
	"go/types"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	"strings"
)

type DatabaseBuilder struct {
	Dir              string
	IncludeTestFiles bool
	interfaces       map[*types.Named]*types.Interface
	structs          map[*types.Named]*types.Struct
	funcs            map[*types.Func]struct{}
	graph            *callgraph.Graph
}

func NewDatabaseBuilder() *DatabaseBuilder {
	return &DatabaseBuilder{
		interfaces: make(map[*types.Named]*types.Interface),
		structs:    make(map[*types.Named]*types.Struct),
		funcs:      make(map[*types.Func]struct{}),
	}
}

func (b *DatabaseBuilder) Build() error {
	prog, err := b.buildProgram()
	if err != nil {
		return err
	}
	b.parseAllInterfaces(prog)
	allFuncs := ssautil.AllFunctions(prog)
	b.graph = vta.CallGraph(allFuncs, nil)
	return nil
}

func (b *DatabaseBuilder) buildProgram() (*ssa.Program, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedExportFile | packages.NeedTypes | packages.NeedSyntax |
			packages.NeedTypesInfo | packages.NeedTypesSizes | packages.NeedModule | packages.NeedEmbedFiles |
			packages.NeedEmbedPatterns,
		Dir: b.Dir,
	}
	cfg.Tests = b.IncludeTestFiles
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, err
	}
	prog, ssaPkgs := ssautil.AllPackages(pkgs, 0)
	_ = ssaPkgs
	prog.Build()
	return prog, nil
}

func (b *DatabaseBuilder) parseAllInterfaces(prog *ssa.Program) {
	for _, pkg := range prog.AllPackages() {
		for _, mem := range pkg.Members {
			if b.parseMemberAsInterface(mem) {
				continue
			}
			if b.parseMemberAsFunc(mem) {
				continue
			}
			b.parseMemberAsStruct(mem)
		}
	}
}

func (b *DatabaseBuilder) parseMemberAsInterface(mem ssa.Member) bool {
	namedType, _ := mem.Type().(*types.Named)
	if namedType == nil {
		return false
	}
	ssaIf, _ := namedType.Underlying().(*types.Interface)
	if ssaIf == nil {
		return false
	}
	b.interfaces[namedType] = ssaIf
	return true
}

func (b *DatabaseBuilder) parseMemberAsStruct(mem ssa.Member) bool {
	namedType, _ := mem.Type().(*types.Named)
	if namedType == nil {
		return false
	}
	ssaStruct, _ := namedType.Underlying().(*types.Struct)
	if ssaStruct == nil {
		return false
	}
	b.structs[namedType] = ssaStruct
	for i := 0; i < namedType.NumMethods(); i++ {
		b.funcs[namedType.Method(i)] = struct{}{}
	}
	return true
}

func (b *DatabaseBuilder) parseMemberAsFunc(mem ssa.Member) bool {
	if !strings.Contains(mem.Package().String(), "testcodedir") {
		return false
	}
	ssaFunc, _ := mem.(*ssa.Function)
	if ssaFunc == nil {
		return false
	}
	typesFunc, _ := ssaFunc.Object().(*types.Func)
	if typesFunc == nil {
		return false
	}
	b.funcs[typesFunc] = struct{}{}
	return true
}

func (b *DatabaseBuilder) findInterfaces(name string) []string {
	var result []string
	for k, _ := range b.interfaces {
		if strings.Contains(k.String(), name) {
			result = append(result, k.String())
		}
	}
	return result
}

func (b *DatabaseBuilder) getInterface(name string) (*types.Named, *types.Interface) {
	for k, v := range b.interfaces {
		if k.String() == name {
			return k, v
		}
	}
	return nil, nil
}

func (b *DatabaseBuilder) findStructs(structName string) []string {
	var result []string
	for k, _ := range b.structs {
		if strings.Contains(k.String(), structName) {
			result = append(result, k.String())
		}
	}
	return result
}

func (b *DatabaseBuilder) getStruct(name string) (*types.Named, *types.Struct) {
	for k, v := range b.structs {
		if k.String() == name {
			return k, v
		}
	}
	return nil, nil
}

func (b *DatabaseBuilder) findFuncs(name string) []string {
	var result []string
	for k, _ := range b.funcs {
		if strings.Contains(k.String(), name) {
			result = append(result, k.String())
		}
	}
	return result
}
