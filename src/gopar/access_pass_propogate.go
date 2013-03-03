// Access pass propogation
//
// Pass accesses up through the blocks to the function declaration
package main

import (
	"go/ast"
)

type AccessPassPropogate struct {
	BasePass
}

type AccessPassPropogateVisitor struct {
	p         *Package
	cur       *BasicBlock
	dataBlock *AccessPassData
	node      ast.Node
}

func (v AccessPassPropogateVisitor) Done(block *BasicBlock) (modified bool, err error) {
	dataBlock := v.dataBlock

	block.Print("== Defines ==")
	for ident, expr := range dataBlock.defines {
		block.Printf("%s = %T %+v", ident, expr, expr)
	}
	block.Print("== Accesses ==")
	for _, access := range dataBlock.accesses {
		block.Printf(access.String())
	}

	MergeDependenciesUpwards(block)
	return
}

func (v AccessPassPropogateVisitor) Visit(node ast.Node) (w BasicBlockVisitor) {
	return v
}

func NewAccessPassPropogate() *AccessPassPropogate {
	return &AccessPassPropogate{
		BasePass: NewBasePass(),
	}
}

func (pass *AccessPassPropogate) GetPassType() PassType {
	return AccessPassPropogateType
}

func (pass *AccessPassPropogate) GetPassMode() PassMode {
	return BasicBlockPassMode
}

func (pass *AccessPassPropogate) GetDependencies() []PassType {
	return []PassType{AccessPassType}
}

func MergeDependenciesUpwards(child *BasicBlock) {
	// TODO: merge reads/writes of identifiers outside this scope
	if child.parent == nil {
		return
	}
	parent := child.parent
	dataBlock := child.Get(AccessPassType).(*AccessPassData)
	parentDataBlock := parent.Get(AccessPassType).(*AccessPassData)
	for _, access := range dataBlock.accesses {
		// move to parent if not defined in this block
		// also don't merge *a or &a accesses
		if _, ok := dataBlock.defines[access.group[0].id]; !ok {
			// if there is an array access that uses an identifier block defined in 
			// this block, change the access from b[idx] to b
			var ig IdentifierGroup = access
			for idx, ident := range access.group {
				if _, ok := dataBlock.defines[ident.index]; ok && ident.isIndexed {
					ig.group = make([]Identifier, idx+1)
					copy(ig.group, access.group)
					parent.Printf("Leaving index scope [%s]", ig.group[idx].index)
					ig.group[idx].isIndexed = false
					ig.group[idx].index = ""
				}
				break
			}
			parentDataBlock.accesses = append(parentDataBlock.accesses, ig)
			parent.Print("<< Merged upwards", ig.String())
		}
	}
	return
}

func (pass *AccessPassPropogate) RunBasicBlockPass(block *BasicBlock, p *Package) BasicBlockVisitor {
	dataBlock := block.Get(AccessPassType).(*AccessPassData)

	return AccessPassPropogateVisitor{cur: block, dataBlock: dataBlock, p: p}
}
