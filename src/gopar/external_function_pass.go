package main

import (
	"fmt"
	"go/ast"
)

// External function analysis pass
//
// Examine the function for any external function calls and record a list of
// them.

// TODO: These functions have a corresponding GPU implementation
var builtinTranslated map[string]bool = map[string]bool{
// Detect constant make lengths and privatize an array
// "make": bool
}

type ExternalFunctionPass struct {
	BasePass
}

func NewExternalFunctionPass() *ExternalFunctionPass {
	return &ExternalFunctionPass{
		BasePass: NewBasePass(),
	}
}

func (pass *ExternalFunctionPass) GetPassType() PassType {
	return ExternalFunctionPassType
}

func (pass *ExternalFunctionPass) GetPassMode() PassMode {
	return FunctionPass
}

func (pass *ExternalFunctionPass) GetDependencies() []PassType {
	return []PassType{}
}

func (pass *ExternalFunctionPass) RunFunctionPass(node ast.Node, c *Compiler) (modified bool, err error) {
	var external []string
	fmt.Println("Inspecting function", node.(*ast.FuncDecl).Name)
	ast.Inspect(node, func(node ast.Node) bool {
		if node != nil {
			switch t := node.(type) {
			case *ast.CallExpr:
				switch f := t.Fun.(type) {
				case *ast.FuncLit:
					// TODO: Allow anonymous functions??
					fmt.Println("Anonymous function")
					external = append(external, "<anonymous>")
				case *ast.Ident:
					// Check if the function is builtin, or defined in the package
					var name string = f.Name
					if c.project.get("main").Lookup(name) == nil {
						if !builtinTranslated[name] {
							fmt.Println("Untranslatable function", name)
							external = append(external, name)
						}
					} else {
						fmt.Println("Found supporting function", name)
					}
				}
			}
			return true
		}
		return false
	})
	fmt.Println("External dependencies", node.(*ast.FuncDecl).Name, external)
	pass.SetResult(node, external)
	return false, nil
}
