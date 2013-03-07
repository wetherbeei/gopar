// Types
//
// Supports operations on types, such as figuring out the type of a struct
// field or array access. Also figures out the result of a binary expression
// between two types, or a dereference (*) or address-of (&) operation.

package main

import (
	"bytes"
	"fmt"
	"go/ast"
)

var builtinIdents map[string]bool = map[string]bool{
	"bool":      true,
	"byte":      true,
	"complex64": true,
}

type TypeCategory uint

const (
	UnknownType TypeCategory = iota
	BasicType
	StringType
	ChanType
	StructType
	ArrayType
	SliceType
	MapType
	InterfaceType
	PointerType
)

type Type struct {
	ast.Node
}

func NewType(expr ast.Node) Type {
	return Type{expr}
}

func (t Type) Category() TypeCategory {
	return UnknownType
}

func (t Type) String() string {
	var buffer bytes.Buffer
	var Format func(cur ast.Node) string
	Format = func(cur ast.Node) string {
		switch t := cur.(type) {
		case *ast.Ident:
			buffer.WriteString(t.Name)
		case *ast.BasicLit:
			buffer.WriteString(t.Value)
		case *ast.ArrayType:
			buffer.WriteString("[")
			if t.Len != nil {
				Format(t.Len)
			}
			buffer.WriteString("]")
			Format(t.Elt)
		case *ast.StarExpr:
			buffer.WriteString("*")
			Format(t.X)
		default:
			buffer.WriteString(fmt.Sprintf("<%T %+v>", cur, cur))
		}
		return ""
	}
	Format(t.Node)
	return buffer.String()
}

// Takes an identifier, returns the node that defines it. This should search all
// scopes up to the package level.
type Resolver func(ident string) Type

func TypeOf(expr ast.Node, resolver Resolver) Type {
	var typ ast.Node = expr

	switch t := expr.(type) {
	case *ast.CallExpr:
		var fnType Type
		switch f := t.Fun.(type) {
		case *ast.Ident:
			if f.Name == "make" {
				fnType = NewType(&ast.FuncType{Results: &ast.FieldList{List: []*ast.Field{&ast.Field{Type: t.Args[0]}}}})
			} else if f.Name == "len" {
				fnType = NewType(&ast.FuncType{Results: &ast.FieldList{List: []*ast.Field{&ast.Field{Type: &ast.Ident{Name: "int"}}}}})
			} else {
				fnType = resolver(f.Name)
			}
		case *ast.FuncLit:
			fnType = NewType(f.Type)
		}
		return fnType
	}
	return NewType(typ)
}

// Helper functions for constructing C/OpenCL structures:
// http://golang.org/ref/spec#Size_and_alignment_guarantees
// http://www.khronos.org/registry/cl/sdk/1.1/docs/man/xhtml/attributes-types.html
// https://code.google.com/p/go/source/browse/src/pkg/go/types/sizes.go
func SizeOf(typ Type) int64 {
	return 1
}

func AlignOf(typ Type) int64 {
	return 1
}
