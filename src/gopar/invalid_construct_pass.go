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

func (pass *InvalidConstructPass) RunFunctionPass(node ast.Node, p *Package) (modified bool, err error) {
	var external []string
	ast.Inspect(node, func(node ast.Node) bool {
		if node != nil {
			switch t := node.(type) {
			case *ast.CallExpr:
				switch f := t.Fun.(type) {
				case *ast.Ident:
					// Check if the function is builtin, or defined in the package
					var name string = f.Name
					if p.Lookup(name) == nil {
						if !builtinTranslated[name] {
							fmt.Println("Untranslatable function", name)
							external = append(external, name)
						}
					} else {
						fmt.Println("Found supporting function", name)
					}
				default:
					fmt.Println("Unsupported function call", f)
					external = append(external, "<anonymous>")
				}
			case *ast.FuncDecl:
				fmt.Println("Embedded function", t.Name)
				external = append(external, t.Name.String())
			case *ast.SelectStmt:
				fmt.Println("Select stmt")
				external = append(external, "select stmt")
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
