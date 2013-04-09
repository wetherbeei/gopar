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
	"strings"
)

type Package struct {
	project *Project
	file    *ast.File // the merge of all Files in this package
	name    string
	scopes  []*ast.Scope // contains all top-level package identifiers
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

func (p *Package) TopLevel() map[string]*ast.Object {
	m := make(map[string]*ast.Object)
	for _, scope := range p.scopes {
		for name, obj := range scope.Objects {
			m[name] = obj
		}
	}
	return m
}

func (p *Package) Location(pos token.Pos) token.Position {
	return p.project.fset.Position(pos)
}

func (p *Package) Imports() []string {
	var imports []string
	for _, i := range p.file.Imports {
		path := i.Path.Value
		imports = append(imports, path[1:len(path)-1])
	}
	return imports
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
func (p *Project) load(pkgPath string) (err error) {
	if _, ok := p.packages[pkgPath]; ok {
		return
	}
	if pkgPath == "C" {
		fmt.Println("Added virtual package", pkgPath)
		// virtual package, set it to nil
		p.packages[pkgPath] = nil
		return
	}
	buildPkg, err := build.Default.Import(pkgPath, ".", 0)
	if err != nil {
		return
	}
	goFiles := make(map[string]bool)
	for _, f := range buildPkg.GoFiles {
		goFiles[f] = true
	}
	pkgs, err := parser.ParseDir(p.fset, buildPkg.Dir, func(f os.FileInfo) bool {
		return goFiles[f.Name()]
	}, 0)

	if err != nil {
		return
	}

	var pkgName string
	idx := strings.LastIndex(pkgPath, "/")
	if idx == -1 {
		pkgName = pkgPath
	} else {
		pkgName = pkgPath[idx+1:]
	}
	//fmt.Println("Looking for package", pkgName)
	for name, pkg := range pkgs {
		// ignore main and *_test packages
		if name != pkgName && name != "main" {
			continue
		}
		var scopes []*ast.Scope
		for _, f := range pkg.Files {
			scopes = append(scopes, f.Scope)
			f.Scope.Outer = nil
		}
		mergedFile := ast.MergePackageFiles(pkg, ast.FilterFuncDuplicates|ast.FilterImportDuplicates)
		p.packages[pkgPath] = &Package{project: p, name: name, file: mergedFile, scopes: scopes}
		return
	}
	return fmt.Errorf("Package %s not found in %s", pkgName, buildPkg.Dir)
}

func (p *Project) get(pkgName string) (pkg *Package) {
	pkg = p.packages[pkgName]
	return
}

var outputConfig = &printer.Config{
	Mode:     printer.TabIndent | printer.UseSpaces,
	Tabwidth: 8,
}

func (p *Project) write(dir string, pkg *Package) (err error) {
	srcDir := path.Join(dir, "src")
	os.Mkdir(srcDir, 0777)
	pkgName := pkg.name

	var f *os.File
	if pkgName == "main" {
		pkgName = p.name
	}
	pkgPath := path.Join(srcDir, pkgName)
	os.Mkdir(pkgPath, 0777)
	filePath := path.Join(pkgPath, pkgName+"_gopar.go")
	f, err = os.Create(filePath)
	if err != nil {
		return
	}
	fmt.Println("Writing", f.Name())
	outputConfig.Fprint(f, p.fset, pkg.file)
	f.Close()

	return
}
