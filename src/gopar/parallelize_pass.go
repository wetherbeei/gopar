// Parallelize pass (finally)
//
// Range statements must be made up of only:
// ReadOnly for any variable
// ReadWrite to any variable indexed by the range index
// WriteFirst access to any variable (make a private copy)
// - for privatizing, copy the last iteration's value into a variable
//     i = 0
//     for i, v = range array {
//     }
//     i and v should equal 

package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type ParallelizePass struct {
	BasePass
}

type ParallelLoopInfo struct {
	name      string       // generated name of this kernel
	indexVar  string       // the unique index for iterating ex: idx in (for idx := range slice {})
	arguments []Dependency // variables to be copied to/from the kernel
	//privatize   []Dependency // variables to be privatized for each loop
	start, stop ast.Expr   // an expression representing the number of iterations, inclusive
	step        ast.Expr   // an expression for the step, "idx" will be replaced with the thread index
	variables   []ast.Stmt // values to set for each loop iteration, including the index and value vars
	// these variables hold references to ast blocks for inserting new code
	//
	// { // .block
	//   __parallel := false
	//   {} // .tests - runtime tests if this loop is parallel
	//   if __parallel {
	//     // .parallel
	//   } else {
	//     // .sequential - holds the original sequential loop
	//   }
	// }
	block, tests, parallel *ast.BlockStmt
	sequential             ast.Stmt // the original node
	kernelSource           string   // the generated OpenCL source
}

type ParallelizeData struct {
	// tag with map[ast.RangeStmt]s = some data
	loops map[ast.Node]*ParallelLoopInfo
}

func NewParallelizeData() *ParallelizeData {
	return &ParallelizeData{
		loops: make(map[ast.Node]*ParallelLoopInfo),
	}
}

func NewParallelizePass() *ParallelizePass {
	return &ParallelizePass{
		BasePass: NewBasePass(),
	}
}

func (pass *ParallelizePass) GetPassType() PassType {
	return ParallelizePassType
}

func (pass *ParallelizePass) GetPassMode() PassMode {
	return BasicBlockPassMode
}

func (pass *ParallelizePass) GetDependencies() []PassType {
	return []PassType{DependencyPassType, InvalidConstructPassType}
}

func allowedType(t Type) (err error) {
	switch s := t.(type) {
	case *BaseType:
		// TODO: string
	case *StructType:
		// check fields
		for _, name := range s.Fields() {
			if err = allowedType(s.Field(name)); err != nil {
				return
			}
		}
	case *IndexedType:
		// check value
		return allowedType(s.IndexValue())
	default:
		return fmt.Errorf("Unsupported type on GPU: %v", t)
	}
	return nil
}

func canParallelize(loop *BasicBlock) (info *ParallelLoopInfo, err error) {
	var block *BasicBlock
	var dependencyData *DependencyPassData
	var invalidData []string
	//var definesData *AccessPassData = 
	switch t := loop.node.(type) {
	case *ast.RangeStmt:
		idxIdent, ok := t.Key.(*ast.Ident)
		if !ok {
			return
		}
		//block = loop.children[0]
		block = loop
		// examine the dependencies of the loop body
		dependencyData = block.Get(DependencyPassType).(*DependencyPassData)
		invalidData = block.Get(InvalidConstructPassType).([]string)
		info = &ParallelLoopInfo{
			sequential: loop.node.(ast.Stmt),
			indexVar:   idxIdent.Name,
			start: &ast.BasicLit{
				Kind:  token.INT,
				Value: "0",
			},
			// len(X)-1
			stop: &ast.BinaryExpr{
				X: &ast.CallExpr{
					Fun:  &ast.Ident{Name: "len"},
					Args: []ast.Expr{t.X},
				},
				Y:  &ast.BasicLit{Kind: token.INT, Value: "-1"},
				Op: token.SUB,
			},
			step: &ast.Ident{
				Name: "idx", // no step for range statements
			},
		}

		// set value variable
		if t.Value != nil {
			info.variables = append(info.variables, &ast.AssignStmt{
				Lhs: []ast.Expr{t.Value},
				Rhs: []ast.Expr{&ast.IndexExpr{
					X:     t.X,
					Index: idxIdent,
				}},
				Tok: token.ASSIGN,
			})
		}
	// case *ast.ForStmt:
	default:
		return // not a loop
	}

	if len(invalidData) > 0 {
		err = fmt.Errorf("Untranslatable loop: %s", invalidData)
		return
	}

	// Generate a list of all leaked dependencies
	// Remove the defined key, val variables (use .defined data)
	// Check array accesses for [key]
	loop.Print("== Dependencies ==")
	for _, dep := range dependencyData.deps {
		loop.Print(dep.String())
		switch dep.depType {
		case ReadOnly:
			// nothing
			info.arguments = append(info.arguments, dep)
		case ReadWrite, WriteFirst:
			// each read and write must be indexed by the iteration variable
			// a[1][idx] ??
			iterationOnly := false
			for _, part := range dep.group {
				if part.isIndexed {
					if part.index == info.indexVar {
						iterationOnly = true
						break // everything under this is fine
					} else {
						err = fmt.Errorf("%s crosses iteration bounds with index '%s'", dep.String(), part.index)
						return
					}
				}
			}
			if !iterationOnly {
				err = fmt.Errorf("%s is accessed by all iterations", dep.String())
				return
			}
			info.arguments = append(info.arguments, dep)
			//case WriteFirst:
			// only privatize simple variables

			//info.privatize = append(info.privatize, dep)
		}
	}

	// Step up one block level and remove slice accesses to figure out which
	// slices need to be transfered completely

	// Check types of all arguments
	for _, dep := range info.arguments {
		if err = allowedType(dep.goType); err != nil {
			return
		}
	}
	return
}

func (pass *ParallelizePass) RunBasicBlockPass(block *BasicBlock, p *Package) BasicBlockVisitor {
	var data *ParallelizeData
	if data, _ = pass.GetResult(nil).(*ParallelizeData); data == nil {
		data := NewParallelizeData()
		pass.SetResult(nil, data)
	}
	var info *ParallelLoopInfo
	var err error
	info, err = canParallelize(block)

	if err != nil {
		block.Printf("\x1b[31;1mCan't parallelize loop\x1b[0m at %s", p.Location(block.node.Pos()))
		block.Printf("-> %s", err.Error())
	} else if info != nil {
		pos := p.Location(block.node.Pos())
		info.name = fmt.Sprintf("%s_%d", strings.Replace(strings.Replace(pos.Filename, ".", "_", -1), "/", "_", -1), pos.Line)
		block.Printf("\x1b[33;1mParallel loop\x1b[0m named %s", info.name)
		block.Printf("Thread index = '%s'", info.indexVar)
		data.loops[block.node] = info
	}
	return DefaultBasicBlockVisitor{}
}
