package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path"
)

type Package struct {
	file   *ast.File // the merge of all Files in this package
	name   string
	scopes []*ast.Scope // contains all top-level package identifiers
}

// Lookup a top-level declaration in this Package
func (p *Package) Lookup(name string) *ast.Object {
	for _, scope := range p.scopes {
		if obj := scope.Lookup(name); obj != nil {
			return obj
		}
	}
	return nil
}

type Project struct {
	name     string
	fset     *token.FileSet
	packages map[string]*Package
}

func NewProject(name string) (p *Project) {
	fset := token.NewFileSet()
	packages := make(map[string]*Package)
	p = &Project{name: name, fset: fset, packages: packages}
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
		var scopes []*ast.Scope
		for _, f := range pkg.Files {
			scopes = append(scopes, f.Scope)
			f.Scope.Outer = nil
		}
		mergedFile := ast.MergePackageFiles(pkg, ast.FilterFuncDuplicates|ast.FilterImportDuplicates)
		p.packages[name] = &Package{name: name, file: mergedFile, scopes: scopes}
	}

	return
}

func (p *Project) get(pkgName string) (pkg *Package) {
	pkg = p.packages[pkgName]
	return
}

var outputConfig = &printer.Config{
	Mode:     printer.TabIndent | printer.UseSpaces,
	Tabwidth: 8,
}

func (p *Project) write(dir string) (err error) {
	srcDir := path.Join(dir, "src")
	os.Mkdir(srcDir, 0777)
	for pkgName, pkg := range p.packages {
		var f *os.File
		if pkgName == "main" {
			pkgName = p.name
		}
		pkgPath := path.Join(srcDir, pkgName)
		os.Mkdir(pkgPath, 0777)
		filePath := path.Join(pkgPath, pkgName+".go")
		f, err = os.Create(filePath)
		if err != nil {
			return
		}
		fmt.Println("Writing", f.Name())
		outputConfig.Fprint(f, p.fset, pkg.file)
	}
	return
}
