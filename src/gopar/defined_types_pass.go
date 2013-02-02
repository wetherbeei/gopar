// Defined types pass
//
// Gather all user-defined types for a package

package main

import (
	"fmt"
	"go/ast"
)

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

func (pass *DefinedTypesPass) RunModulePass(file *ast.File, c *Compiler) (modified bool, err error) {
	for _, decl := range file.Decls {
		fmt.Printf("%T %v\n", decl, decl)
	}
	return
}
