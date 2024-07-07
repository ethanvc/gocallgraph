package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func main() {
	err := realMain()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
}

func realMain() error {
	conf := &Config{}
	err := parseCommandLine(conf)
	if err != nil {
		return err
	}
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
	visitConf := &newVisitorConfig{
		CalleeFunc:    conf.FocusFunc,
		RemoveTopRoot: conf.RemoveTopRoot,
	}
	v, err := newVisitor(visitConf)
	if err != nil {
		return err
	}
	callgraph.GraphVisitEdges(cg, v.Visit)
	fmt.Printf("print stack root:\n")
	for _, stack := range v.result {
		fmt.Printf("%v\n", stack[len(stack)-1])
	}
	return nil
}

type visitor struct {
	conf             *newVisitorConfig
	removeTopRootReg *regexp.Regexp
	currentStack     []*callgraph.Node
	visitedNode      map[*callgraph.Node]struct{}
	result           [][]*callgraph.Node
}

type newVisitorConfig struct {
	CalleeFunc    string
	RemoveTopRoot string
}

func newVisitor(conf *newVisitorConfig) (*visitor, error) {
	removeTopRootReg, err := regexp.Compile(conf.RemoveTopRoot)
	if err != nil {
		return nil, err
	}
	return &visitor{
		conf:             conf,
		removeTopRootReg: removeTopRootReg,
		visitedNode:      make(map[*callgraph.Node]struct{}),
	}, nil
}

func (v *visitor) Visit(edge *callgraph.Edge) error {
	name := edge.Callee.Func.String()
	if strings.Contains(name, v.conf.CalleeFunc) {
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
	if v.removeTopRootReg.MatchString(n.Func.String()) {
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

type Config struct {
	FocusFunc     string
	RemoveTopRoot string
}

func (conf *Config) validate() error {
	if conf.FocusFunc == "" {
		return errors.New("FocusFunc is required")
	}
	return nil
}

func parseCommandLine(conf *Config) error {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	focusFunc := flagSet.String("focus_func", "", "")
	removeTopRoot := flagSet.String("remove_top_root", "", "")
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return err
	}
	conf.FocusFunc = *focusFunc
	conf.RemoveTopRoot = *removeTopRoot
	err = conf.validate()
	if err != nil {
		return err
	}
	return nil
}
