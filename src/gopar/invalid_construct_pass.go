// Invalid constructs pass
//
// 
package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

type InvalidConstructPass struct {
	BasePass
}

func (pass *InvalidConstructPass) RunFunctionPass(node ast.Node, c *Compiler) (modified bool, err error) {
	var external []string
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
			case *ast.FuncDecl:
				fmt.Println("Embedded function", t.Name)
				external = append(external, t.Name.String())
			case *ast.SwitchStmt:
				fmt.Println("Switch stmt")
				external = append(external, "switch stmt")
			case *ast.GoStmt:
				fmt.Println("Go stmt")
				external = append(external, "go stmt")
			case *ast.BranchStmt:
				if t.Tok == token.GOTO {
					fmt.Println("Goto stmt")
					external = append(external, "goto stmt")
				}
			case *ast.DeferStmt:
				fmt.Println("Defer stmt")
				external = append(external, "defer stmt")
			}
			return true
		}
		return false
	})
	fmt.Println("External dependencies", node.(*ast.FuncDecl).Name, external)
	pass.SetResult(node, external)
	return false, nil
}
