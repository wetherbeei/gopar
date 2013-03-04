// Dependency pass analysis
//
// Use the individual accesses in the Access Pass to classify each variable as
// WriteFirst, ReadOnly or ReadWrite

package main

import (
	"bytes"
)

type DependencyPass struct {
	BasePass
}

type DependencyType uint

const (
	Unknown DependencyType = iota
	ReadOnly
	WriteFirst
	ReadWrite
)

var dependencyTypeString = map[DependencyType]string{
	Unknown:    "\x1b[31mUnknown\x1b[0m",
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

// Store all dependencies for this block
type Dependency struct {
	group   []Identifier
	depType DependencyType
}

func (d *Dependency) String() string {
	var buffer bytes.Buffer
	for _, i := range d.group {
		buffer.WriteString(i.id)
		if i.isIndexed {
			buffer.WriteString("[")
			buffer.WriteString(i.index)
			buffer.WriteString("]")
		}
		buffer.WriteString(".")
	}
	buffer.WriteString(dependencyTypeString[d.depType])
	return buffer.String()
}

func ClassifyAccess(prev DependencyType, t AccessType) DependencyType {
	if prev != Unknown {
		// upgrade the previous access
		if prev == ReadOnly && t == WriteAccess {
			return ReadWrite
		}
		return prev
	} else {
		if t == ReadAccess {
			return ReadOnly
		} else if t == WriteAccess {
			return WriteFirst
		}
		return Unknown
	}
	return Unknown
}

type DependencyPassData struct {
	deps []Dependency
}

func NewDependencyPassData() *DependencyPassData {
	return &DependencyPassData{
		deps: make([]Dependency, 0),
	}
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
	return []PassType{AccessPassFuncPropogateType}
}

func (pass *DependencyPass) RunBasicBlockPass(block *BasicBlock, p *Package) BasicBlockVisitor {
	dataBlock := block.Get(AccessPassType).(*AccessPassData)
	dependencyData := NewDependencyPassData()
	block.Set(DependencyPassType, dependencyData)
	for _, access := range dataBlock.accesses {
		block.Printf(access.String())
	}
	for _, access := range dataBlock.accesses {
		// propogate the access to potentially multiple dependency entries for
		// subaccesses, arrays, etc
		//
		// Check the dep list for sub or super accesses (a.b or a.b.c.d for a.b.c)
		// Super and exact accesses (dep = a.b, access = a.b.c or a.b[idx])
		//
		// a.b = ReadOnly (dep)
		// a.b.c = WriteAccess (access)
		// a.b -> ReadWrite
		for idx, dep := range dependencyData.deps {
			if len(access.group) < len(dep.group) {
				// sub-access
				for i := 0; i < len(access.group); i++ {
					if !access.group[i].Equals(dep.group[i]) {
						break
					}
					if i == len(access.group)-1 {
						// at the end, everything matches up to here
						dependencyData.deps[idx].depType = ClassifyAccess(dep.depType, access.t)
						block.Printf("Sub-access %s < %s", dep.String(), access.String())
						block.Printf("  => %s", dependencyData.deps[idx].String())
					}
				}
			} else {
				// super-access
				for i := 0; i < len(dep.group); i++ {
					if !dep.group[i].Equals(access.group[i]) {
						break
					}
					if i == len(dep.group)-1 {
						// at the end, everything matches up to here
						dependencyData.deps[idx].depType = ClassifyAccess(dep.depType, access.t)
						block.Printf("Super-access %s >= %s", dep.String(), access.String())
						block.Printf("  => %s", dependencyData.deps[idx].String())
					}
				}
			}
		}

		// Add this access if it's unique (a.b[idx])
		var matched bool
		for _, dep := range dependencyData.deps {
			if len(dep.group) != len(access.group) {
				continue // not an exact match
			}
			for i := 0; i < len(dep.group); i++ {
				if !dep.group[i].Equals(access.group[i]) {
					break
				}
				// if exact match
				if i == len(dep.group)-1 {
					matched = true
				}
			}
		}
		if !matched {
			dep := Dependency{
				group:   access.group,
				depType: ClassifyAccess(Unknown, access.t),
			}
			dependencyData.deps = append(dependencyData.deps, dep)
			block.Printf("Added dep %s", dep.String())
		}
	}

	block.Print("== Dependencies ==")
	for _, dep := range dependencyData.deps {
		block.Print(dep.String())
	}

	return DefaultBasicBlockVisitor{}
}
