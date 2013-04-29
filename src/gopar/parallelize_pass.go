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
	start, stop ast.Expr     // an expression representing the number of iterations, inclusive start, exclusive stop
	variables   []ast.Stmt   // values to set for each loop iteration, including the index and value vars
	alias       []Dependency // values to be compared for aliasing
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
	sequential             ast.Stmt       // the original node
	kernel                 *ast.BlockStmt // the kernel to be generated
	kernelSource           string         // the generated OpenCL source
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
	return []PassType{DependencyPassType}
}

func allowedType(t Type) (err error) {
	if !t.PassByValue() {
		return fmt.Errorf("Unsupported type (pass-by-reference): %v", t)
	}
	switch s := t.(type) {
	case *StructType:
		// check fields
		for _, name := range s.Fields() {
			if err = allowedType(s.Field(name)); err != nil {
				return
			}
		}
	case *IndexedType:
		// only allow full arrays
		return allowedType(s.IndexValue())
	}
	return nil
}

func canParallelize(loop *BasicBlock, resolver Resolver) (info *ParallelLoopInfo, err error) {
	var block *BasicBlock
	var dependencyData *DependencyPassData
	var listIg IdentifierGroup
	var listTyp Type
	var list ast.Expr

	switch t := loop.node.(type) {
	case *ast.RangeStmt:
		idxIdent, ok := t.Key.(*ast.Ident)
		if !ok {
			err = fmt.Errorf("Range must have an explicit index variable")
			return
		}
		//block = loop.children[0]
		block = loop
		// examine the dependencies of the loop body
		dependencyData = block.Get(DependencyPassType).(*DependencyPassData)
		list = t.X
		err = AccessIdentBuild(&listIg, list, nil)
		listTyp = TypeOf(list, resolver)
		if err != nil {
			return
		}

		idx := idxIdent.Name
		if idx == "_" {
			idx = ""
		}
		info = &ParallelLoopInfo{
			sequential: loop.node.(ast.Stmt),
			indexVar:   idx,
			start: &ast.BasicLit{
				Kind:  token.INT,
				Value: "0",
			},
			// len(X)
			stop: &ast.CallExpr{
				Fun:  &ast.Ident{Name: "len"},
				Args: []ast.Expr{list},
			},
			kernel: t.Body,
		}

		if t.Key != nil && len(info.indexVar) > 0 {
			info.variables = append(info.variables, &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: info.indexVar}},
				Rhs: []ast.Expr{&ast.Ident{Name: "_idx"}},
				Tok: token.DEFINE,
			})
		}

		// set value variable
		if t.Value != nil {
			valueIdx := info.indexVar
			if len(info.indexVar) == 0 {
				valueIdx = "_idx"
			}
			info.variables = append(info.variables, &ast.AssignStmt{
				Lhs: []ast.Expr{t.Value},
				Rhs: []ast.Expr{&ast.IndexExpr{
					X:     list,
					Index: &ast.Ident{Name: valueIdx},
				}},
				Tok: token.DEFINE,
			})
		}
	// case *ast.ForStmt:
	default:
		return // not a loop
	}

	// Fill in type d
	// Check parallelization conditions
	// - Check that array reads/writes are only to [indexVar]
	// - Check that there are no reads from the list outside of [indexVar]

	blockDefines := block.Get(AccessPassType).(*AccessPassData).defines
	for _, dep := range dependencyData.deps {
		// ignore variables defined in the loop statement
		if _, ok := blockDefines[dep.group[0].id]; ok {
			continue
		}

		switch dep.depType {
		case ReadOnly, ReadWrite:
			// nothing
			for i, listPart := range listIg.group {
				if i < len(dep.group) {
					if dep.group[i].id != listPart.id {
						break
					}
					if i == len(listIg.group)-1 {
						if !dep.group[i].isIndexed || dep.group[i].index != info.indexVar {
							err = fmt.Errorf("Cannot access loop data '%s' with index '%s'",
								dep.String(), dep.group[i].index)
							return
						}
					}
				} else {
					break
				}
			}
		}
	}

	for _, dep := range dependencyData.deps {
		// ignore variables defined in the loop statement
		if _, ok := blockDefines[dep.group[0].id]; ok {
			continue
		}

		switch dep.depType {
		case ReadOnly:
			// nothing
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
			//case WriteFirst:
			// TODO: only privatize simple variables

			//info.privatize = append(info.privatize, dep)
		}
	}

	// Generate a list of all leaked dependencies to pass as arguments
	// Remove the defined key, val variables (use .defined data)
	for _, dep := range dependencyData.deps {
		if _, ok := blockDefines[dep.group[0].id]; !ok {
			// wasn't defined in the loop definition, it's an external dep

			// check if it's index was defined too, usually a[idx]
			var newDep Dependency = dep
			for idx, ident := range dep.group {
				if _, ok := blockDefines[ident.index]; ok && ident.isIndexed {
					// cut off array at this point
					newDep.group = make([]Identifier, idx+1)
					copy(newDep.group, dep.group)
					newDep.group[idx].isIndexed = false
					newDep.group[idx].index = ""
					break
				}
			}
			// check if arguments already has this dep since we may be creating a new
			// copy
			var matched bool
			for j, dep := range info.arguments {
				if len(dep.group) != len(newDep.group) {
					continue // not an exact match
				}
				for i := 0; i < len(dep.group); i++ {
					if !dep.group[i].Equals(newDep.group[i]) {
						break
					}
					// if exact match
					if i == len(dep.group)-1 {
						matched = true
						// reclassify
						if dep.depType == ReadOnly && (newDep.depType == ReadWrite || newDep.depType == WriteFirst) {
							info.arguments[j].depType = ReadWrite
						}
					}
				}
			}
			if !matched {
				info.arguments = append(info.arguments, newDep)
			}
		}
	}

	// Check for a.b.c arguments (TODO remove this requirement)
	/*
		for _, dep := range info.arguments {
			if len(dep.group) > 1 {
				err = fmt.Errorf("Selectors not allowed as arguments: %s", dep.String())
				return
			}
			if dep.group[0].isIndexed {
				err = fmt.Errorf("Index expressions not allowed as arguments: %s", dep.String())
			}
		}
	*/

	// Check types of all arguments
	for i, dep := range info.arguments {
		if *verbose {
			block.Print(dep.String())
		}
		if dep.group[0].id == "$external" {
			continue // can't resolve these
			// TODO: what if external accesses alias?
		}
		dep.goType = TypeOf(dep.MakeNode(), resolver)
		// see if this argument is also pass-by-reference (slice only), and add it
		// to the aliased list
		if !dep.goType.PassByValue() {
			if _, ok := dep.goType.(*IndexedType); ok {
				if err = allowedType(dep.goType.IndexValue()); err != nil {
					err = fmt.Errorf("Dependency %s contains invalid value: %s", dep.String(), err.Error())
					return
				}
				info.alias = append(info.alias, dep)
				continue
			}
		}
		if err = allowedType(dep.goType); err != nil {
			err = fmt.Errorf("%s: %s", dep.String(), err.Error())
			return
		}
		info.arguments[i] = dep
	}

	// Check array type
	// TODO: not a map or channel
	var listDep Dependency
	listDep.group = listIg.group
	listDep.depType = ReadWrite
	listDep.goType = listTyp
	if err = allowedType(listTyp.IndexValue()); err != nil {
		err = fmt.Errorf("Loop %s contains invalid value: %s", listDep.String(), err.Error())
		return
	}

	info.alias = append(info.alias, listDep)
	return
}

func (pass *ParallelizePass) RunBasicBlockPass(block *BasicBlock, p *Package) BasicBlockVisitor {
	var data *ParallelizeData
	if data, _ = pass.GetResult(p).(*ParallelizeData); data == nil {
		data := NewParallelizeData()
		pass.SetResult(p, data)
	}
	var info *ParallelLoopInfo
	var err error
	resolver := MakeResolver(block, p, pass.compiler)
	info, err = canParallelize(block, resolver)
	if err != nil {
		fmt.Printf("\n\x1b[31;1mCan't parallelize loop\x1b[0m at %s\n", p.Location(block.node.Pos()))
		fmt.Printf("-> %s\n", err.Error())
	} else if info != nil {
		pos := p.Location(block.node.Pos())
		info.name = fmt.Sprintf("%s_%d", strings.Replace(strings.Replace(pos.Filename, ".", "_", -1), "/", "_", -1), pos.Line)
		fmt.Printf("\n\x1b[33;1mParallel loop\x1b[0m at %s:%d\n", pos.Filename, pos.Line)
		fmt.Printf("Thread index = '%s'\n", info.indexVar)
		fmt.Printf("Arguments:\n")
		for _, arg := range info.arguments {
			fmt.Printf("  %s\n", arg.String())
		}
		fmt.Printf("Alias checks:\n")
		for _, arg := range info.alias {
			fmt.Printf("  %s\n", arg.String())
		}
		data.loops[block.node] = info
	}
	return DefaultBasicBlockVisitor{}
}
