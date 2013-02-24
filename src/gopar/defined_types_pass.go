// Defined types pass
//
// Gather all user-defined types for a package

package main

import (
	"fmt"
	"go/ast"
)

type DefinedType struct {
	ident string
	// Link to type definition, nil if builtin
	decl ast.Node
}

type DefinedTypesData struct {
	defined map[string]*DefinedType
}

func NewDefinedTypesData() *DefinedTypesData {
	builtin := []string{
		"uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64",
		"float32", "float64", "complex64", "complex128", "uint", "int", "uintptr",
		"rune", "byte", // aliases
	}
	d := &DefinedTypesData{
		defined: make(map[string]*DefinedType),
	}
	for _, ident := range builtin {
		d.defined[ident] = &DefinedType{ident: ident}
	}

	return d
}

func (d *DefinedTypesData) Get(ident string) *DefinedType {
	return d.defined[ident]
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
		case *ast.GenDecl:
			for _, spec := range t.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					var name = s.Name.Name
					data.defined[name] = &DefinedType{ident: name, decl: s}
					fmt.Printf("New type %s = %T %v\n", name, s, s)
				}
			}
		}
	}
	pass.SetResult(nil, data)
	return
}
