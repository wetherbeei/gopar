// Write kernels pass
//
// Generate the kernel source for loops, and write them to a text constant
//
// Conversion between Go and C:
// If, range, for, switch init declarations are wrapped in another closure:
// {
//    a = 1
//    if a {
//
//    }
// }
//
// Assign := statements declare their variable type first
// int a;
// a = 1; // a := 1
//
// Function call return values are passed as pointer arguments on the end of a
// call.
// a, b, c := func(1, 2)
// int a, b, c;
// func(1, 2, &a, &b, &c);
//
// Switch statements have implicit breaks in Go, and explicitly declare 
// "fallthrough"
//
// 
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"io"
	"strings"
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
	return []PassType{ParallelizePassType}
}

type CWriter struct {
	writer *bytes.Buffer
	indent int
}

func NewCWriter() *CWriter {
	return &CWriter{writer: &bytes.Buffer{}}
}

func (w *CWriter) Writeln(str string) (n int, err error) {
	io.WriteString(w.writer, strings.Repeat(" ", w.indent))
	return io.WriteString(w.writer, str+"\n")
}

func (w *CWriter) Writelns(str string) (n int, err error) {
	for {
		index := strings.Index(str, "\n")
		if index == -1 {
			break
		}
		w.Writeln(str[:index])
		str = str[index+1:]
	}
	return
}

func (w *CWriter) Indent() {
	w.indent += 2
}

func (w *CWriter) Dedent() {
	w.indent -= 2
}

func (w *CWriter) String() string {
	return w.writer.String()
}

const opencl_header = `
/* Header */

`

func generateOpenCL(data *ParallelLoopInfo, c *Compiler) (kernel string, err error) {
	// parse the loop body
	// - define main kernel function with all args used
	// - TOOD: types/data structures
	// - TODO: other functions
	writer := NewCWriter()
	writer.Writeln(fmt.Sprintf("/* Generated OpenCL for %s */", data.name))
	writer.Writelns(opencl_header)
	writer.Writeln(fmt.Sprintf("__kernel void %s(__global", data.name))
	writer.Indent()
	for _, arg := range data.arguments {
		writer.Writeln(fmt.Sprintf("%s %s,", arg.goType.CType(), arg.group[0].id))
	}
	writer.Dedent()
	writer.Writeln(") {")
	generateBlock(data.block, writer)
	writer.Writeln("}")
	return writer.String(), nil
}

func generateBlock(block ast.Node, writer *CWriter) {
	writer.Indent()
	writer.Writeln("Block")
	writer.Dedent()
}

func (pass *WriteKernelsPass) RunModulePass(file *ast.File, p *Package) (modified bool, err error) {
	data := pass.compiler.GetPassResult(ParallelizePassType, nil).(*ParallelizeData)

	for _, loop := range data.loops {
		loop.kernelSource, err = generateOpenCL(loop, pass.compiler)
		fmt.Println(loop.kernelSource)
		if err != nil {
			return
		}
		// if rtlib.HasGPU() {} else {}
		hasGPU := &ast.IfStmt{
			Cond: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "rtlib"},
					Sel: &ast.Ident{Name: "HasGPU"},
				},
			},
		}

		loop.parallel.List = append(loop.parallel.List, hasGPU)
		hasGPU.Body = &ast.BlockStmt{} // Launch OpenCL

		cpuParBlock := &ast.BlockStmt{}
		hasGPU.Else = cpuParBlock // Launch multiple goroutines
		/*
			var a, b []int
			rtlib.CPUParallel(func (_idx int) {
				<loop.variables>
				a[idx] += b[idx]
			}, start, stop)
		*/
		funcBody := &ast.BlockStmt{}
		funcBody.List = append(funcBody.List, loop.variables...)
		funcBody.List = append(funcBody.List, loop.kernel)

		cpuParBlock.List = append(cpuParBlock.List, &ast.ExprStmt{X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "rtlib"},
				Sel: &ast.Ident{Name: "CPUParallel"},
			},
			Args: []ast.Expr{
				&ast.FuncLit{
					Type: &ast.FuncType{
						Params: &ast.FieldList{List: []*ast.Field{
							&ast.Field{
								Names: []*ast.Ident{&ast.Ident{Name: "_idx"}},
								Type:  &ast.Ident{Name: "int"},
							},
						}},
					},
					Body: funcBody,
				},
				loop.start,
				loop.stop,
			},
		}})
	}
	return
}
