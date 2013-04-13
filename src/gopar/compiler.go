package main

// Compiler framework, LLVM-inspired
//
// Built upon a number of Passes with dependencies to analyze and modify the
// AST. Passes are run until all passes have completed without modifying the
// AST.
//
// Only one Compiler per Project is created, and it can be run multiple times
// with different Packages.
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
	GetResult(interface{}) interface{}
	SetResult(interface{}, interface{})
	Reset()
}

type BasePass struct {
	analysis map[interface{}]interface{}
	compiler *Compiler
}

func NewBasePass() BasePass {
	return BasePass{analysis: make(map[interface{}]interface{})}
}

func (pass *BasePass) GetBase() *BasePass {
	return pass
}

func (pass *BasePass) GetCompiler() *Compiler {
	return pass.compiler
}

func (pass *BasePass) GetResult(v interface{}) (analysis interface{}) {
	return pass.analysis[v]
}

func (pass *BasePass) SetResult(v interface{}, i interface{}) {
	pass.analysis[v] = i
}

func (pass *BasePass) Reset() {
	pass.analysis = make(map[interface{}]interface{})
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
	passStatus       map[*Package]map[PassType]bool
	passes           map[PassType]Pass
	passDependencies map[PassType][]PassType
}

func NewCompiler(project *Project) *Compiler {
	return &Compiler{
		project:          project,
		passStatus:       make(map[*Package]map[PassType]bool),
		passes:           make(map[PassType]Pass),
		passDependencies: make(map[PassType][]PassType),
	}
}

func (c *Compiler) AddPass(pass Pass) {
	bp := pass.GetBase()
	bp.compiler = c
	var t PassType = pass.GetPassType()
	c.passes[t] = pass
	c.passDependencies[t] = pass.GetDependencies()
}

func (c *Compiler) GetPassResult(t PassType, v interface{}) interface{} {
	return c.passes[t].GetResult(v)
}

func (c *Compiler) ResetPass(t PassType) {
	c.passes[t].Reset()
}

// Run all passes while dependencies are met
func (c *Compiler) Run(pkg *Package) (err error) {
	fmt.Printf("\x1b[33;1m%s\x1b[0m\n", strings.Repeat("=", 80))
	fmt.Printf("\x1b[32;1mRunning package %s\x1b[0m\n", pkg.name)
	passStatus := c.passStatus[pkg]
	if passStatus == nil {
		passStatus = make(map[PassType]bool)
		// set all of the existing pass results to false
		for pt := range c.passes {
			passStatus[pt] = false
		}
		c.passStatus[pkg] = passStatus
	}
	for {
		for t, passDeps := range c.passDependencies {
			if passStatus[t] {
				continue
			}
			var canRun bool = true
			for _, dep := range passDeps {
				canRun = canRun && passStatus[dep]
			}
			if canRun {
				fmt.Println(strings.Repeat("-", 80))
				fmt.Printf("\x1b[32;1mRunning %T\x1b[0m\n", c.passes[t])
				var modified bool
				modified, err = c.RunPass(c.passes[t], pkg)
				if err != nil {
					return
				}
				if modified {
					for i, _ := range passStatus {
						passStatus[i] = false
						c.ResetPass(i)
					}
				} else {
					passStatus[t] = true
				}
			}
		}
		var allDone bool = true
		for t, done := range passStatus {
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
		if _, bad := node.(*ast.FuncLit); bad {
			return nil // don't go down lits
		}
		b.passVisitor = b.passVisitor.Visit(node)
		return b
	}
	return nil
}

func RunBasicBlock(pass Pass, root *BasicBlock, p *Package) (modified bool, err error) {
	pos := p.Location(root.node.Pos())
	root.Printf("\x1b[32;1mBasicBlockPass %s:%d\x1b[0m %T %+v", pos.Filename, pos.Line, root.node, root.node)
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

func (c *Compiler) RunPass(pass Pass, pkg *Package) (modified bool, err error) {
	switch pass.GetPassMode() {
	case ModulePassMode:
		fmt.Printf("\x1b[32;1mModulePass %s\x1b[0m\n", pkg.name)
		return pass.RunModulePass(pkg.file, pkg)
	case FunctionPassMode:
		for _, decl := range pkg.file.Decls {
			if fnDecl, ok := decl.(*ast.FuncDecl); ok {
				pos := pkg.Location(fnDecl.Pos())
				fmt.Printf("\x1b[32;1mFunctionPass %s:%d\x1b[0m %s\n", pos.Filename, pos.Line, fnDecl.Name)
				var passMod bool
				passMod, err = pass.RunFunctionPass(fnDecl, pkg)
				modified = modified || passMod
				if err != nil {
					return
				}
			}
		}
	case BasicBlockPassMode:
		for _, decl := range pkg.file.Decls {
			if fnDecl, ok := decl.(*ast.FuncDecl); ok {
				var passMod bool
				b := c.GetPassResult(BasicBlockPassType, fnDecl).(*BasicBlock)
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
