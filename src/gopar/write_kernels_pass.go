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

func generateOpenCL(data *ParallelizeData, p *Package) (kernel string, err error) {
	// parse the loop body, generate a list of all data structures to define and
	// all functions called
	return
}

func (pass *WriteKernelsPass) RunModulePass(file *ast.File, p *Package) (modified bool, err error) {
	data := pass.compiler.GetPassResult(ParallelizePassType, nil).(*ParallelizeData)

	//kernel, err := generateOpenCL(data, p)

	for _, loop := range data.loops {
		fmt.Println(loop.arguments)
	}
	return
}
