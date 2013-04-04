package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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

	var importGraph func(string) ([]string, error)
	importGraph = func(pkgPath string) (result []string, err error) {
		if exists := project.get(pkgPath); exists != nil {
			return
		}
		err = project.load(pkgPath)
		if err != nil {
			return
		}
		fmt.Println(pkgPath)
		var packageImports []string
		pkg := project.get(pkgPath)
		if pkg == nil {
			return
		}
		fmt.Println("->", pkg.Imports())
		for _, i := range pkg.Imports() {
			packageImports, err = importGraph(i)
			if err != nil {
				return
			}

			result = append(result, packageImports...)
		}
		result = append(result, pkgPath)
		return
	}
	// Generate an import dependency graph of all used packages
	var importOrder []string
	importOrder, err = importGraph(compilepkg)
	if err != nil {
		panic(err)
	}

	fmt.Println(importOrder)

	mainPkg := project.get(compilepkg)
	if mainPkg.name != "main" {
		panic(fmt.Errorf("%s is not a main package", compilepkg))
	}

	AddAnalysisPasses := func(compiler *Compiler) {
		compiler.AddPass(NewDefinedTypesPass())
		compiler.AddPass(NewBasicBlockPass())
		compiler.AddPass(NewInvalidConstructPass())
		compiler.AddPass(NewCallGraphPass())
		// analysis starts
		compiler.AddPass(NewAccessPass())
		compiler.AddPass(NewAccessPassPropogate())
		compiler.AddPass(NewAccessPassFuncPropogate())
		compiler.AddPass(NewDependencyPass())
	}

	AddParallelizePasses := func(compiler *Compiler) {
		compiler.AddPass(NewParallelizePass())
		// modification starts
		compiler.AddPass(NewInsertBlocksPass())
		compiler.AddPass(NewWriteKernelsPass())
	}

	// analyze all of the used packages
	var analyzed map[string]bool = make(map[string]bool)
	compiler := NewCompiler(project)
	AddAnalysisPasses(compiler)

	for _, pkgPath := range importOrder {
		if analyzed[pkgPath] {
			continue
		}
		//showGraph(project, pkgPath)

		pkg := project.get(pkgPath)
		err = compiler.Run(pkg)
		if err != nil {
			panic(err)
		}
		analyzed[pkgPath] = true
	}

	// modify the main package
	AddParallelizePasses(compiler)
	if err = compiler.Run(mainPkg); err != nil {
		panic(err)
	}

	if err = project.write(dir, mainPkg); err != nil {
		panic(err)
	}

	goparPath, _ := exec.LookPath(os.Args[0])
	goparRoot, _ := filepath.Abs(path.Clean(path.Dir(goparPath) + "/../"))
	cmd := exec.Command("go", compilecmd, compilepkg)
	env := os.Environ()
	for i, _ := range env {
		if strings.HasPrefix(env[i], "GOPATH=") {
			env[i] = os.ExpandEnv(fmt.Sprintf("GOPATH=%s:%s:${GOPATH}", dir, goparRoot))
			fmt.Println(env[i])
		}
	}
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
