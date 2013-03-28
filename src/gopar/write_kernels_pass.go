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
	return w.Write(str + "\n")
}

// Search for each \n and add "indent" spaces before it
func (w *CWriter) Write(str string) (n int, err error) {
	for {
		index := strings.Index(str, "\n")
		if index == -1 {
			break
		}
		_, err = io.WriteString(w.writer, str[:index])
		if err != nil {
			return
		}
		_, err = io.WriteString(w.writer, strings.Repeat(" ", w.indent))
		if err != nil {
			return
		}
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

func generateOpenCL(data *ParallelLoopInfo, c *Compiler) (kernel string, err error) {
	// parse the loop body
	// - define main kernel function with all args used
	// - types/data structures
	// - other functions
	writer := NewCWriter()
	writer.Writeln(fmt.Sprintf("/* Generated OpenCL for %s */", data.name))
	return writer.String(), nil
}

func (pass *WriteKernelsPass) RunModulePass(file *ast.File, p *Package) (modified bool, err error) {
	data := pass.compiler.GetPassResult(ParallelizePassType, nil).(*ParallelizeData)
	//kernel, err := generateOpenCL(data, p)

	for _, loop := range data.loops {
		loop.kernelSource, err = generateOpenCL(loop, pass.compiler)
		fmt.Println(loop.kernelSource)
		if err != nil {
			return
		}
	}
	return
}
