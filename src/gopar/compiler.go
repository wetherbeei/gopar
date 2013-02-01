package main

// Compiler framework, LLVM-inspired
//
// Built upon a number of Passes with dependencies to analyze and modify the
// AST. Passes are run until all passes have completed without modifying the
// AST.
//
import (
	"fmt"
	"go/ast"
)

type PassMode uint

const (
	ModulePass PassMode = iota
	FunctionPass
)

// All of the available passes
type PassType uint

const (
	ExternalFunctionPassType PassType = iota
	DependencyPassType
)

type Pass interface {
	GetPassMode() PassMode

	GetPassType() PassType

	GetDependencies() []PassType

	// Run is executed by the Compiler, and should return true if the AST node was
	// modified.
	RunModulePass(ast.Node, *Compiler) (bool, error)
	RunFunctionPass(ast.Node, *Compiler) (bool, error)

	// Defined by BasePass
	GetResult(ast.Node) interface{}
	SetResult(ast.Node, interface{})
	Reset()
}

type BasePass struct {
	analysis map[ast.Node]interface{}
}

func NewBasePass() BasePass {
	return BasePass{analysis: make(map[ast.Node]interface{})}
}

func (pass *BasePass) GetResult(node ast.Node) (analysis interface{}) {
	return pass.analysis[node]
}

func (pass *BasePass) SetResult(node ast.Node, i interface{}) {
	pass.analysis[node] = i
}

func (pass *BasePass) Reset() {
	pass.analysis = make(map[ast.Node]interface{})
}

func (pass *BasePass) RunModulePass(node ast.Node, c *Compiler) (bool, error) {
	return false, fmt.Errorf("Undefined pass")
}

func (pass *BasePass) RunFunctionPass(node ast.Node, c *Compiler) (bool, error) {
	return false, fmt.Errorf("Undefined pass")
}

func (pass *BasePass) RunLoopPass(node ast.Node, c *Compiler) (bool, error) {
	return false, fmt.Errorf("Undefined pass")
}

type Compiler struct {
	project          *Project
	passStatus       map[PassType]bool
	passes           map[PassType]Pass
	passDependencies map[PassType][]PassType
}

func NewCompiler(project *Project) *Compiler {
	return &Compiler{
		project:          project,
		passStatus:       make(map[PassType]bool),
		passes:           make(map[PassType]Pass),
		passDependencies: make(map[PassType][]PassType),
	}
}

func (c *Compiler) AddPass(pass Pass) {
	var t PassType = pass.GetPassType()
	c.passes[t] = pass
	c.passStatus[t] = false
	c.passDependencies[t] = pass.GetDependencies()
}

func (c *Compiler) GetPassResult(t PassType, node ast.Node) interface{} {
	return c.passes[t].GetResult(node)
}

func (c *Compiler) ResetPass(t PassType) {
	c.passes[t].Reset()
}

// Run all passes while dependencies are met
func (c *Compiler) Run() (err error) {
	for {
		for t, passDeps := range c.passDependencies {
			if c.passStatus[t] {
				continue
			}
			var canRun bool = true
			for _, dep := range passDeps {
				canRun = canRun && c.passStatus[dep]
			}
			if canRun {
				fmt.Printf("Running %T\n", c.passes[t])
				var modified bool
				modified, err = c.RunPass(c.passes[t])
				if err != nil {
					return
				}
				if modified {
					for i, _ := range c.passStatus {
						c.passStatus[i] = false
						c.ResetPass(i)
					}
				} else {
					c.passStatus[t] = true
				}
			} else {
				fmt.Printf("Can't run %T yet\n", c.passes[t])
			}
		}
		var allDone bool = true
		for t, done := range c.passStatus {
			fmt.Printf("Status %T = %t\n", c.passes[t], done)
			allDone = allDone && done
		}
		if allDone {
			break // all passes completed successfully, exit
		}
	}
	return
}

func (c *Compiler) RunPass(pass Pass) (modified bool, err error) {
	// TODO: only does main package so far
	switch pass.GetPassMode() {
	case ModulePass:
		return pass.RunModulePass(c.project.get("main").file, c)
	case FunctionPass:
		pkg := c.project.get("main")
		for _, obj := range pkg.TopLevel() {
			if obj.Kind == ast.Fun {
				var passMod bool
				passMod, err = pass.RunFunctionPass(obj.Decl.(*ast.FuncDecl), c)
				modified = modified || passMod
				if err != nil {
					return
				}
			}
		}
	}
	return
}
