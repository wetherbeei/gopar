package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
)

type Package struct {
	file  *ast.File // the merge of all Files in this package
	name  string
	scope *ast.Scope // contains all top-level package identifiers
}

type Project struct {
	fset     *token.FileSet
	packages map[string]*Package
}

func NewProject() (p *Project) {
	fset := token.NewFileSet()
	packages := make(map[string]*Package)
	p = &Project{fset: fset, packages: packages}
	return
}

// Load a package into the Project
func (p *Project) load(pkgName string) (err error) {
	if _, ok := p.packages[pkgName]; ok {
		return
	}
	buildPkg, err := build.Default.Import(pkgName, ".", 0)
	if err != nil {
		return
	}

	pkgs, err := parser.ParseDir(p.fset, buildPkg.Dir, nil, 0)
	if err != nil {
		return
	}

	for name, pkg := range pkgs {
		mergedFile := ast.MergePackageFiles(pkg, ast.FilterFuncDuplicates|ast.FilterImportDuplicates)
		p.packages[name] = &Package{name: name, file: mergedFile, scope: pkg.Scope}
	}

	return
}

func (p *Project) get(pkgName string) (pkg *Package) {
	pkg = p.packages[pkgName]
	return
}
