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
	v := &visitor{
		calleeFunc: "WriteString",
		rootFunc:   "gocallgraph",
	}
	callgraph.GraphVisitEdges(cg, v.Visit)
	return nil
}

type visitor struct {
	calleeFunc string
	rootFunc   string
}

func (v *visitor) Visit(edge *callgraph.Edge) error {
	name := edge.Callee.Func.String()
	if strings.Contains(name, v.calleeFunc) {
		v.findRoot(edge.Callee)
	}
	return nil
}

func (v *visitor) findRoot(callee *callgraph.Node) {
	processedNode := make(map[*callgraph.Node]struct{})
	nodeTasks := []*callgraph.Node{callee}
	var result []*callgraph.Node
	for len(nodeTasks) > 0 {
		n := nodeTasks[len(nodeTasks)-1]
		nodeTasks = nodeTasks[:len(nodeTasks)-1]
		processedNode[n] = struct{}{}
		for _, edge := range n.In {
			parent := edge.Caller
			if len(parent.In) == 0 {
				if strings.Contains(parent.Func.String(), v.rootFunc) {
					result = append(result, parent)
				}
				continue
			}
			for _, p := range parent.In {
				if _, ok := processedNode[p.Caller]; !ok {
					nodeTasks = append(nodeTasks, p.Caller)
				}
			}
		}
	}
	fmt.Println(result)
}
