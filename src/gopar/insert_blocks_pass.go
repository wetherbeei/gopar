// Insert blocks pass
//
// Replaces the parallel loop in the AST with the structure for inserting the
// parallel code

package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

const GOPAR_RTLIB = "gopar_rtlib"

type InsertBlockVisitor struct {
	data   *ParallelizeData
	parent ast.Node
}

func (v InsertBlockVisitor) Visit(node ast.Node) (w ast.Visitor) {
	var data *ParallelLoopInfo
	var ok bool
	if data, ok = v.data.loops[node]; !ok {
		v.parent = node
		return v
	}

	// find the node's position in the parent

	newBlock := &ast.BlockStmt{}
	switch t := v.parent.(type) {
	case *ast.BlockStmt:
		// search the whole statement
		for idx, b := range t.List {
			if node == b {
				t.List[idx] = newBlock
			}
		}
	case *ast.IfStmt:
		if node == t.Body {
			t.Body = newBlock
		}
	case *ast.RangeStmt:
		if node == t.Body {
			t.Body = newBlock
		}
	case *ast.ForStmt:
		if node == t.Body {
			t.Body = newBlock
		}
	default:
		if *verbose {
			fmt.Printf("Unknown parent %T\n", v.parent)
		}
	}

	// insert a new empty block and record it
	data.block = newBlock

	// insert runtime parallel conditions
	newBlock.List = append(newBlock.List, &ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{&ast.Ident{Name: "__parallel"}},
		Rhs: []ast.Expr{&ast.Ident{Name: "true"}},
	})
	testBlock := &ast.BlockStmt{}
	newBlock.List = append(newBlock.List, testBlock)
	data.tests = testBlock

	// insert if statement, record empty new parallel block
	parallelBlock := &ast.BlockStmt{}
	data.parallel = parallelBlock
	newBlock.List = append(newBlock.List, &ast.IfStmt{
		Cond: &ast.Ident{Name: "__parallel"},
		Body: parallelBlock,
		// insert else statement and reattach sequential loop
		Else: &ast.BlockStmt{List: []ast.Stmt{data.sequential}},
	})

	v.parent = node
	return v
}

type InsertBlocksPass struct {
	BasePass
}

func NewInsertBlocksPass() *InsertBlocksPass {
	return &InsertBlocksPass{
		BasePass: NewBasePass(),
	}
}

func (pass *InsertBlocksPass) GetPassType() PassType {
	return InsertBlocksPassType
}

func (pass *InsertBlocksPass) GetPassMode() PassMode {
	return ModulePassMode
}

func (pass *InsertBlocksPass) GetDependencies() []PassType {
	return []PassType{ParallelizePassType}
}

func (pass *InsertBlocksPass) RunModulePass(file *ast.File, p *Package) (modified bool, err error) {
	data := pass.compiler.GetPassResult(ParallelizePassType, p).(*ParallelizeData)
	if len(data.loops) > 0 {
		rtimport := &ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{&ast.ImportSpec{Name: &ast.Ident{Name: GOPAR_RTLIB}, Path: &ast.BasicLit{Kind: token.STRING, Value: `"rtlib"`}}}}
		file.Decls = append(file.Decls, nil)
		copy(file.Decls[2:], file.Decls[1:])
		file.Decls[1] = rtimport
		v := InsertBlockVisitor{data: data}
		ast.Walk(v, p.file)
	}
	return // TODO: return true??
}
