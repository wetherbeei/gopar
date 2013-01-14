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
	//ast.Print(project.fset, mainPkg.file)

	buildCallGraph(project)
	//parallelLoops := pickParallelLoops(callGraph)
	//for loop, _ := range parallelLoops {
	// Copy package if not already copied

	// Generate OpenCL kernel, store in string

	// Modify loop AST to copy data and call the OpenCL kernel
	//}
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
