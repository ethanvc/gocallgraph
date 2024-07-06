package main

import (
	"errors"
	"fmt"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	"slices"
	"strings"
)

func main() {
	err := realMain()
	if err != nil {
		fmt.Errorf("%s\n", err.Error())
		return
	}
}

func realMain() error {
	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
	}
	initial, err := packages.Load(cfg, "./...")
	if err != nil {
		return err
	}
	if packages.PrintErrors(initial) > 0 {
		return errors.New("errors found")
	}

	// Create and build SSA-form program representation.
	mode := ssa.InstantiateGenerics // instantiate generics by default for soundness
	prog, _ := ssautil.AllPackages(initial, mode)
	prog.Build()

	cg := vta.CallGraph(ssautil.AllFunctions(prog), cha.CallGraph(prog))
	cg.DeleteSyntheticNodes()
	v := newVisitor()
	v.CalleeFunc = "initPos"
	callgraph.GraphVisitEdges(cg, v.Visit)
	return nil
}

type visitor struct {
	CalleeFunc   string
	currentStack []*callgraph.Node
	visitedNode  map[*callgraph.Node]struct{}
	result       [][]*callgraph.Node
}

func newVisitor() *visitor {
	return &visitor{
		visitedNode: make(map[*callgraph.Node]struct{}),
	}
}

func (v *visitor) Visit(edge *callgraph.Edge) error {
	name := edge.Callee.Func.String()
	if strings.Contains(name, v.CalleeFunc) {
		v.findRoot(edge.Callee)
		return errors.New("FoundAndParsed")
	}
	return nil
}

func (v *visitor) findRoot(n *callgraph.Node) {
	if _, ok := v.visitedNode[n]; ok {
		v.result = append(v.result, slices.Clone(v.currentStack))
		return
	}
	v.visitedNode[n] = struct{}{}
	defer delete(v.visitedNode, n)
	v.currentStack = append(v.currentStack, n)
	defer func() {
		v.currentStack = v.currentStack[:len(v.currentStack)-1]
	}()
	if len(n.In) == 0 {
		v.result = append(v.result, slices.Clone(v.currentStack))
		return
	}
	for _, parent := range n.In {
		v.findRoot(parent.Caller)
	}
}
