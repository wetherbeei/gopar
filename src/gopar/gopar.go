package main

import (
	"flag"
	"fmt"
	"go/ast"
	"io/ioutil"
	"os"
)

func main() {
	flag.Parse()
	compilepkg := flag.Arg(0)
	fmt.Println("GoPar Compiler")
	dir, err := ioutil.TempDir(os.TempDir(), "gopar_")
	if err != nil {
		panic(err)
	}
	fmt.Println(dir)
	defer os.RemoveAll(dir)

	project := NewProject()
	err = project.load(compilepkg)

	if err != nil {
		panic(err)
	}

	fmt.Println(project)
	mainPkg := project.get("main")
	if mainPkg == nil {
		panic(fmt.Errorf("%s is not a main package", compilepkg))
	}
	ast.Print(project.fset, mainPkg.file)

	//callGraph := buildCallGraph()
	//parallelLoops := pickParallelLoops(callGraph)
	//for loop, _ := range parallelLoops {
	// Copy package if not already copied

	// Generate OpenCL kernel, store in string

	// Modify loop AST to copy data and call the OpenCL kernel
	//}
}
