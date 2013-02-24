// Dependency pass analysis
//
// Use the individual accesses in the Access Pass to classify each variable as
// WriteFirst, ReadOnly or ReadWrite

package main

/*
func ClassifyAccess(ident string, t AccessType) {
	if prev, ok := dataBlock.accesses[ident]; ok {
		// upgrade the previous access
		if prev == ReadOnly && t == WriteAccess {
			//dataBlock.accesses[ident] = ReadWrite
		}
	} else {
		if t == ReadAccess {
			//dataBlock.accesses[ident] = ReadOnly
		} else if t == WriteAccess {
			//dataBlock.accesses[ident] = WriteFirst
		}
	}
}
*/
type DependencyPass struct {
	BasePass
}

type DependencyType uint

const (
	ReadOnly DependencyType = iota
	WriteFirst
	ReadWrite
)

var dependencyTypeString = map[DependencyType]string{
	ReadOnly:   "\x1b[32mReadOnly\x1b[0m",
	WriteFirst: "\x1b[33mWriteFirst\x1b[0m",
	ReadWrite:  "\x1b[35mReadWrite\x1b[0m",
}

type DependencyLevel struct {
	id         string
	dependency DependencyType
	isIndexed  bool // is the children map by sub-expression (x.y) or index (x[y])
	children   map[string]*DependencyLevel
}

func NewDependencyLevel() *DependencyLevel {
	return &DependencyLevel{
		children: make(map[string]*DependencyLevel),
	}
}

type DependencyPassData struct {
}

func NewDependencyPass() *DependencyPass {
	return &DependencyPass{
		BasePass: NewBasePass(),
	}
}

func (pass *DependencyPass) GetPassType() PassType {
	return DependencyPassType
}

func (pass *DependencyPass) GetPassMode() PassMode {
	return BasicBlockPassMode
}

func (pass *DependencyPass) GetDependencies() []PassType {
	return []PassType{AccessPassPropogateType}
}

func (pass *DependencyPass) RunBasicBlockPass(block *BasicBlock, p *Package) BasicBlockVisitor {
	dataBlock := block.Get(AccessPassType).(*AccessPassData)
	for _, access := range dataBlock.accesses {
		block.Print(access.String())
	}
	return DefaultBasicBlockVisitor{}
}
