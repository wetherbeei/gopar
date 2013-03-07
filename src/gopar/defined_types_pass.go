// Defined types pass
//
// Gather all user-defined types for a package

package main

import (
	"fmt"
	"go/ast"
)

type DefinedTypesData struct {
	defined map[string]Type
}

func NewDefinedTypesData() *DefinedTypesData {
	builtin := []string{
		"uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64",
		"float32", "float64", "complex64", "complex128", "uint", "int", "uintptr",
		"rune", "byte", // aliases
	}
	d := &DefinedTypesData{
		defined: make(map[string]Type),
	}
	for _, ident := range builtin {
		d.defined[ident] = Type{&ast.Ident{Name: ident}}
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
	for _, decl := range file.Decls {
		switch t := decl.(type) {
		case *ast.FuncDecl:
			data.defined[t.Name.Name] = NewType(t.Type)
		case *ast.GenDecl:
			for _, spec := range t.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					var name = s.Name.Name
					data.defined[name] = NewType(s)
				}
			}
		default:
			fmt.Printf("Unhandled Decl %T %+v", decl, decl)
		}
	}
	pass.SetResult(nil, data)
	return
}
