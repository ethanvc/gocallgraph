package callgraph

import (
	"go/types"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type DatabaseBuilder struct {
	Dir              string
	IncludeTestFiles bool
	interfaces       map[string]*types.Interface
}

func NewDatabaseBuilder() *DatabaseBuilder {
	return &DatabaseBuilder{
		interfaces: make(map[string]*types.Interface),
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
			namedType := mem.Type()
			ssaIf, _ := namedType.Underlying().(*types.Interface)
			if ssaIf != nil {
				b.interfaces[namedType.String()] = ssaIf
			}
		}
	}
}
