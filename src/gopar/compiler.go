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
	"strings"
)

type PassMode uint

const (
	ModulePassMode PassMode = iota
	FunctionPassMode
	BasicBlockPassMode
)

// All of the available passes
type PassType uint

const (
	BasicBlockPassType PassType = iota
	InvalidConstructPassType
	CallGraphPassType
	DefinedTypesPassType
	AccessPassType
	AccessPassPropogateType
	AccessPassFuncPropogateType
	DependencyPassType
	ParallelizePassType
	InsertBlocksPassType
	WriteKernelsPassType
)

type Pass interface {
	GetBase() *BasePass
	GetCompiler() *Compiler
	GetPassMode() PassMode

	GetPassType() PassType

	GetDependencies() []PassType

	// Run is executed by the Compiler, and should return true if the AST node was
	// modified.
	RunModulePass(*ast.File, *Package) (bool, error)
	RunFunctionPass(*ast.FuncDecl, *Package) (bool, error)
	RunBasicBlockPass(*BasicBlock, *Package) BasicBlockVisitor

	// Defined by BasePass
	GetResult(ast.Node) interface{}
	SetResult(ast.Node, interface{})
	Reset()
}

type BasePass struct {
	analysis map[ast.Node]interface{}
	compiler *Compiler
}

func NewBasePass() BasePass {
	return BasePass{analysis: make(map[ast.Node]interface{})}
}

func (pass *BasePass) GetBase() *BasePass {
	return pass
}

func (pass *BasePass) GetCompiler() *Compiler {
	return pass.compiler
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

func (pass *BasePass) RunModulePass(file *ast.File, p *Package) (bool, error) {
	return false, fmt.Errorf("Undefined pass")
}

func (pass *BasePass) RunFunctionPass(function *ast.FuncDecl, p *Package) (bool, error) {
	return false, fmt.Errorf("Undefined pass")
}

type BasicBlockVisitor interface {
	Visit(ast.Node) BasicBlockVisitor

	// Called when the visit to this BasicBlock is done. Return true if the AST
	// was modified.
	Done(*BasicBlock) (bool, error)
}

type DefaultBasicBlockVisitor struct {
}

func (d DefaultBasicBlockVisitor) Visit(node ast.Node) BasicBlockVisitor {
	return d
}

func (d DefaultBasicBlockVisitor) Done(b *BasicBlock) (modified bool, err error) {
	return
}

// Run through each basic block in depth-first order. When a new basic block is
// encountered, call RunBasicBlockPass on it, then return to executing the 
// current pass.
func (pass *BasePass) RunBasicBlockPass(b *BasicBlock, p *Package) BasicBlockVisitor {
	return DefaultBasicBlockVisitor{}
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
	bp := pass.GetBase()
	bp.compiler = c
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
				fmt.Println(strings.Repeat("-", 80))
				fmt.Printf("\x1b[32;1mRunning %T\x1b[0m\n", c.passes[t])
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

type BasicBlockVisitorImpl struct {
	pass        Pass
	p           *Package
	block       *BasicBlock
	passVisitor BasicBlockVisitor
	modified    bool
	err         error
}

func (b BasicBlockVisitorImpl) Visit(node ast.Node) ast.Visitor {
	switch {
	case isBasicBlockNode(node) && node != b.block.node:
		basicBlock := b.pass.GetCompiler().GetPassResult(BasicBlockPassType, node).(*BasicBlock)
		var modified bool
		modified, b.err = RunBasicBlock(b.pass, basicBlock, b.p)
		b.modified = b.modified || modified
		return nil // already traversed
	case b.passVisitor != nil:
		if b.err != nil {
			return nil
		}
		b.passVisitor = b.passVisitor.Visit(node)
		return b
	}
	return nil
}

func RunBasicBlock(pass Pass, root *BasicBlock, p *Package) (modified bool, err error) {
	root.Printf("\x1b[32;1mBasicBlockPass\x1b[0m %T %+v", root.node, root.node)
	passVisitor := pass.RunBasicBlockPass(root, p)
	n := BasicBlockVisitorImpl{pass: pass, p: p, block: root, passVisitor: passVisitor}
	ast.Walk(n, root.node)
	if passVisitor == nil {
		return
	}
	mod, e := passVisitor.Done(root)
	n.modified = n.modified || mod
	if n.err == nil {
		n.err = e
	}
	return n.modified, n.err
}

func (c *Compiler) RunPass(pass Pass) (modified bool, err error) {
	// TODO: only does main package so far
	pkg := c.project.get("main")
	switch pass.GetPassMode() {
	case ModulePassMode:
		return pass.RunModulePass(pkg.file, pkg)
	case FunctionPassMode:
		for _, obj := range pkg.TopLevel() {
			if obj.Kind == ast.Fun {
				fmt.Println("\x1b[32;1mFunctionPass\x1b[0m", obj.Decl.(*ast.FuncDecl).Name)
				var passMod bool
				passMod, err = pass.RunFunctionPass(obj.Decl.(*ast.FuncDecl), pkg)
				modified = modified || passMod
				if err != nil {
					return
				}
			}
		}
	case BasicBlockPassMode:
		for _, obj := range pkg.TopLevel() {
			if obj.Kind == ast.Fun {
				var passMod bool
				b := c.GetPassResult(BasicBlockPassType, obj.Decl.(*ast.FuncDecl)).(*BasicBlock)
				passMod, err = RunBasicBlock(pass, b, pkg)
				modified = modified || passMod
				if err != nil {
					return
				}
			}
		}
	}
	return
}
