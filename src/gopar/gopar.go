package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func main() {
	flag.Parse()
	compilecmd := flag.Arg(0) // run|build|install
	compilepkg := flag.Arg(1)
	if compilepkg == "" {
		panic(fmt.Errorf("Empty compile package"))
	}
	fmt.Println("GoPar Compiler")
	dir, err := ioutil.TempDir(os.TempDir(), "gopar_")
	if err != nil {
		panic(err)
	}
	fmt.Println(dir)
	//defer os.RemoveAll(dir)

	project := NewProject(compilepkg)
	err = project.load(compilepkg)

	if err != nil {
		panic(err)
	}

	mainPkg := project.get("main")
	if mainPkg == nil {
		panic(fmt.Errorf("%s is not a main package", compilepkg))
	}

	showGraph(project)

	compiler := NewCompiler(project)
	compiler.AddPass(NewDefinedTypesPass())
	compiler.AddPass(NewBasicBlockPass())
	compiler.AddPass(NewInvalidConstructPass())
	compiler.AddPass(NewCallGraphPass())
	// analysis starts
	compiler.AddPass(NewAccessPass())
	//compiler.AddPass(NewAccessPassPropogate())
	//compiler.AddPass(NewAccessPassFuncPropogate())
	//compiler.AddPass(NewDependencyPass())
	//compiler.AddPass(NewParallelizePass())
	// modification starts
	//compiler.AddPass(NewInsertBlocksPass())
	//compiler.AddPass(NewWriteKernelsPass())
	// pick parallel loops
	err = compiler.Run()

	//showGraph(project)

	err = project.write(dir)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("go", compilecmd, compilepkg)
	env := os.Environ()
	for i, _ := range env {
		if strings.HasPrefix(env[i], "GOPATH=") {
			env[i] = os.ExpandEnv(fmt.Sprintf("GOPATH=%s:${GOPATH}", dir))
		}
	}
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
