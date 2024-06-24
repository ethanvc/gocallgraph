package callgraph

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	"golang.org/x/tools/refactor/satisfy"
)

func Test_SSA(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test_code_test.go", nil, parser.ParseComments)
	require.NoError(t, err)
	files := []*ast.File{f}

	// Create the type-checker's package.
	pkg := types.NewPackage("hello", "")

	// Type-check the package, load dependencies.
	// Create and build the SSA program.
	ssaPkg, typeInfo, err := ssautil.BuildPackage(
		&types.Config{Importer: importer.Default()}, fset, pkg, files, ssa.SanityCheckFunctions)
	require.NoError(t, err)
	require.Equal(t, `package hello`, ssaPkg.String())
	{
		// get string const value by variable name
		nc := ssaPkg.Const("globalStr")
		require.Equal(t, "globalStr", nc.Name())
		require.Equal(t, "hello.globalStr", nc.String())
		require.Equal(t, "const", nc.Token().String())
		typ := nc.Type()
		require.Equal(t, "untyped string", typ.String())
		// Underlying may return itself.
		require.Equal(t, "untyped string", typ.Underlying().String())

		obj := nc.Object()
		require.Equal(t, "globalStr", obj.Name())
		require.Equal(t, "const hello.globalStr untyped string", obj.String())
		require.Equal(t, "hello.globalStr", obj.Id())
	}
	{
		fc := ssaPkg.Func("GlobalFunc")
		require.Equal(t, "GlobalFunc", fc.Name())
		require.Equal(t, "hello.GlobalFunc", fc.String())
	}
	{
		abcStruct := ssaPkg.Type("Abc")
		require.Equal(t, "Abc", abcStruct.Name())
		require.Equal(t, "hello.Abc", abcStruct.String())
	}
	{
		allFunctions := ssautil.AllFunctions(ssaPkg.Prog)
		require.NotNil(t, allFunctions)
	}
	{
		finder := satisfy.Finder{}
		finder.Find(typeInfo, files)
		require.NotZero(t, len(finder.Result))
	}
}

func TestLoadPackages(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes,
		Dir:  "",
	}
	startT := time.Now()
	pkgs, err := packages.Load(cfg, "./...")
	require.NoError(t, err)
	d := time.Now().Sub(startT).Seconds()
	require.Equal(t, 0, d)
	_ = pkgs
}

func Test_StaticCallGraph(t *testing.T) {
	prog := parseWithSSA(t)
	allFuncs := ssautil.AllFunctions(prog)
	f := getFunc(allFuncs, "CallInterfaceFunc")
	require.NotNil(t, f)
	var callInstr ssa.CallInstruction
Leave:
	for _, b := range f.Blocks {
		for _, instr := range b.Instrs {
			switch site := instr.(type) {
			case ssa.CallInstruction:
				callInstr = site
				break Leave
			}
		}
	}
	require.NotNil(t, callInstr)
	require.Equal(t, "invoke abc.MemberFunc()", callInstr.String())
	ssaCall := callInstr.(*ssa.Call)
	require.Equal(t, "t0", ssaCall.Name())
	require.Equal(t, "invoke abc.MemberFunc()", ssaCall.String())
	require.Equal(t, "invoke abc.MemberFunc()", ssaCall.Common().String())
	method := ssaCall.Call.Method
	require.Equal(t, "func (github.com/ethanvc/gocallgraph.AbcInterface).MemberFunc()",
		method.String())
	require.Equal(t, "MemberFunc",
		method.Name())
	require.Equal(t, "(github.com/ethanvc/gocallgraph.AbcInterface).MemberFunc",
		method.FullName())
	require.Equal(t, "main", method.Pkg().Name())
	require.Equal(t, "github.com/ethanvc/gocallgraph", method.Pkg().Path())
	sig := method.Type().(*types.Signature)
	require.Equal(t, "func()", sig.String())
	require.Equal(t, "var  github.com/ethanvc/gocallgraph.AbcInterface", sig.Recv().String())
	require.Equal(t, "", sig.Recv().Name())
	require.Equal(t, "github.com/ethanvc/gocallgraph.", sig.Recv().Id())
	require.Equal(t, "github.com/ethanvc/gocallgraph.AbcInterface", sig.Recv().Type().String())

	f = getFunc(allFuncs, "CallGlobalFunc")
	require.NotNil(t, f)
	callInstr = nil
Leave1:
	for _, b := range f.Blocks {
		for _, instr := range b.Instrs {
			switch site := instr.(type) {
			case ssa.CallInstruction:
				callInstr = site
				break Leave1
			}
		}
	}
	require.NotNil(t, callInstr)
	ssaVal := callInstr.Common().Value
	require.Equal(t, "github.com/ethanvc/gocallgraph.GlobalFunc", ssaVal.String())
	require.Equal(t, "GlobalFunc", ssaVal.Name())
}

func getFunc(funcs map[*ssa.Function]bool, name string) *ssa.Function {
	for f, _ := range funcs {
		if strings.Contains(f.String(), name) {
			return f
		}
	}
	return nil
}

func Test_ParseAllStruct(t *testing.T) {
	prog := parseWithSSA(t)
	PkgMembers(t, prog)
	AllParsedFiles(prog.Fset, "callgraph")
	require.NotNil(t, prog)
	ssaPkg := getSSAPackage(prog, "gocallgraph.test")
	require.NotNil(t, ssaPkg)
}

func getSSAPackage(prog *ssa.Program, pkgName string) *ssa.Package {
	pkgs := prog.AllPackages()
	var result []*ssa.Package
	for _, ssaPkg := range pkgs {
		if strings.Contains(ssaPkg.String(), pkgName) {
			result = append(result, ssaPkg)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result[0]
}

func AllParsedFiles(fset *token.FileSet, part string) {
	var result []string
	fset.Iterate(func(f *token.File) bool {
		result = append(result, f.Name())
		return true
	})
	var result2 []string
	for _, s := range result {
		if strings.Contains(s, part) {
			result2 = append(result2, s)
		}
	}
	return
}

func PkgMembers(t *testing.T, prog *ssa.Program) {
	var intfTyp types.Type
	var intImplTyp types.Type
	for _, pkg := range prog.AllPackages() {
		for _, mem := range pkg.Members {
			pos := prog.Fset.Position(mem.Pos())
			if !strings.Contains(pos.Filename, "test_code_test.go") {
				continue
			}
			name := mem.Name()
			if name == "AbcInterface" {
				intfTyp = mem.Type()
			}
			if name == "Abc" {
				intImplTyp = mem.Type()
			}
		}
	}
	require.True(t, types.IsInterface(intfTyp))
	_ = intImplTyp
}

// https://github.com/kisielk/godepgraph/blob/master/main.go
