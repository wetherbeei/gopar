// Defined types pass
//
// Gather all user-defined types for a package

package main

import (
	"fmt"
	"go/ast"
	"strings"
)

type DefinedTypesData struct {
	defined  map[string]Type
	embedded []Type // used when other packages import . "something"
}

func NewDefinedTypesData() *DefinedTypesData {
	d := &DefinedTypesData{
		defined: make(map[string]Type),
	}
	return d
}

type DefinedTypesPass struct {
	BasePass
}

func NewDefinedTypesPass() *DefinedTypesPass {
	return &DefinedTypesPass{
		BasePass: NewBasePass(),
	}
}

func (pass *DefinedTypesPass) GetPassType() PassType {
	return DefinedTypesPassType
}

func (pass *DefinedTypesPass) GetPassMode() PassMode {
	return ModulePassMode
}

func (pass *DefinedTypesPass) GetDependencies() []PassType {
	return []PassType{}
}

func (pass *DefinedTypesPass) RunModulePass(file *ast.File, p *Package) (modified bool, err error) {
	data := NewDefinedTypesData()
	var methods []*ast.FuncDecl
	var defined map[string]Type = make(map[string]Type)
	Define := func(name string, t Type) {
		if val, exists := defined[name]; exists {
			// if they're packages with the same name, ignore it
			pkg1, ok1 := val.(*PackageType)
			pkg2, ok2 := val.(*PackageType)
			if !ok1 || !ok2 || pkg1.path != pkg2.path {
				err = fmt.Errorf("Redefining identifier %s = %s with %s", name, val.String(), t.String())
				return
			}
		}
		defined[name] = t // TODO: write directly to data.defined?
	}
	// Top-level definitions don't have to be done in any order, so use the 
	// FutureType
	for _, decl := range file.Decls {
		switch t := decl.(type) {
		case *ast.FuncDecl:
			if t.Recv != nil {
				methods = append(methods, t)
			} else {
				Define(t.Name.Name, TypeDecl(t))
			}
		case *ast.GenDecl:
			var prev ast.Expr // used for constants
			for _, spec := range t.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					var name = s.Name.Name
					Define(name, TypeDecl(s.Type))
				case *ast.ValueSpec:

					if s.Type != nil {
						for _, name := range s.Names {
							Define(name.Name, newFutureType(s.Type))
						}
					} else {
						// constants
						// might be iota declarations, so if s.Values is missing then use
						// the previous declaration
						for i, name := range s.Names {
							if i < len(s.Values) {
								prev = s.Values[i]
							}
							Define(name.Name, newFutureType(prev))
						}
					}
				case *ast.ImportSpec:
					// TODO: package imports should only be for this file
					// http://golang.org/ref/spec#Declarations_and_scope
					var name string
					var path string
					if s.Name != nil {
						name = s.Name.Name
						path = name
					} else {
						// pull from import path
						path = s.Path.Value
						idx := strings.LastIndex(path, "/")
						if idx == -1 {
							name = path
						} else {
							name = path[idx:]
						}
					}
					name = name[1 : len(name)-1] // strip off ["] in the back and [/"] in front
					path = path[1 : len(path)-1]
					if name != "_" {
						// attach the types found in the other package here
						otherPackage := pass.compiler.project.get(path)
						otherPackageTypes, ok := pass.compiler.GetPassResult(DefinedTypesPassType, otherPackage).(*DefinedTypesData)

						var packageResolver Resolver
						if ok {
							packageResolver = func(name string) Type {
								pkgTyp := otherPackageTypes.defined[name]
								fmt.Printf("Cross-package lookup: %s.%s = %s\n", path, name, pkgTyp)
								return pkgTyp
							}
						} else {
							// package not found
							// TODO: make this an error?
							fmt.Println("Package not found:", path)
							packageResolver = func(name string) Type {
								fmt.Println("Empty package resolver", name)
								return nil
							}
						}
						pkgType := TypeDecl(s)
						pkgType.Complete(packageResolver)
						if name == "." {
							data.embedded = append(data.embedded, pkgType)
						} else {
							Define(name, pkgType)
						}
					}
				}
			}
		default:
			fmt.Printf("Unhandled Decl %T %+v", decl, decl)
		}
	}

	pass.SetResult(p, data)

	resolver := MakeResolver(nil, p, pass.compiler)
	// add all of the defines to the package scope so they can be found from
	// .Complete()
	for name, typ := range defined {
		data.defined[name] = typ
	}

	// fill in embedded fields
	for _, typ := range defined {
		typ.Complete(resolver)
	}

	// fill in methods
	for _, method := range methods {
		recvTyp := TypeOfDecl(method.Recv.List[0].Type, resolver)

		methodTyp := newMethodType(method.Type, method, recvTyp)
		methodTyp.Complete(resolver)
		fmt.Printf("Adding method %+v to %+v\n", method, methodTyp)
		recvTyp.AddMethod(method.Name.Name, methodTyp)
	}

	for name, typ := range data.defined {
		fmt.Printf("%s = %v\n", name, typ)
	}
	return
}
