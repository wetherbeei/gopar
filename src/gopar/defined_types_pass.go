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
	defined map[string]Type
}

func NewDefinedTypesData() *DefinedTypesData {
	d := &DefinedTypesData{
		defined: make(map[string]Type),
	}

	for k, v := range BuiltinTypes {
		d.defined[k] = v
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
	for _, decl := range file.Decls {
		switch t := decl.(type) {
		case *ast.FuncDecl:
			if t.Recv != nil {
				methods = append(methods, t)
			} else {
				defined[t.Name.Name] = TypeDecl(t)
			}
		case *ast.GenDecl:
			for _, spec := range t.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					var name = s.Name.Name
					defined[name] = TypeDecl(s.Type)
				case *ast.ImportSpec:
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
					if name != "_" {
						// attach the types found in the other package here
						otherPackage := pass.compiler.project.get(path)
						otherPackageTypes, ok := pass.compiler.GetPassResult(DefinedTypesPassType, otherPackage).(*DefinedTypesData)

						var packageResolver Resolver
						if ok {
							packageResolver = func(name string) Type {
								return otherPackageTypes.defined[name]
							}
						} else {
							// package not found
							fmt.Println("Package not found:", path)
							packageResolver = func(name string) Type {
								return nil
							}
						}
						defined[name] = TypeDecl(s)
						defined[name].Complete(packageResolver)
					}
				}
			}
		default:
			fmt.Printf("Unhandled Decl %T %+v", decl, decl)
		}
	}

	pass.SetResult(p, data)
	fmt.Println(data.defined)

	resolver := MakeResolver(nil, p, pass.compiler)
	// fill in embedded fields
	for name, typ := range defined {
		typ.Complete(resolver)
		data.defined[name] = typ
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
