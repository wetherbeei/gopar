// Write kernels pass
//
// Generate the kernel source for loops, and write them to a text constant

package main

import (
	"fmt"
	"go/ast"
)

type WriteKernelsPass struct {
	BasePass
}

func NewWriteKernelsPass() *WriteKernelsPass {
	return &WriteKernelsPass{
		BasePass: NewBasePass(),
	}
}

func (pass *WriteKernelsPass) GetPassType() PassType {
	return WriteKernelsPassType
}

func (pass *WriteKernelsPass) GetPassMode() PassMode {
	return ModulePassMode
}

func (pass *WriteKernelsPass) GetDependencies() []PassType {
	return []PassType{InsertBlocksPassType}
}

func (pass *WriteKernelsPass) RunModulePass(file *ast.File, p *Package) (modified bool, err error) {
	data := pass.compiler.GetPassResult(ParallelizePassType, nil).(*ParallelizeData)

	for _, loop := range data.loops {
		fmt.Println(loop.arguments)
	}
	return
}
