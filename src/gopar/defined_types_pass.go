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
	for _, decl := range file.Decls {
		switch t := decl.(type) {
		case *ast.FuncDecl:
			if t.Recv != nil {
				methods = append(methods, t)
			} else {
				data.defined[t.Name.Name] = TypeDecl(t.Type)
			}
		case *ast.GenDecl:
			for _, spec := range t.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					var name = s.Name.Name
					data.defined[name] = TypeDecl(s.Type)
				case *ast.ImportSpec:
					var name string
					if s.Name != nil {
						name = s.Name.Name
					} else {
						// pull from import path
						path := s.Path.Value
						idx := strings.LastIndex(path, "/")
						if idx == -1 {
							name = path
						} else {
							name = path[idx:]
						}
					}
					name = name[1 : len(name)-1] // strip off " in the back and /" in front
					if name != "_" {
						data.defined[name] = TypeDecl(s)
					}
				}
			}
		default:
			fmt.Printf("Unhandled Decl %T %+v", decl, decl)
		}
	}

	var resolver Resolver
	resolver = func(name string) Type {
		return data.defined[name]
	}

	pass.SetResult(nil, data)
	fmt.Println(data.defined)
	// fill in embedded fields
	for _, typ := range data.defined {
		typ.Complete(resolver)
	}

	// fill in methods
	//for _, method := range methods {
	//NewType(method.Recv.List[0].Type)
	//}

	for name, typ := range data.defined {
		fmt.Printf("%s = %v\n", name, typ)
	}
	return
}
