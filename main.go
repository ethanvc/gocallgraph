package main

import (
	"github.com/ethanvc/gocallgraph/callgraph"
)

func main() {
	rootPath := "."
	db := callgraph.NewDatabase()
	db.LoadPackagesInPath(rootPath)
}
