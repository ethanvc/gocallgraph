package callgraph

import (
	"go/types"
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
}

func NewDatabaseBuilder() *DatabaseBuilder {
	return &DatabaseBuilder{
		interfaces: make(map[*types.Named]*types.Interface),
		structs:    make(map[*types.Named]*types.Struct),
	}
}

func (b *DatabaseBuilder) Build() error {
	prog, err := b.buildProgram()
	if err != nil {
		return err
	}
	b.parseAllInterfaces(prog)
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
	return true
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
