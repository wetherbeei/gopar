package main

import (
	"flag"
	"fmt"
	//"go/build"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	flag.Parse()
	compilepkg := flag.Arg(0)
	fmt.Println("GoPar Compiler")
	// Examine source code packages. Start with the default package or a given one
	// then transform it and examine its includes as well.
	/*
		pkg, err := build.Default.Import(compilepkg, ".", 0)
		if err != nil {
			panic(err)
		}
		for _, gofile := range pkg.GoFiles {
			fmt.Println(gofile)
		}*/
	fset := token.NewFileSet()
	pkgmap, _ := parser.ParseDir(fset, compilepkg, nil, 0)
	for _, pkg := range pkgmap {
		for _, file := range pkg.Files {
			//fmt.Println(file)
			ast.Print(fset, file)
		}
		//funcs := FindParallelFuncs(pkg)

	}
	// Create temporary source directories

	// Launch 'go' with GOPATH edited with new directories
}
