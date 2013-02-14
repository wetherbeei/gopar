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
