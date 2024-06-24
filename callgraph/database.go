package callgraph

import (
	"bytes"
	"fmt"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type Database struct {
	funcs map[string]*FuncInfo
}

func NewDatabase() *Database {
	return &Database{
		funcs: make(map[string]*FuncInfo),
	}
}

func (db *Database) LoadPackagesInPath(p string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports |
			packages.NeedDeps | packages.NeedExportFile | packages.NeedTypes | packages.NeedSyntax |
			packages.NeedTypesInfo | packages.NeedTypesSizes | packages.NeedModule | packages.NeedEmbedFiles |
			packages.NeedEmbedPatterns,
		Dir: p,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return err
	}
	prog, _ := ssautil.AllPackages(pkgs, 0)
	prog.Build()
	allFuncs := ssautil.AllFunctions(prog)
	for f, _ := range allFuncs {
		for _, block := range f.Blocks {
			for _, instr := range block.Instrs {
				callInstr, ok := instr.(ssa.CallInstruction)
				if !ok {
					continue
				}
				caller := NewFuncInfoBySSAFunction(prog.Fset, f)
				db.parseCallInstruction(caller, callInstr)
			}
		}
	}
	return nil
}

func (db *Database) parseCallInstruction(caller *FuncInfo, instr ssa.CallInstruction) {

}

type FuncInfo struct {
	PkgName      string
	ReceiverName string
	MethodName   string
	pos          string
}

func NewFuncInfoBySSAFunction(fset *token.FileSet, f *ssa.Function) *FuncInfo {
	info := &FuncInfo{}
	info.MethodName = f.Name()
	if f.Pkg != nil && f.Pkg.Pkg != nil {
		info.PkgName = f.Pkg.Pkg.Path()
	}
	info.initPos(fset, f.Pos())
	return info
}

func NewFuncInfoBySSAMethod(fset *token.FileSet, method *types.Func) (*FuncInfo, error) {
	info := &FuncInfo{}
	info.MethodName = method.Name()
	info.initPos(fset, method.Pos())
	return info, nil
}

func (info *FuncInfo) initPos(fset *token.FileSet, pos token.Pos) {
	plainPos := fset.Position(pos)
	info.pos = fmt.Sprintf("%s:%d", plainPos.Filename, plainPos.Line)
}

func (info *FuncInfo) GetFullName() string {
	var buf bytes.Buffer
	buf.WriteString(info.PkgName + ".")
	if info.ReceiverName != "" {
		buf.WriteString("(" + info.ReceiverName + ")")
	}
	buf.WriteString(info.MethodName)
	return buf.String()
}
